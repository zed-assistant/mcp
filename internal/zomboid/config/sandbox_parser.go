package config

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	decimalIntegerRegex = regexp.MustCompile(`^-?\d+$`)
	decimalNumberRegex  = regexp.MustCompile(`^-?(\d+\.\d*|\.\d+|\d+)([eE][-+]?\d+)?$`)
	hexNumberRegex      = regexp.MustCompile(`^-?0[xX][0-9a-fA-F]+$`)
)

// isSandboxNumberLiteral reports whether s is any recognized Lua numeric
// literal, integer or float.
func isSandboxNumberLiteral(s string) bool {
	return decimalNumberRegex.MatchString(s) || hexNumberRegex.MatchString(s)
}

// isSandboxIntegerLiteral reports whether s is specifically a whole-number
// literal (no decimal point, no exponent). Hex literals are always integers.
func isSandboxIntegerLiteral(s string) bool {
	return decimalIntegerRegex.MatchString(s) || hexNumberRegex.MatchString(s)
}

// classifySandboxNumberLiteral distinguishes integer from float purely by the
// literal's written form: hex is always an integer; any decimal literal
// containing '.' or an exponent is a float; otherwise it's an integer. Hex is
// checked first because a hex literal like "0xE5" contains an 'E' that would
// otherwise be mistaken for a decimal exponent.
func classifySandboxNumberLiteral(s string) sandboxLeafKind {
	if hexNumberRegex.MatchString(s) {
		return sandboxLeafInteger
	}
	if strings.ContainsAny(s, ".eE") {
		return sandboxLeafFloat
	}
	return sandboxLeafInteger
}

// parseSandboxFile parses a SandboxVars.lua file into a tree rooted at the
// SandboxVars table. It fails closed (returns an error) on any genuine syntax
// problem - unbalanced brackets, unterminated strings, unsupported escapes -
// since those mean we can no longer be sure where entries begin and end. It is
// lenient about *values* it can't classify as a scalar (arrays, function calls):
// those are parsed enough to find their extent and then silently omitted from
// the tree, leaving their bytes untouched but not exposed for read/update.
func parseSandboxFile(src []byte) (*sandboxNode, error) {
	p := &sandboxParser{src: src}

	if _, err := p.skipInsignificant(); err != nil {
		return nil, err
	}
	name, ok := p.readIdentifier()
	if !ok || name != "SandboxVars" {
		return nil, fmt.Errorf("expected 'SandboxVars' identifier at start of file (byte %d)", p.pos)
	}
	if _, err := p.skipInsignificant(); err != nil {
		return nil, err
	}
	if p.peek() != '=' {
		return nil, fmt.Errorf("expected '=' after 'SandboxVars' (byte %d)", p.pos)
	}
	p.pos++
	if _, err := p.skipInsignificant(); err != nil {
		return nil, err
	}
	if p.peek() != '{' {
		return nil, fmt.Errorf("expected '{' to start the SandboxVars table (byte %d)", p.pos)
	}

	return p.parseTable("")
}

// parseTable parses a Lua table literal starting at the current '{' position,
// consuming through its matching '}'.
func (p *sandboxParser) parseTable(ownDescription string) (*sandboxNode, error) {
	if p.peek() != '{' {
		return nil, fmt.Errorf("expected '{' at byte %d", p.pos)
	}
	p.pos++

	node := newSandboxGroupNode()
	node.Description = ownDescription

	for {
		pendingDescription, err := p.skipInsignificant()
		if err != nil {
			return nil, err
		}
		if p.atEnd() {
			return nil, fmt.Errorf("unexpected end of file inside table (byte %d)", p.pos)
		}
		if p.peek() == '}' {
			p.pos++
			return node, nil
		}

		if p.peek() == '[' {
			keyName, keyOk, err := p.tryReadBracketedKey()
			if err != nil {
				return nil, err
			}
			if !keyOk {
				// A bracketed key we can't read as a plain string (numeric/array-style
				// index, or computed key). Skip the whole entry generically: its bytes
				// are preserved, it's just not exposed as an editable entry.
				if err := p.skipUnsupportedArrayEntry(); err != nil {
					return nil, err
				}
				continue
			}
			childNode, err := p.parseAssignedValue(pendingDescription)
			if err != nil {
				return nil, err
			}
			p.consumeOptionalComma()
			if childNode != nil {
				node.addChild(keyName, childNode)
			}
			continue
		}

		keyName, ok := p.readIdentifier()
		if !ok {
			return nil, fmt.Errorf("expected a key name at byte %d", p.pos)
		}
		childNode, err := p.parseAssignedValue(pendingDescription)
		if err != nil {
			return nil, err
		}
		p.consumeOptionalComma()
		if childNode != nil {
			node.addChild(keyName, childNode)
		}
	}
}

