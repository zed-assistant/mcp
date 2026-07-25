package config

import (
	"bytes"
	"fmt"
	"strings"
)

// sandboxParser holds the raw source bytes of a SandboxVars lua file and a cursor
// into it, used only for the two things golua's AST can't give us directly: the
// exact byte extent of a value it doesn't classify as a leaf we support (so a
// splice-affecting edit never touches it), and the leading comment text that
// precedes a key (golua's scanner discards comments as trivia, like every
// standard Lua tokenizer). It never copies/rewrites the source; it only ever
// records or walks byte ranges so that writes can splice the exact bytes in
// place.
type sandboxParser struct {
	src []byte
	pos int
}

func (p *sandboxParser) peek() byte {
	if p.pos >= len(p.src) {
		return 0
	}
	return p.src[p.pos]
}

func (p *sandboxParser) hasPrefix(s string) bool {
	return p.pos+len(s) <= len(p.src) && string(p.src[p.pos:p.pos+len(s)]) == s
}

func (p *sandboxParser) skipWhitespace() {
	for p.pos < len(p.src) {
		switch p.src[p.pos] {
		case ' ', '\t', '\r', '\n':
			p.pos++
		default:
			return
		}
	}
}

// tryLongBracketOpen recognizes a Lua long-bracket opener: '[' '='* '['.
// On success it consumes the opener and returns the '=' level. On failure it
// leaves the cursor untouched.
func (p *sandboxParser) tryLongBracketOpen() (level int, ok bool) {
	save := p.pos
	if p.pos >= len(p.src) || p.src[p.pos] != '[' {
		return 0, false
	}
	i := p.pos + 1
	eq := 0
	for i < len(p.src) && p.src[i] == '=' {
		eq++
		i++
	}
	if i < len(p.src) && p.src[i] == '[' {
		p.pos = i + 1
		return eq, true
	}
	p.pos = save
	return 0, false
}

// consumeLongBracketBody consumes up to and including the matching closing
// long bracket (']' '='*level ']'), assuming the opener has already been consumed.
func (p *sandboxParser) consumeLongBracketBody(level int) (string, error) {
	closer := "]" + strings.Repeat("=", level) + "]"
	idx := bytes.Index(p.src[p.pos:], []byte(closer))
	if idx < 0 {
		return "", fmt.Errorf("unterminated long bracket (level %d) starting near byte %d", level, p.pos)
	}
	content := string(p.src[p.pos : p.pos+idx])
	p.pos += idx + len(closer)
	return content, nil
}

// readStringLiteral reads a single-quoted or double-quoted Lua string literal
// starting at the current position. Escape sequences are decoded only well
// enough to correctly find the literal's end (their decoded values are
// discarded); golua's ast.String already provides the authoritative decoded
// value for literals we expose as leaves. It fully commits (mutates p.pos) on
// both success and hard failure; callers that want to try speculatively should
// save/restore p.pos themselves.
func (p *sandboxParser) readStringLiteral() (end int, ok bool, err error) {
	if p.pos >= len(p.src) {
		return 0, false, nil
	}
	c := p.src[p.pos]
	if c != '"' && c != '\'' {
		return 0, false, nil
	}
	quote := c
	start := p.pos
	p.pos++
	for {
		if p.pos >= len(p.src) {
			return 0, false, fmt.Errorf("unterminated string literal starting at byte %d", start)
		}
		ch := p.src[p.pos]
		if ch == quote {
			p.pos++
			return p.pos, true, nil
		}
		if ch == '\n' {
			return 0, false, fmt.Errorf("unterminated string literal starting at byte %d (newline before closing quote)", start)
		}
		if ch != '\\' {
			p.pos++
			continue
		}
		p.pos++
		if p.pos >= len(p.src) {
			return 0, false, fmt.Errorf("unterminated escape sequence at byte %d", p.pos)
		}
		p.pos++
	}
}

// tryConsumeComment consumes a Lua comment ("--" line comment, or "--[[ ]]" style
// long comment) starting at the current position. Returns ok=false without moving
// the cursor if there is no comment here.
func (p *sandboxParser) tryConsumeComment() (text string, ok bool, err error) {
	if !p.hasPrefix("--") {
		return "", false, nil
	}
	p.pos += 2
	if level, isLong := p.tryLongBracketOpen(); isLong {
		content, err := p.consumeLongBracketBody(level)
		if err != nil {
			return "", false, err
		}
		return content, true, nil
	}
	start := p.pos
	for p.pos < len(p.src) && p.src[p.pos] != '\n' {
		p.pos++
	}
	return string(p.src[start:p.pos]), true, nil
}

