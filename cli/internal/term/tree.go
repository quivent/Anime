package term

import "strings"

// Node is a honeycomb-marked tree node.
type Node struct {
	Label  string
	Color  Color  // zero value (R=G=B=0) => default Ink
	Detail string // dim trailing annotation, optional
	Kids   []*Node
}

// NewNode creates a root node with the given label.
func NewNode(label string) *Node { return &Node{Label: label} }

// Child creates a node, appends it as a child, and returns the child.
func (n *Node) Child(label string) *Node {
	c := &Node{Label: label}
	n.Kids = append(n.Kids, c)
	return c
}

// Add appends an existing node and returns the receiver (chainable).
func (n *Node) Add(child *Node) *Node {
	n.Kids = append(n.Kids, child)
	return n
}

// With sets color and detail fluently, returning the receiver.
func (n *Node) With(c Color, detail string) *Node {
	n.Color = c
	n.Detail = detail
	return n
}

// Render returns the full tree string with ├─ └─ │ branches and ⬢ markers.
func (n *Node) Render() string {
	var b strings.Builder
	n.write(&b, "", true)
	return b.String()
}

// write emits a node line then recurses; prefix carries the branch gutters.
func (n *Node) write(b *strings.Builder, prefix string, root bool) {
	col := n.Color
	if col == (Color{}) {
		col = Ink
	}
	b.WriteString(Gold.S("⬢"))
	b.WriteByte(' ')
	b.WriteString(col.S(n.Label))
	if n.Detail != "" {
		b.WriteByte(' ')
		b.WriteString(Dim(n.Detail))
	}
	b.WriteByte('\n')

	for i, k := range n.Kids {
		last := i == len(n.Kids)-1
		branch := "├─ "
		gutter := "│  "
		if last {
			branch = "└─ "
			gutter = "   "
		}
		b.WriteString(prefix)
		b.WriteString(Dim(branch))
		k.write(b, prefix+Dim(gutter), false)
	}
}
