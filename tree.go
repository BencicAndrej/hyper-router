package hyper

import (
    "bytes"
    "fmt"
    "math"
    "strings"
)

type node struct {
    label    string
    handler  Handler

    parent   *node
    // children represents an array of child nodes, ordered by priority:
    // wildcard then parameters then static.
    children []*node
}

func (tree node) getHandler(label string) Handler {
    return nil
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
func (tree *node) insert(label string, handler Handler) *node {
    // #1) If tree is empty populate the current element.
    if tree.label == "" && tree.handler == nil {
        // Label must start with a '/'.
        if len(label) == 0 || label[0] != byte('/') {
            panic(fmt.Sprintf("route '%s' must start with '/'", tree.prefix() + label))
        }

        if variablePos := strings.IndexAny(label, ":*"); variablePos != -1 {
            // Will never be 0 because a '/' is required to be first.
            tree.label = label[:variablePos]

            return tree.insert(label[variablePos:], handler)
        }

        tree.label = label
        tree.handler = handler
        return tree
    }

    if parameterPos := strings.Index(label, ":"); parameterPos != -1 {
        if parameterPos == 0 {
            // Find end of parameter
            parameterEnd := strings.Index(label, "/")
            if parameterEnd == -1 {
                parameterEnd = len(label)
            }

            if !(len(tree.children) == 1 && tree.children[0].label == label[:parameterEnd]) || len(tree.children) != 0 {
                panic(fmt.Sprintf("handler for route '%s' already exists", tree.path() + label))
            }

            newNode := node{
                label: label[:parameterEnd],

                parent: tree,
            }

            // Insert new node at the start of the children nodes.
            tree.children = append([]*node{&newNode}, tree.children...)

            if parameterEnd == len(label) {
                newNode.handler = handler
                return &newNode
            }

            return newNode.insert(label[parameterEnd:], handler)
        }

        newNode := tree.insert(label[:parameterPos], nil)

        return newNode.insert(label[parameterPos:], handler)
    }

    if wildcardPos := strings.Index(label, "*"); wildcardPos != -1 {
        if wildcardPos == 0 {
            if wildCardEnd := strings.Index(label, "/"); wildCardEnd != -1 {
                panic(
                    fmt.Sprintf("wildcard parameter must be the last element of the route '%s'", tree.prefix() + label),
                )
            }

            if len(tree.children) > 0 {
                panic(fmt.Sprintf("handler for route '%s' already exists", tree.prefix() + label))
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
    prefixLength := findPrefixLength(tree.label, label)

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
        if child.isWildcard() || child.isParameter() || child.label[0] == label[prefixLength] {
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

func (tree node) isWildcard() bool {
    return tree.label[0] == byte('*')
}

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
    return tree.prefix() + tree.label
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
        fmt.Fprint(buff, strings.Repeat("   ", int(math.Max(float64(0), float64(offset - 1)))))
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
        fmt.Fprintf(buff, child.string(offset + 1))
    }
    return string(buff.Bytes())
}

// findPrefixLength returns the size of the longest common prefix.
func findPrefixLength(a, b string) int {
    i := 0
    max := min(len(a), len(b))
    for i < max && a[i] == b[i] {
        i++
    }

    return i
}

func min(a, b int) int {
    if a < b {
        return a
    }

    return b
}