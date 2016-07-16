package hyper

import (
	"bytes"
	"context"
	"fmt"
	"math"
	"net/http"
	"strings"

    "github.com/bencicandrej/hyper-router/params"
)

type node struct {
	label   nodeLabel
	handler http.Handler

	parent *node
	// children represents an array of child nodes, ordered by priority:
	// wildcard then parameters then static.
	children []*node
}

func (tree node) getHandler(ctx context.Context, label nodeLabel) (http.Handler, context.Context) {
	if tree.isEmpty() {
		return nil, ctx
	}

	if tree.isWildcard() {
		return tree.handler, params.NewContext(ctx, string(tree.label)[1:], string(label))
	}

	if tree.isParameter() {
		paramEnd, finishedBeforeEnd := label.getEndOfVariable()
		if !finishedBeforeEnd {
			return tree.handler, params.NewContext(ctx, string(tree.label)[1:], string(label))
		}

		for _, child := range tree.children {
			if child.supports(label[paramEnd:]) {
				handler, ctx := child.getHandler(ctx, label[paramEnd:])

				return handler, params.NewContext(ctx, string(tree.label)[1:], string(label))
			}
		}

		return nil, ctx
	}

	// node is static
	if match, fullMatch := tree.matches(label); match {
		if fullMatch {
			return tree.handler, ctx
		}

		treeLen := len(tree.label)
		for _, child := range tree.children {
			if child.supports(label[treeLen:]) {
				return child.getHandler(ctx, label[treeLen:])
			}
		}
	}

	return nil, ctx
}

// insert associates the new handler with the route provided,
// and panics if encounters any anomalies.
//
// Method flow:
// #1) If the current node is empty, populate it and exit.
// #2) If label and tree.label are equal, we match or panic if handler exists.
// #3) If the prefix is equal to the label, we must split the node and associate the handler with the parent node.
// #4) If the prefix < label && prefix < tree.label && prefix > 0, split and pass to new node.
// #5) If the prefix is equal to the tree.label, we must create a new node, or pass insertion to a child
// #6) If the prefix is equal to 0, we must panic
func (tree *node) insert(label nodeLabel, handler http.Handler) *node {
	// #1) If tree is empty populate the current element.
	if tree.isEmpty() {
		// Label must start with a '/'.
		if !label.isValidRootLabel() {
			panic(fmt.Sprintf("route '%s%s' must start with '/'", tree.prefix(), label))
		}

		if variablePos, ok := label.getVariable(); ok {
			// Will never be 0 because a '/' is required to be first.
			tree.label = label[:variablePos]

			return tree.insert(label[variablePos:], handler)
		}

		tree.label = label
		tree.handler = handler
		return tree
	}

	if parameterPos, ok := label.getParameter(); ok {
		if parameterPos == 0 {
			// Find end of parameter
			parameterEnd, finishedBeforeEnd := label.getEndOfVariable()

			if len(tree.children) > 1 {
				panic(fmt.Sprintf("handler for route '%s%s' already exists", tree.path(), label))
			}

			if len(tree.children) == 1 {
				child := tree.children[0]

				if child.label != label[:parameterEnd] || parameterEnd == len(label) {
					panic(fmt.Sprintf("handler for route '%s%s' already exists", tree.path(), label))
				}

				return child.insert(label[parameterEnd:], handler)
			}

			newNode := node{
				label: label[:parameterEnd],

				parent: tree,
			}

			// Insert new node at the start of the children nodes.
			tree.children = append([]*node{&newNode}, tree.children...)

			if finishedBeforeEnd {
				return newNode.insert(label[parameterEnd:], handler)
			}

			newNode.handler = handler
			return &newNode
		}

		newNode := tree.insert(label[:parameterPos], nil)

		return newNode.insert(label[parameterPos:], handler)
	}

	if wildcardPos, ok := label.getWildcard(); ok {
		if wildcardPos == 0 {
			if _, finishedBeforeEnd := label.getEndOfVariable(); finishedBeforeEnd {
				panic(
					fmt.Sprintf("wildcard parameter must be the last element of the route '%s%s'", tree.prefix(), label),
				)
			}

			if len(tree.children) > 0 {
				panic(fmt.Sprintf("handler for route '%s%s' already exists", tree.prefix(), label))
			}

			newNode := node{
				label:   label,
				handler: handler,

				parent: tree,
			}

			// Insert new node at the start of the children nodes.
			tree.children = append([]*node{&newNode}, tree.children...)

			return &newNode
		}

		newNode := tree.insert(label[:wildcardPos], nil)

		return newNode.insert(label[wildcardPos:], handler)
	}

	// #2) If we get a route match and the handler slot is free,
	// we populate it and return.
	if tree.label == label && tree.handler == nil {
		tree.handler = handler
		return tree
	} else if tree.label == label && tree.handler != nil {
		if handler == nil {
			return tree
		}

		panic(fmt.Sprintf("handler for route '%s' already exists", label))
	}

	// Find the common prefix for the two labels.
	prefixLength := tree.label.findPrefixLength(label)

	// #3) If the tree.label is longer that the label and the label
	// is contained inside the tree.label, we split the node and
	// associate the handler to the current node.
	if tree.canSplit() && len(label) == prefixLength {
		tree.split(prefixLength)
		tree.handler = handler
		return tree
	}

	// #4) If the current node can be split, and the common prefix
	// is shorted than the label of the current node, we split
	// the current node and continue insertion.
	if tree.canSplit() && 0 < prefixLength && prefixLength < len(tree.label) {
		tree.split(prefixLength)
	}

	for _, child := range tree.children {
		if child.isWildcard() || child.isParameter() {
			panic(fmt.Sprintf("handler for route '%s' already exists", label))
		}
		if child.label[0] == label[prefixLength] {
			return child.insert(label[prefixLength:], handler)
		}
	}

	newNode := node{
		label:   label[prefixLength:],
		handler: handler,
		parent:  tree,
	}

	tree.children = append(tree.children, &newNode)

	return &newNode
}

