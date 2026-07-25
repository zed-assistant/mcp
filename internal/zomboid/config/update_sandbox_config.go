package config

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"

	"github.com/zed-assistant/mcp/internal/configuration"
	domainerror "github.com/zed-assistant/mcp/internal/domain_error"
)

var (
	decimalIntegerRegex = regexp.MustCompile(`^-?\d+$`)
	decimalNumberRegex  = regexp.MustCompile(`^-?(\d+\.\d*|\.\d+|\d+)([eE][-+]?\d+)?$`)
	hexNumberRegex      = regexp.MustCompile(`^-?0[xX][0-9a-fA-F]+$`)
)

// isSandboxNumberLiteral reports whether s is any recognized Lua numeric
// literal, integer or float, that a user may submit as a replacement value.
func isSandboxNumberLiteral(s string) bool {
	return decimalNumberRegex.MatchString(s) || hexNumberRegex.MatchString(s)
}

// isSandboxIntegerLiteral reports whether s is specifically a whole-number
// literal (no decimal point, no exponent). Hex literals are always integers.
func isSandboxIntegerLiteral(s string) bool {
	return decimalIntegerRegex.MatchString(s) || hexNumberRegex.MatchString(s)
}

// sandboxEdit is one validated, ready-to-apply change: replace exactly the
// bytes [leaf.ValueStart, leaf.ValueEnd) of the source file with newText.
type sandboxEdit struct {
	leaf    *sandboxLeaf
	newText string
}

// validateSandboxUpdates walks a (possibly nested) requested update tree
// alongside the parsed sandbox tree, collecting every problem it finds -
// unknown keys, group/leaf shape mismatches, values of the wrong type/kind -
// into problems, and every valid leaf change into edits. It never stops at the
// first problem: the caller must reject the whole update if problems is
// non-empty, so every issue can be reported at once and nothing is partially
// applied.
func validateSandboxUpdates(node *sandboxNode, updates map[string]any, pathPrefix string, edits *[]sandboxEdit, problems *[]string) {
	for key, rawVal := range updates {
		path := key
		if pathPrefix != "" {
			path = pathPrefix + "." + key
		}

		child, exists := node.Children[key]
		if !exists {
			*problems = append(*problems, fmt.Sprintf("unknown key '%s'", path))
			continue
		}

		if child.isGroup() {
			nested, ok := rawVal.(map[string]any)
			if !ok {
				*problems = append(*problems, fmt.Sprintf("'%s' is a group of settings; provide a nested object to update entries within it, got %s", path, jsonTypeName(rawVal)))
				continue
			}
			validateSandboxUpdates(child, nested, path, edits, problems)
			continue
		}

		strVal, ok := rawVal.(string)
		if !ok {
			*problems = append(*problems, fmt.Sprintf("'%s': value must be provided as a JSON string, got %s", path, jsonTypeName(rawVal)))
			continue
		}

		newText, err := formatSandboxLeafUpdate(child.Leaf, strVal)
		if err != nil {
			*problems = append(*problems, fmt.Sprintf("'%s': %s", path, err.Error()))
			continue
		}

		*edits = append(*edits, sandboxEdit{leaf: child.Leaf, newText: newText})
	}
}

// formatSandboxLeafUpdate validates newValue against the existing leaf's Lua
// literal kind and, on success, returns the exact replacement source text for
// that leaf's value span. A field's type (integer/float/bool/string) can never
// change via update - only its value. Integer fields specifically reject
// fractional input (e.g. "4.5"), since PZ's Java-side option loader may not
// tolerate a whole-number setting suddenly becoming fractional; float fields
// accept whole-number input (e.g. "1" for a field previously "1.0") since that
// is numerically harmless.
func formatSandboxLeafUpdate(leaf *sandboxLeaf, newValue string) (string, error) {
	switch leaf.Kind {
	case sandboxLeafInteger:
		trimmed := strings.TrimSpace(newValue)
		if !isSandboxIntegerLiteral(trimmed) {
			return "", errors.New("must be a whole number (no decimal point or exponent)")
		}
		return trimmed, nil
	case sandboxLeafFloat:
		trimmed := strings.TrimSpace(newValue)
		if !isSandboxNumberLiteral(trimmed) {
			return "", errors.New("must be a valid number")
		}
		if decimalIntegerRegex.MatchString(trimmed) {
			// Numerically identical in Lua either way, but keep the literal
			// syntactically a float (e.g. "2" -> "2.0") so a later read still
			// reports type='number' rather than flipping to 'integer'.
			trimmed += ".0"
		}
		return trimmed, nil
	case sandboxLeafBool:
		trimmed := strings.TrimSpace(newValue)
		if trimmed != "true" && trimmed != "false" {
			return "", errors.New("must be 'true' or 'false'")
		}
		return trimmed, nil
	case sandboxLeafString:
		return encodeLuaString(newValue, leaf.QuoteChar), nil
	default:
		return "", errors.New("unsupported value kind")
	}
}