func stripLuaCommentPrefix(s string) string {
	return strings.TrimRight(strings.TrimLeft(s, " \t"), " \t\r")
}

// skipInsignificant skips whitespace and any run of contiguous comments,
// returning their concatenated, prefix-stripped text as a candidate description
// for whatever key follows.
func (p *sandboxParser) skipInsignificant() (string, error) {
	var lines []string
	for {
		p.skipWhitespace()
		if !p.hasPrefix("--") {
			break
		}
		text, ok, err := p.tryConsumeComment()
		if err != nil {
			return "", err
		}
		if !ok {
			break
		}
		lines = append(lines, stripLuaCommentPrefix(text))
	}
	return strings.Join(lines, "\n"), nil
}

// consumeOptionalComma skips an optional trailing comma after a table-entry
// value. A comma may be preceded by whitespace/comments; those comments (if
// any) are discarded rather than misattributed, since real SandboxVars files
// never place a comment between a value and its trailing comma.
func (p *sandboxParser) consumeOptionalComma() {
	save := p.pos
	if _, err := p.skipInsignificant(); err != nil {
		p.pos = save
		return
	}
	if p.peek() == ',' {
		p.pos++
	}
}

// scanValueSpan scans a table-entry value expression starting at the current
// position (after skipping leading whitespace) up to - but not including - the
// first top-level ',' or the enclosing table's closing '}'. It returns the
// [start,end) byte range of the expression with trailing whitespace trimmed.
// It is used only for values golua's AST didn't classify as a leaf kind we
// support (e.g. a function call or arithmetic expression): we still need to
// know their extent so comment/description scanning for the next key resumes
// in the right place, without corrupting a byte splice for anything nearby.
func (p *sandboxParser) scanValueSpan() (start, end int, err error) {
	p.skipWhitespace()
	start = p.pos
	depth := 0
	for {
		if p.pos >= len(p.src) {
			return 0, 0, fmt.Errorf("unexpected end of file while scanning value starting at byte %d", start)
		}
		c := p.src[p.pos]
		switch {
		case c == '"' || c == '\'':
			_, ok, serr := p.readStringLiteral()
			if serr != nil {
				return 0, 0, serr
			}
			if !ok {
				return 0, 0, fmt.Errorf("internal error reading string literal at byte %d", p.pos)
			}
		case c == '-' && p.pos+1 < len(p.src) && p.src[p.pos+1] == '-':
			p.pos += 2
			if level, isLong := p.tryLongBracketOpen(); isLong {
				if _, cerr := p.consumeLongBracketBody(level); cerr != nil {
					return 0, 0, cerr
				}
			} else {
				for p.pos < len(p.src) && p.src[p.pos] != '\n' {
					p.pos++
				}
			}
		case c == '[':
			if level, isLong := p.tryLongBracketOpen(); isLong {
				if _, cerr := p.consumeLongBracketBody(level); cerr != nil {
					return 0, 0, cerr
				}
				continue
			}
			depth++
			p.pos++
		case c == '(' || c == '{':
			depth++
			p.pos++
		case c == ')' || c == ']':
			depth--
			if depth < 0 {
				return 0, 0, fmt.Errorf("unbalanced closing bracket at byte %d", p.pos)
			}
			p.pos++
		case c == '}':
			if depth == 0 {
				end = trimSpanEnd(p.src, start, p.pos)
				return start, end, nil
			}
			depth--
			if depth < 0 {
				return 0, 0, fmt.Errorf("unbalanced '}' at byte %d", p.pos)
			}
			p.pos++
		case c == ',':
			if depth == 0 {
				end = trimSpanEnd(p.src, start, p.pos)
				return start, end, nil
			}
			p.pos++
		default:
			p.pos++
		}
	}
}

func trimSpanEnd(src []byte, start, end int) int {
	for end > start {
		switch src[end-1] {
		case ' ', '\t', '\r', '\n':
			end--
			continue
		}
		break
	}
	return end
}