// canSplit tests whether the current node can be divided into at least two nodes
func (tree node) canSplit() bool {
	return len(tree.label) > 1
}

// split separates the current node into two parts,
// the length of which is specified by the splitting point.
func (tree *node) split(splitPoint int) *node {
	newNode := node{
		label:   tree.label[splitPoint:],
		handler: tree.handler,

		parent:   tree,
		children: tree.children,
	}

	tree.label = tree.label[:splitPoint]
	tree.handler = nil
	tree.children = []*node{&newNode}

	return &newNode
}

// supports checks if the node can support the provided label.
// To do so, the node must be either a wildcard, a parameter,
// or starts with the same character (in the tree, it is not
// possible to have two child nodes that start with the same
// character).
func (tree node) supports(label nodeLabel) bool {
	return tree.isWildcard() || tree.isParameter() || tree.label[0] == byte(label[0])
}

// matches checks if the provided label is prefixed with the
// current nodes label.
func (tree node) matches(label nodeLabel) (match bool, fullMatch bool) {
	treeLen := len(tree.label)
	cmpLen := len(label)

	if treeLen > cmpLen {
		return false, false
	}

	for i := 0; i < treeLen; i++ {
		if tree.label[i] != label[i] {
			return false, false
		}
	}

	if treeLen == cmpLen {
		return true, true
	}

	return true, false
}

// exactlyMatches checks if the current node's label is
// the complete match of the label provided.
func (tree node) exactlyMatches(label nodeLabel) bool {
	return string(tree.label) == string(label)
}

// isEmpty checks if the tree node is empty.
func (tree node) isEmpty() bool {
	return tree.label == "" && tree.handler == nil
}

// isWildcard checks if the node is marked with
// a catchall character at the start of the string.
func (tree node) isWildcard() bool {
	return tree.label[0] == byte('*')
}

// isParameter checks if the node is marked with
// a parameter character at the start of the string.
func (tree node) isParameter() bool {
	return tree.label[0] == byte(':')
}

func (tree node) prefix() string {
	if tree.parent == nil {
		return ""
	}

	return tree.parent.path()
}

func (tree node) path() string {
	return tree.prefix() + tree.label.String()
}

// String conforms to the fmt.Stringer interface,
// so we can easily print out internal tree structure.
func (tree node) String() string {
	if tree.label == "" && tree.handler == nil {
		return "NIL TREE"
	}

	return tree.string(0)
}

// string is used to properly indent the tree structure.
func (tree node) string(offset int) string {
	buff := &bytes.Buffer{}
	if offset > 1 {
		fmt.Fprint(buff, strings.Repeat("   ", int(math.Max(float64(0), float64(offset-1)))))
	}
	if offset > 0 {
		fmt.Fprintf(buff, "└── ")
	}

	fmt.Fprintf(buff, "%s", tree.label)

	if tree.handler != nil {
		fmt.Fprint(buff, " ✓")
	}

	fmt.Fprintln(buff)

	for _, child := range tree.children {
		fmt.Fprintf(buff, child.string(offset+1))
	}
	return string(buff.Bytes())
}

func min(a, b int) int {
	if a < b {
		return a
	}

	return b
}

// nodeLabel is a string alias specialized for manipulating the
// labels of tree nodes.
type nodeLabel string

// String method satisfies the stringer interface.
func (label nodeLabel) String() string {
	return string(label)
}

// findPrefixLength returns the size of the longest common prefix.
func (label nodeLabel) findPrefixLength(newLabel nodeLabel) int {
	i := 0
	max := min(len(label), len(newLabel))
	for i < max && label[i] == newLabel[i] {
		i++
	}

	return i
}

// isValidRootLabel checks if the label can be used as a root node.
func (label nodeLabel) isValidRootLabel() bool {
	return len(label) > 0 && label[0] == byte('/')
}

// getWildcard returns an index of the wildcard and a boolean
// for signaling whether the wildcard was found.
func (label nodeLabel) getWildcard() (index int, ok bool) {
	index = strings.IndexByte(label.String(), byte('*'))
	if index == -1 {
		return 0, false
	}

	return index, true
}

// getParameter returns an index of the parameter and a boolean
// for signaling whether the parameter was found.
func (label nodeLabel) getParameter() (index int, ok bool) {
	index = strings.IndexByte(label.String(), byte(':'))
	if index == -1 {
		return 0, false
	}

	return index, true
}

// getVariable returns an index of the first parameter or wildcard
// encountered, and a boolean for signaling whether there are any
// variables in the label.
func (label nodeLabel) getVariable() (index int, ok bool) {
	index = strings.IndexAny(label.String(), ":*")
	if index == -1 {
		return 0, false
	}

	return index, true
}

// getEndOfVariable returns the index of the first '/' in a label,
// or, if not found, returns the index of the end of the label,
// and a ok=false indicating that the '/' was not found before
// the end of the label.
func (label nodeLabel) getEndOfVariable() (index int, ok bool) {
	index = strings.Index(label.String(), "/")
	if index == -1 {
		return len(label), false
	}

	return index, true
}
