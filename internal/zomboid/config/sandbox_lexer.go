package config

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
)

// sandboxParser holds the raw source bytes of a SandboxVars lua file and a cursor
// into it. It never copies/rewrites the source; every parsed leaf just records the
// byte range of its value so that writes can splice the exact bytes in place.
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

func (p *sandboxParser) atEnd() bool {
	return p.pos >= len(p.src)
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

func isLuaIdentStart(b byte) bool {
	return b == '_' || (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z')
}

func isLuaIdentPart(b byte) bool {
	return isLuaIdentStart(b) || (b >= '0' && b <= '9')
}

func (p *sandboxParser) readIdentifier() (string, bool) {
	if p.pos >= len(p.src) || !isLuaIdentStart(p.src[p.pos]) {
		return "", false
	}
	start := p.pos
	p.pos++
	for p.pos < len(p.src) && isLuaIdentPart(p.src[p.pos]) {
		p.pos++
	}
	return string(p.src[start:p.pos]), true
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
// starting at the current position, decoding escape sequences. It fully commits
// (mutates p.pos) on both success and hard failure; callers that want to try
// speculatively should save/restore p.pos themselves.
func (p *sandboxParser) readStringLiteral() (value string, quote byte, end int, ok bool, err error) {
	if p.pos >= len(p.src) {
		return "", 0, 0, false, nil
	}
	c := p.src[p.pos]
	if c != '"' && c != '\'' {
		return "", 0, 0, false, nil
	}
	quote = c
	start := p.pos
	p.pos++
	var sb strings.Builder
	for {
		if p.pos >= len(p.src) {
			return "", 0, 0, false, fmt.Errorf("unterminated string literal starting at byte %d", start)
		}
		ch := p.src[p.pos]
		if ch == quote {
			p.pos++
			return sb.String(), quote, p.pos, true, nil
		}
		if ch == '\n' {
			return "", 0, 0, false, fmt.Errorf("unterminated string literal starting at byte %d (newline before closing quote)", start)
		}
		if ch != '\\' {
			sb.WriteByte(ch)
			p.pos++
			continue
		}

		// escape sequence
		p.pos++
		if p.pos >= len(p.src) {
			return "", 0, 0, false, fmt.Errorf("unterminated escape sequence at byte %d", p.pos)
		}
		e := p.src[p.pos]
		switch {
		case e == 'n':
			sb.WriteByte('\n')
			p.pos++
		case e == 't':
			sb.WriteByte('\t')
			p.pos++
		case e == 'r':
			sb.WriteByte('\r')
			p.pos++
		case e == 'a':
			sb.WriteByte(7)
			p.pos++
		case e == 'b':
			sb.WriteByte(8)
			p.pos++
		case e == 'f':
			sb.WriteByte(12)
			p.pos++
		case e == 'v':
			sb.WriteByte(11)
			p.pos++
		case e == '\\':
			sb.WriteByte('\\')
			p.pos++
		case e == '"':
			sb.WriteByte('"')
			p.pos++
		case e == '\'':
			sb.WriteByte('\'')
			p.pos++
		case e == '\n':
			sb.WriteByte('\n')
			p.pos++
		case e == 'x':
			if p.pos+2 >= len(p.src) {
				return "", 0, 0, false, fmt.Errorf("invalid \\x escape at byte %d", p.pos)
			}
			hex := string(p.src[p.pos+1 : p.pos+3])
			v, hexErr := strconv.ParseUint(hex, 16, 8)
			if hexErr != nil {
				return "", 0, 0, false, fmt.Errorf("invalid \\x escape %q at byte %d", hex, p.pos)
			}
			sb.WriteByte(byte(v))
			p.pos += 3
		case e >= '0' && e <= '9':
			j := p.pos
			digits := 0
			for digits < 3 && j < len(p.src) && p.src[j] >= '0' && p.src[j] <= '9' {
				j++
				digits++
			}
			numStr := string(p.src[p.pos:j])
			v, numErr := strconv.ParseUint(numStr, 10, 32)
			if numErr != nil || v > 255 {
				return "", 0, 0, false, fmt.Errorf("invalid decimal escape %q at byte %d", numStr, p.pos)
			}
			sb.WriteByte(byte(v))
			p.pos = j
		case e == 'z':
			p.pos++
			for p.pos < len(p.src) {
				switch p.src[p.pos] {
				case ' ', '\t', '\r', '\n':
					p.pos++
					continue
				}
				break
			}
		default:
			return "", 0, 0, false, fmt.Errorf("unsupported escape sequence '\\%c' at byte %d", e, p.pos)
		}
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

// scanBalanced scans a bracketed expression starting at the current position,
// which must be '(', '{' or '[', up to and including its matching close,
// honoring nested brackets, string/long-string literals and comments.
func (p *sandboxParser) scanBalanced() error {
	if p.pos >= len(p.src) {
		return fmt.Errorf("unexpected end of file while scanning bracketed expression")
	}
	if p.src[p.pos] == '[' {
		if level, isLong := p.tryLongBracketOpen(); isLong {
			_, err := p.consumeLongBracketBody(level)
			return err
		}
	}
	depth := 0
	for {
		if p.pos >= len(p.src) {
			return fmt.Errorf("unexpected end of file while scanning bracketed expression")
		}
		c := p.src[p.pos]
		switch {
		case c == '"' || c == '\'':
			_, _, _, ok, err := p.readStringLiteral()
			if err != nil {
				return err
			}
			if !ok {
				return fmt.Errorf("internal error reading string literal at byte %d", p.pos)
			}
		case c == '-' && p.pos+1 < len(p.src) && p.src[p.pos+1] == '-':
			p.pos += 2
			if level, isLong := p.tryLongBracketOpen(); isLong {
				if _, err := p.consumeLongBracketBody(level); err != nil {
					return err
				}
			} else {
				for p.pos < len(p.src) && p.src[p.pos] != '\n' {
					p.pos++
				}
			}
		case c == '[':
			if level, isLong := p.tryLongBracketOpen(); isLong {
				if _, err := p.consumeLongBracketBody(level); err != nil {
					return err
				}
				continue
			}
			depth++
			p.pos++
		case c == '(' || c == '{':
			depth++
			p.pos++
		case c == ')' || c == '}' || c == ']':
			depth--
			p.pos++
			if depth == 0 {
				return nil
			}
			if depth < 0 {
				return fmt.Errorf("unbalanced closing bracket at byte %d", p.pos)
			}
		default:
			p.pos++
		}
	}
}

// scanValueSpan scans a table-entry value expression starting at the current
// position (after skipping leading whitespace) up to - but not including - the
// first top-level ',' or the enclosing table's closing '}'. It returns the
// [start,end) byte range of the expression with trailing whitespace trimmed.
// It does not attempt to understand the expression, only to find its extent,
// so arrays, function calls, and other constructs we don't model as leaves are
// safely skipped over rather than causing a parse failure.
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
			_, _, _, ok, serr := p.readStringLiteral()
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