// encodeLuaString re-quotes s using the same quote character as the original
// literal, escaping backslashes, that quote character, and control characters.
func encodeLuaString(s string, quote byte) string {
	var sb strings.Builder
	sb.WriteByte(quote)
	for _, r := range s {
		switch {
		case r == '\\':
			sb.WriteString(`\\`)
		case r == '\n':
			sb.WriteString(`\n`)
		case r == '\r':
			sb.WriteString(`\r`)
		case r == '\t':
			sb.WriteString(`\t`)
		case r == rune(quote):
			sb.WriteByte('\\')
			sb.WriteByte(quote)
		case r < 0x20:
			fmt.Fprintf(&sb, "\\%d", r)
		default:
			sb.WriteRune(r)
		}
	}
	sb.WriteByte(quote)
	return sb.String()
}

func (m *ConfigManager) UpdateSandboxConfig(instanceConfig *configuration.ZomboidInstanceConfig, updates map[string]any) error {
	root, path, src, mode, err := loadSandboxFile(instanceConfig.HomeDir, instanceConfig.ServerName)
	if err != nil {
		return fmt.Errorf("unable to load sandbox config for update: %w", err)
	}

	var edits []sandboxEdit
	var problems []string
	validateSandboxUpdates(root, updates, "", &edits, &problems)
	if len(problems) > 0 {
		return NewInvalidConfigUpdateError(problems)
	}

	// Apply edits as byte-splices only over each leaf's exact value span, from
	// the end of the file backward, so every other byte (comments, whitespace,
	// unrelated entries) is left completely untouched and earlier offsets stay
	// valid as we go.
	sort.Slice(edits, func(i, j int) bool {
		return edits[i].leaf.ValueStart > edits[j].leaf.ValueStart
	})

	out := append([]byte(nil), src...)
	for _, e := range edits {
		newBytes := []byte(e.newText)
		merged := make([]byte, 0, len(out)-(e.leaf.ValueEnd-e.leaf.ValueStart)+len(newBytes))
		merged = append(merged, out[:e.leaf.ValueStart]...)
		merged = append(merged, newBytes...)
		merged = append(merged, out[e.leaf.ValueEnd:]...)
		out = merged
	}

	// Our edits' byte offsets were computed against `src`. If the file on disk
	// no longer matches `src` - the game server rewrote it, an admin edited it
	// by hand, another process raced us - writing `out` now would silently
	// discard whatever changed it. Re-check right before saving and fail
	// loudly instead of clobbering; this can't close the window entirely (a
	// write could still land between this check and our own), but it turns
	// the common case of "something else touched the file" into a clear,
	// retryable error instead of silent data loss.
	if err := ensureSandboxFileUnchanged(path, src); err != nil {
		return err
	}

	if err := os.WriteFile(path, out, mode); err != nil {
		return fmt.Errorf("failed to save sandbox lua file: %w", err)
	}

	return nil
}

// ensureSandboxFileUnchanged re-reads path and confirms it still matches
// expected byte-for-byte.
func ensureSandboxFileUnchanged(path string, expected []byte) error {
	current, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("unable to re-read sandbox lua file before saving: %w", err)
	}
	if !bytes.Equal(current, expected) {
		return &domainerror.DomainError{
			InternalMessage: "sandbox config file changed on disk since it was read for this update",
			PublicMessage:   "The sandbox config file changed on disk after it was read for this update (e.g. by the game server or another process). The update was NOT applied, to avoid discarding that change. Read the config again and retry.",
			InternalCode:    domainerror.Conflict,
		}
	}
	return nil
}
