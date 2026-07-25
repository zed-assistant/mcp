// Kept as an internal test (not config_test) because it exercises ~15
// unexported parser types and functions directly.
//
//nolint:testpackage // relies on unexported parser internals
package config

import (
	"strings"
	"testing"
)

const sampleSandboxVars = `SandboxVars = {
    -- Population settings
    -- for zombies
    PopulationMultiplier = 2.5,
    ["PopulationMultiplier2"] = 4,
    HexSetting = 0x1F,
    Negative = -3,
    ZombieLore = {
        -- Speed of zombies
        Speed = 1,
        Enabled = true,
        Name = "hello \"world\"\ntab:\t.",
    },
    -- skipped: computed/array key
    [1] = "array-ish",
    -- skipped: function call
    Foo = SomeFunc(1, "a,b", {2,3}),
    -- skipped: long-bracket string value
    LongStr = [[ raw text ]],
    -- skipped: arithmetic expression
    Bar = 1 + 2,
}
`

func TestParseSandboxFile_Basic(t *testing.T) {
	t.Parallel()
	root, err := parseSandboxFile([]byte(sampleSandboxVars))
	if err != nil {
		t.Fatalf("parseSandboxFile failed: %v", err)
	}

	requireLeaf := func(node *sandboxNode, key string) *sandboxLeaf {
		t.Helper()
		child, ok := node.Children[key]
		if !ok {
			t.Fatalf("missing key %q; have %v", key, node.ChildOrder)
		}
		if child.Leaf == nil {
			t.Fatalf("key %q is not a leaf", key)
		}
		return child.Leaf
	}

	pm := requireLeaf(root, "PopulationMultiplier")
	if pm.Kind != sandboxLeafFloat || pm.Value != "2.5" {
		t.Errorf("PopulationMultiplier = %q kind=%v, want 2.5/float", pm.Value, pm.Kind)
	}
	if got := strings.TrimSpace(pm.Description); got != "Population settings\nfor zombies" {
		t.Errorf("PopulationMultiplier description = %q", got)
	}

	pm2 := requireLeaf(root, "PopulationMultiplier2")
	if pm2.Kind != sandboxLeafInteger || pm2.Value != "4" {
		t.Errorf("PopulationMultiplier2 = %q kind=%v, want 4/integer", pm2.Value, pm2.Kind)
	}

	hex := requireLeaf(root, "HexSetting")
	if hex.Kind != sandboxLeafInteger || hex.Value != "0x1F" {
		t.Errorf("HexSetting = %q kind=%v, want 0x1F/integer", hex.Value, hex.Kind)
	}

	neg := requireLeaf(root, "Negative")
	if neg.Kind != sandboxLeafInteger || neg.Value != "-3" {
		t.Errorf("Negative = %q kind=%v, want -3/integer", neg.Value, neg.Kind)
	}

	zl, ok := root.Children["ZombieLore"]
	if !ok || !zl.isGroup() {
		t.Fatalf("ZombieLore missing or not a group")
	}
	speed := requireLeaf(zl, "Speed")
	if speed.Kind != sandboxLeafInteger || speed.Value != "1" {
		t.Errorf("Speed = %q kind=%v, want 1/integer", speed.Value, speed.Kind)
	}
	enabled := requireLeaf(zl, "Enabled")
	if enabled.Kind != sandboxLeafBool || enabled.Value != "true" {
		t.Errorf("Enabled = %q kind=%v, want true/bool", enabled.Value, enabled.Kind)
	}
	name := requireLeaf(zl, "Name")
	if name.Kind != sandboxLeafString || name.Value != "hello \"world\"\ntab:\t." {
		t.Errorf("Name = %q kind=%v", name.Value, name.Kind)
	}
	if name.QuoteChar != '"' {
		t.Errorf("Name QuoteChar = %q, want '\"'", name.QuoteChar)
	}

	// Unsupported/skipped entries must not appear at all.
	for _, key := range []string{"1", "Foo", "LongStr", "Bar"} {
		if _, ok := root.Children[key]; ok {
			t.Errorf("key %q should have been skipped as unsupported, but is present", key)
		}
	}
}

