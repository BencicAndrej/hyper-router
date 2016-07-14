package hyper

import (
	"bytes"
	"fmt"
	"math"
	"strings"
)

type treeNode struct {
	label    string
	handler  Handler
	children []*treeNode
}

// String() implements the Stringer interface,
// so we can easily see the current state of the
// tree, mostly for debugging purposes.
func (n *treeNode) String() string {
	return n.printTree(0)
}

// printTree is a utility method to pretty print
// the tree during the String() call.
func (n *treeNode) printTree(level int) string {
	buff := &bytes.Buffer{}
	if level > 1 {
		fmt.Fprint(buff, strings.Repeat("   ", int(math.Max(float64(0), float64(level-1)))))
	}
	if level > 0 {
		fmt.Fprintf(buff, "└── ")
	}

    hasHandler := ""
    if n.handler != nil {
        hasHandler = "✓"
    }
	fmt.Fprintf(buff, "%s %s\n", n.label, hasHandler)
	for _, child := range n.children {
		fmt.Fprintf(buff, child.printTree(level+1))
	}
	return string(buff.Bytes())
}

// getHandler returns the handler registered with the provided
// route, if it exists, returns nil otherwise
func (n *treeNode) getHandler(route string) Handler {
	// If the current route is longer than the requested
	// route, we don't have a handler registered with the
	// tree.
	if n.label > route {
		return nil
	}

	// If the routes match, return the handler.
	if n.label == route {
		return n.handler
	}

	offset := len(n.label)

	for _, child := range n.children {
		if child.label[0] == route[offset] {
			return child.getHandler(route[offset:])
		}
	}

	return nil
}

// insertNode associates tha handler with the label.
//
// Sequence of operations:
//
// #1 Find the common prefix size
//
// #2 If the prefix is shorter that the current node,
//    we need to split the current node, pass the handler to that child,
//    and add a new child with the new handler.
//
// #3 If the prefix is the same length as both the current node and
//    the provided label, and we have two handlers registered on the same
//    route, so we panic! Otherwise, we just register the new handler to
//    the current node.
//
// #4 If the prefix is longer that the current node, we look for a child
//    of the current node that starts with the first character of the
//    remainder of the label, and recursively continue until there are
//    no more children that match the label
//
// #5 If no children start with the same character as the next new label
//    character, we create a new child for the current node with the rest
//    of the label.
func (n *treeNode) insertNode(label string, handler Handler) {
	// Initialize the empty node element
	if len(n.label) == 0 && len(n.children) == 0 {
		n.label = label
		n.handler = handler
		return
	}

	// #1 Find the common prefix size
	prefixSize := 0

	switch true {
	case n.isCatchAll():
		//@TODO: Build full path here, for debugging purposes.
		panic("there is a catch all route registered for label " + label)
	case n.isParam():
		prefixSize = findPrefixLength(n.label, label, "/")

		if n.label != label[:prefixSize] {
			//@TODO: Build full path here, for debugging purposes.
			panic("route already defined for label '" + label + "' check for " + n.label)
		}
		break
	default:
		prefixSize = findPrefixLength(n.label, label, ":*")

		// #2 If the current node is not the whole prefix, we need
		// to split the current node into the prefix node and a child
		// node representing the current node.
		if len(n.label) > prefixSize {
			n.splitNode(prefixSize)
		}

		// #3 If the
		if len(label) == prefixSize {
			if n.handler != nil {
				//@TODO: Build full path here, for debugging purposes.
				panic("a handler is already registered for label " + label)
			}

			n.handler = handler
			return
		}
	}

	// #4
    var nextNode *treeNode
    for _, child := range n.children {
        if child.label[0] == byte('*') {
            nextNode = child
            break
        }
		if child.label[0] == label[prefixSize] && nextNode == nil {
            nextNode = child
		}
	}

    if nextNode != nil {
        nextNode.insertNode(label[prefixSize:], handler)
        return
    }

	// #5 There is no child that matches, create a new child.
	newChild := treeNode{
		label:   label[prefixSize:],
		handler: handler,
	}

	n.children = append(n.children, &newChild)
}

func (tree treeNode) isParam() bool {
	return len(tree.label) > 0 && tree.label[0] == ':'
}

func (tree treeNode) isCatchAll() bool {
	return len(tree.label) > 0 && tree.label[0] == '*'
}

// findPrefixLength returns the size of the longest common prefix,
// breaking early if encountering any of the breakAt characters.
func findPrefixLength(a, b string, breakAt string) int {
	i := 0
	max := min(len(a), len(b))
	for i < max && a[i] == b[i] && !strings.ContainsAny(string(a[i]), breakAt) {
		i++
	}

	return i
}

// Since there is no math.Min() for int type,
// only of specific size (int64, int32, ...) we must
// implement our own version that works with plain ints.
func min(a, b int) int {
	if a <= b {
		return a
	}

	return b
}

// splitNode splits the current node into two,
// in order to create room for future insertions,
// and moves the handler to the child node.
func (n *treeNode) splitNode(splitPoint int) *treeNode {
	child := treeNode{
		label:    n.label[splitPoint:],
		handler:  n.handler,
		children: n.children,
	}

	n.label = n.label[:splitPoint]
	n.handler = nil
	n.children = []*treeNode{&child}

	return &child
}
