package config

import (
	"fmt"

	"github.com/arnodel/golua/ast"
	"github.com/arnodel/golua/ops"
	"github.com/arnodel/golua/parsing"
	"github.com/arnodel/golua/scanner"
)

// parseSandboxFile parses a SandboxVars.lua file into a tree rooted at the
// SandboxVars table. Structure, nesting and literal classification (is this
// value an integer, a float, a bool, a string, a nested table, or something
// else entirely) are delegated to golua's real Lua grammar
// (github.com/arnodel/golua), which fails closed on any genuine syntax
// problem - unbalanced brackets, unterminated strings, unsupported escapes -
// with a proper line:column error, since those mean we can no longer be sure
// where entries begin and end.
//
// Two things golua's AST does not give us directly are recovered locally by
// re-scanning small, precisely-bounded slices of the original source with
// golua's own scanner/lexer primitives rather than by hand-rolling Lua
// grammar ourselves:
//   - The exact byte end of a scalar literal. golua's Location only records a
//     single-token node's *start* (it was built for error messages, not
//     source-preserving edits), so scalarTokenEnd re-scans just that one
//     token to learn its length.
//   - Leading comment text. Every standard Lua tokenizer - golua's included -
//     discards comments as trivia, so descriptions are recovered by scanning
//     the gap between one field's end and the next field's start for
//     contiguous "--" runs.
//
// It is lenient about *values* it can't classify as a scalar (arrays,
// function calls, arithmetic/concat expressions): those are parsed by golua
// like everything else, just not exposed as a leaf. Their bytes are left
// untouched, but we still need their extent (via the value-span scanner) so
// comment/gap-scanning for the *next* key resumes in the right place.
func parseSandboxFile(src []byte) (*sandboxNode, error) {
	block, err := parsing.ParseChunk(scanner.New("SandboxVars.lua", src))
	if err != nil {
		return nil, fmt.Errorf("lua syntax error: %w", err)
	}
	if len(block.Stats) == 0 {
		return nil, fmt.Errorf("expected 'SandboxVars = { ... }' as the file's only statement")
	}
	assign, ok := block.Stats[0].(ast.AssignStat)
	if !ok || len(assign.Dest) != 1 || len(assign.Src) != 1 {
		return nil, fmt.Errorf("expected 'SandboxVars = { ... }' as the file's only statement")
	}
	name, ok := assign.Dest[0].(ast.Name)
	if !ok || name.Val != "SandboxVars" {
		return nil, fmt.Errorf("expected 'SandboxVars' identifier at start of file")
	}
	table, ok := assign.Src[0].(ast.TableConstructor)
	if !ok {
		return nil, fmt.Errorf("expected 'SandboxVars' to be assigned a table literal")
	}

	root, _, err := sandboxTableToNode(src, table)
	if err != nil {
		return nil, err
	}
	return root, nil
}

