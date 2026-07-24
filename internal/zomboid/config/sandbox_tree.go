package config

// sandboxLeafKind identifies the Lua literal type backing a sandboxLeaf, so that
// updates can be validated and re-serialized using the same literal form as the
// original value (a plain string update must never silently change a field from,
// say, a boolean to a string).
type sandboxLeafKind int

const (
	// sandboxLeafInteger is a whole number written without a decimal point or
	// exponent (e.g. `4`, `-1`, `0x1F`). Kept distinct from sandboxLeafFloat so
	// that an update can never silently turn a whole-number setting into a
	// fractional one (which PZ's Java-side option loader may not tolerate).
	sandboxLeafInteger sandboxLeafKind = iota
	sandboxLeafFloat
	sandboxLeafBool
	sandboxLeafString
)

// jsonType names the ConfigEntry.Type value that should be reported for a leaf
// of this kind. "integer"/"number" follow JSON Schema's own convention, where
// "integer" is the whole-number subset of "number".
func (k sandboxLeafKind) jsonType() string {
	switch k {
	case sandboxLeafInteger:
		return "integer"
	case sandboxLeafFloat:
		return "number"
	case sandboxLeafBool:
		return "boolean"
	case sandboxLeafString:
		return "string"
	default:
		return ""
	}
}

// sandboxLeaf is a scalar entry in the SandboxVars table. ValueStart/ValueEnd are
// byte offsets into the original source identifying exactly the value literal
// (e.g. `4`, `true`, `"foo"` including quotes) so an update can be applied as a
// surgical byte-splice that leaves every other byte of the file untouched.
type sandboxLeaf struct {
	Kind        sandboxLeafKind
	Value       string // decoded value: raw literal text for number/bool, unescaped content for string
	QuoteChar   byte   // meaningful only for sandboxLeafString: '"' or '\''
	Description string
	ValueStart  int
	ValueEnd    int
}

// sandboxNode is either a scalar leaf (Leaf != nil) or a group/table of further
// named nodes (Children != nil). A Lua table entry whose value we can't confidently
// classify as a recognized scalar (arrays, function calls, etc.) is simply omitted
// from its parent's Children - it stays byte-for-byte in the file but is neither
// readable nor updatable through this tool.
type sandboxNode struct {
	Leaf        *sandboxLeaf
	Children    map[string]*sandboxNode
	ChildOrder  []string
	Description string
}

func newSandboxGroupNode() *sandboxNode {
	return &sandboxNode{Children: make(map[string]*sandboxNode)}
}

func (n *sandboxNode) isGroup() bool {
	return n.Children != nil
}

func (n *sandboxNode) addChild(name string, child *sandboxNode) {
	if _, exists := n.Children[name]; !exists {
		n.ChildOrder = append(n.ChildOrder, name)
	}
	n.Children[name] = child
}