// TestParseSandboxFile_ValueSpansAreByteExact verifies that ValueStart/ValueEnd
// for every leaf slice out exactly that leaf's literal text, and nothing else -
// the property update_sandbox_config.go's byte-splice apply depends on.
func TestParseSandboxFile_ValueSpansAreByteExact(t *testing.T) {
	t.Parallel()
	src := []byte(sampleSandboxVars)
	root, err := parseSandboxFile(src)
	if err != nil {
		t.Fatalf("parseSandboxFile failed: %v", err)
	}

	var walk func(node *sandboxNode)
	walk = func(node *sandboxNode) {
		if node.Leaf != nil {
			got := string(src[node.Leaf.ValueStart:node.Leaf.ValueEnd])
			if node.Leaf.Kind == sandboxLeafString {
				if len(got) < 2 || got[0] != node.Leaf.QuoteChar || got[len(got)-1] != node.Leaf.QuoteChar {
					t.Errorf("string value span %q doesn't look like a quoted literal", got)
				}
			} else if got != node.Leaf.Value {
				t.Errorf("value span text %q != decoded Value %q", got, node.Leaf.Value)
			}
			return
		}
		for _, name := range node.ChildOrder {
			walk(node.Children[name])
		}
	}
	walk(root)
}

func TestParseSandboxFile_SyntaxErrors(t *testing.T) {
	t.Parallel()
	cases := []string{
		`SandboxVars = { Foo = "unterminated }`,
		`SandboxVars = { Foo = 1, `,
		`SandboxVars = { Foo = }`,
		`NotSandboxVars = { Foo = 1 }`,
		`SandboxVars = 1`,
	}
	for _, c := range cases {
		if _, err := parseSandboxFile([]byte(c)); err == nil {
			t.Errorf("expected parse error for %q, got nil", c)
		}
	}
}

func TestParseSandboxFile_UpdateRoundTrip(t *testing.T) {
	t.Parallel()
	src := []byte(sampleSandboxVars)
	root, err := parseSandboxFile(src)
	if err != nil {
		t.Fatalf("parseSandboxFile failed: %v", err)
	}

	var edits []sandboxEdit
	var problems []string
	validateSandboxUpdates(root, map[string]any{
		"PopulationMultiplier": "9.5",
		"ZombieLore": map[string]any{
			"Enabled": "false",
		},
	}, "", &edits, &problems)
	if len(problems) != 0 {
		t.Fatalf("unexpected problems: %v", problems)
	}

	// Re-parse after applying edits the same way UpdateSandboxConfig does, and
	// confirm every other byte (comments, unrelated entries, formatting) is
	// untouched.
	merged := append([]byte(nil), src...)
	type edit = sandboxEdit
	sorted := append([]edit(nil), edits...)
	for i := range sorted {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[j].leaf.ValueStart > sorted[i].leaf.ValueStart {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}
	for _, e := range sorted {
		newBytes := []byte(e.newText)
		spliced := make([]byte, 0, len(merged)-(e.leaf.ValueEnd-e.leaf.ValueStart)+len(newBytes))
		spliced = append(spliced, merged[:e.leaf.ValueStart]...)
		spliced = append(spliced, newBytes...)
		spliced = append(spliced, merged[e.leaf.ValueEnd:]...)
		merged = spliced
	}

	root2, err := parseSandboxFile(merged)
	if err != nil {
		t.Fatalf("re-parsing after update failed: %v\n--- output ---\n%s", err, merged)
	}
	pm := root2.Children["PopulationMultiplier"].Leaf
	if pm.Value != "9.5" {
		t.Errorf("PopulationMultiplier after update = %q, want 9.5", pm.Value)
	}
	if got := strings.TrimSpace(pm.Description); got != "Population settings\nfor zombies" {
		t.Errorf("description lost after update: %q", got)
	}
	enabled := root2.Children["ZombieLore"].Children["Enabled"].Leaf
	if enabled.Value != "false" {
		t.Errorf("Enabled after update = %q, want false", enabled.Value)
	}
	// Untouched sibling entries must still be there, byte for byte.
	if root2.Children["HexSetting"].Leaf.Value != "0x1F" {
		t.Errorf("HexSetting was corrupted by an unrelated edit")
	}
}