// parseAssignedValue expects '=' followed by a value at the current position
// (whitespace/comments tolerated in between) and parses that value.
func (p *sandboxParser) parseAssignedValue(pendingDescription string) (*sandboxNode, error) {
	if _, err := p.skipInsignificant(); err != nil {
		return nil, err
	}
	if p.peek() != '=' {
		return nil, fmt.Errorf("expected '=' at byte %d", p.pos)
	}
	p.pos++
	return p.parseValue(pendingDescription)
}

func (p *sandboxParser) consumeOptionalComma() {
	// A comma may be preceded by whitespace/comments; those comments (if any)
	// are discarded rather than misattributed, since real SandboxVars files
	// never place a comment between a value and its trailing comma.
	save := p.pos
	if _, err := p.skipInsignificant(); err != nil {
		p.pos = save
		return
	}
	if p.peek() == ',' {
		p.pos++
	}
}

// parseValue parses a single table-entry value: either a nested table, or a
// scalar we can classify as number/bool/string. Anything else is skipped and
// reported back as (nil, nil) so the caller omits it from the tree without
// failing the parse.
func (p *sandboxParser) parseValue(pendingDescription string) (*sandboxNode, error) {
	p.skipWhitespace()
	if p.peek() == '{' {
		return p.parseTable(pendingDescription)
	}

	start, end, err := p.scanValueSpan()
	if err != nil {
		return nil, err
	}
	if end <= start {
		return nil, fmt.Errorf("empty value at byte %d", start)
	}
	text := string(p.src[start:end])

	if isSandboxNumberLiteral(text) {
		return &sandboxNode{Leaf: &sandboxLeaf{
			Kind:        classifySandboxNumberLiteral(text),
			Value:       text,
			Description: pendingDescription,
			ValueStart:  start,
			ValueEnd:    end,
		}}, nil
	}
	if text == "true" || text == "false" {
		return &sandboxNode{Leaf: &sandboxLeaf{
			Kind:        sandboxLeafBool,
			Value:       text,
			Description: pendingDescription,
			ValueStart:  start,
			ValueEnd:    end,
		}}, nil
	}
	if len(text) >= 2 && (text[0] == '"' || text[0] == '\'') {
		// Re-derive the decoded value by re-reading the string literal from its
		// start; this is the same, already-validated span scanValueSpan just
		// walked over (scanValueSpan itself calls readStringLiteral to skip
		// strings), so this can't newly fail. We only accept it as a *leaf* if
		// the literal spans exactly [start,end) - i.e. the whole value is one
		// clean string, not a concatenation or other expression starting with a
		// quote.
		strParser := &sandboxParser{src: p.src, pos: start}
		strVal, quote, strEnd, ok, strErr := strParser.readStringLiteral()
		if strErr == nil && ok && strEnd == end {
			return &sandboxNode{Leaf: &sandboxLeaf{
				Kind:        sandboxLeafString,
				Value:       strVal,
				QuoteChar:   quote,
				Description: pendingDescription,
				ValueStart:  start,
				ValueEnd:    end,
			}}, nil
		}
	}

	// Unrecognized value shape (array table already handled via '{' above isn't
	// reachable here; this covers function calls, concatenation, etc). Bytes are
	// preserved untouched; it's just not exposed as an editable entry.
	return nil, nil
}

// tryReadBracketedKey reads a `["name"]` / `['name']` style key. ok=false (with
// the cursor left where it started scanning the key contents) means the
// bracketed key isn't a plain string literal (e.g. a numeric or computed key).
func (p *sandboxParser) tryReadBracketedKey() (string, bool, error) {
	save := p.pos
	if p.peek() != '[' {
		return "", false, nil
	}
	p.pos++
	p.skipWhitespace()
	if p.peek() != '"' && p.peek() != '\'' {
		p.pos = save
		return "", false, nil
	}
	strVal, _, _, ok, err := p.readStringLiteral()
	if err != nil {
		return "", false, err
	}
	if !ok {
		p.pos = save
		return "", false, nil
	}
	p.skipWhitespace()
	if p.peek() != ']' {
		return "", false, fmt.Errorf("expected ']' to close bracketed key at byte %d", p.pos)
	}
	p.pos++
	return strVal, true, nil
}

// skipUnsupportedArrayEntry skips a whole `[expr] = value` table entry whose
// key isn't a plain string literal, positioning the cursor just past its
// optional trailing comma.
func (p *sandboxParser) skipUnsupportedArrayEntry() error {
	if err := p.scanBalanced(); err != nil {
		return err
	}
	if _, err := p.skipInsignificant(); err != nil {
		return err
	}
	if p.peek() != '=' {
		return fmt.Errorf("expected '=' after bracketed key at byte %d", p.pos)
	}
	p.pos++
	if _, _, err := p.scanValueSpan(); err != nil {
		return err
	}
	p.consumeOptionalComma()
	return nil
}