// sandboxTableToNode converts one already-parsed table literal into a
// sandboxNode, and returns the byte offset just past its closing '}' so the
// caller can resume comment/gap scanning immediately after it.
func sandboxTableToNode(src []byte, table ast.TableConstructor) (*sandboxNode, int, error) {
	node := newSandboxGroupNode()
	gapStart := table.StartPos().Offset + 1 // just after '{'

	for _, field := range table.Fields {
		description, err := extractSandboxDescription(src, gapStart)
		if err != nil {
			return nil, 0, err
		}

		key, isStringKey := field.Key.(ast.String)
		valueStart := field.Value.Locate().StartPos().Offset

		var (
			child    *sandboxNode
			valueEnd int
		)
		switch value := field.Value.(type) {
		case ast.TableConstructor:
			nested, end, nestedErr := sandboxTableToNode(src, value)
			if nestedErr != nil {
				return nil, 0, nestedErr
			}
			valueEnd = end
			if isStringKey {
				nested.Description = description
				child = nested
			}
		case ast.Bool:
			valueEnd = scalarTokenEnd(src, valueStart)
			if isStringKey {
				child = newSandboxLeafNode(sandboxLeafBool, src, valueStart, valueEnd, description)
			}
		case ast.Int:
			valueEnd = scalarTokenEnd(src, valueStart)
			if isStringKey {
				child = newSandboxLeafNode(sandboxLeafInteger, src, valueStart, valueEnd, description)
			}
		case ast.Float:
			valueEnd = scalarTokenEnd(src, valueStart)
			if isStringKey {
				child = newSandboxLeafNode(sandboxLeafFloat, src, valueStart, valueEnd, description)
			}
		case *ast.UnOp:
			// Lua has no negative-number token: `-3` parses as unary minus
			// applied to `3`. Fuse the two back into a single leaf spanning
			// "-3" (matching the original parser, which recognized negative
			// numerals as one literal) only when it's a bare "-<number>" with
			// nothing between the sign and the digits; anything else (a
			// negated variable, `not x`, a space before the digits) is left
			// as an unsupported expression, exactly as the original regex-
			// based classification would have treated it.
			operandStart := value.Operand.Locate().StartPos().Offset
			_, operandIsNumber := value.Operand.(ast.Int)
			_, operandIsFloat := value.Operand.(ast.Float)
			if value.Op == ops.OpNeg && operandStart == valueStart+1 && (operandIsNumber || operandIsFloat) {
				valueEnd = scalarTokenEnd(src, operandStart)
				if isStringKey {
					kind := sandboxLeafInteger
					if operandIsFloat {
						kind = sandboxLeafFloat
					}
					child = newSandboxLeafNode(kind, src, valueStart, valueEnd, description)
				}
			} else {
				valueEnd, err = measureUnsupportedSandboxValue(src, valueStart)
				if err != nil {
					return nil, 0, err
				}
			}
		case ast.String:
			// Only a plain quoted literal ('"...' / '\'...') is treated as an
			// editable leaf: a long-bracket string value ([[...]]) has no
			// natural QuoteChar to preserve on update, so - matching the
			// original parser's behavior - we leave it untouched instead.
			if src[valueStart] == '"' || src[valueStart] == '\'' {
				valueEnd = scalarTokenEnd(src, valueStart)
				if isStringKey {
					child = newSandboxStringLeafNode(string(value.Val), src[valueStart], valueStart, valueEnd, description)
				}
			} else {
				valueEnd, err = measureUnsupportedSandboxValue(src, valueStart)
				if err != nil {
					return nil, 0, err
				}
			}
		default:
			// Function calls, arithmetic/concat expressions, unary ops, etc:
			// a shape we don't model as a leaf. Preserve the bytes untouched
			// and just measure past them.
			valueEnd, err = measureUnsupportedSandboxValue(src, valueStart)
			if err != nil {
				return nil, 0, err
			}
		}

		if isStringKey && child != nil {
			node.addChild(string(key.Val), child)
		}

		gapStart = consumeSandboxTrailingComma(src, valueEnd)
	}

	return node, table.EndPos().Offset, nil
}

// newSandboxLeafNode builds a leaf whose Value is the raw source text of
// [start,end) - correct for bool/int/float, whose written form we must
// preserve exactly (e.g. "0x1F", "2.50"). String leaves use
// newSandboxStringLeafNode instead, since their Value is the decoded content,
// not the raw quoted text.
func newSandboxLeafNode(kind sandboxLeafKind, src []byte, start, end int, description string) *sandboxNode {
	return &sandboxNode{Leaf: &sandboxLeaf{
		Kind:        kind,
		Value:       string(src[start:end]),
		Description: description,
		ValueStart:  start,
		ValueEnd:    end,
	}}
}

func newSandboxStringLeafNode(decoded string, quote byte, start, end int, description string) *sandboxNode {
	return &sandboxNode{Leaf: &sandboxLeaf{
		Kind:        sandboxLeafString,
		Value:       decoded,
		QuoteChar:   quote,
		Description: description,
		ValueStart:  start,
		ValueEnd:    end,
	}}
}

// scalarTokenEnd returns the byte offset just past the single token located
// at src[start:] (a boolean, number or string literal). It reuses golua's own
// scanner rather than reimplementing Lua's numeral/string/escape grammar: a
// token's raw text (before any escape decoding) is exactly
// src[start:start+len(Lit)], since Lua's lexical grammar never depends on
// anything past the token itself.
func scalarTokenEnd(src []byte, start int) int {
	tok := scanner.New("", src[start:]).Scan()
	return start + len(tok.Lit)
}

// measureUnsupportedSandboxValue finds the byte extent of a table-entry value
// golua parsed successfully but that we don't expose as a leaf, so that
// comment/gap scanning for the next key resumes in the right place without
// corrupting a byte splice for anything nearby (e.g. a function call
// argument containing a string with a literal ',' or '}' inside it).
func measureUnsupportedSandboxValue(src []byte, start int) (int, error) {
	p := &sandboxParser{src: src, pos: start}
	_, end, err := p.scanValueSpan()
	if err != nil {
		return 0, err
	}
	return end, nil
}

// extractSandboxDescription collects the comment text (if any) starting at
// byte offset from, stopping at the first non-whitespace, non-comment byte -
// which in a well-formed file is exactly the start of the next field.
func extractSandboxDescription(src []byte, from int) (string, error) {
	p := &sandboxParser{src: src, pos: from}
	return p.skipInsignificant()
}

// consumeSandboxTrailingComma skips an optional ',' (and any whitespace
// before it) right after a value, returning the offset to resume scanning
// from for the next field.
func consumeSandboxTrailingComma(src []byte, from int) int {
	p := &sandboxParser{src: src, pos: from}
	p.consumeOptionalComma()
	return p.pos
}
