package hyper

import (
	"golang.org/x/net/context"
	"net/http"
	"testing"
    "fmt"
)

var emptyHandler = HandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {})

func TestEmptyNodeTree(t *testing.T) {
	tree := &node{}

	handler := tree.getHandler("/route/404")

	if handler != nil {
		t.Errorf("node.getHandler('%s'): got %v, wanted nil", "/route/404", handler)
	}
}

func TestInsertRoute(t *testing.T) {
	tests := []struct {
		routes []string
		want   *node
	}{
		{
			// Empty tree.
			routes: []string{},
			want: &node{
				label:    "",
				children: []*node{},
			},
		},
		{
			// Single route in tree.
			routes: []string{"/foo"},
			want:   leaf("/foo"),
		},
		{
			// Single route in tree.
			routes: []string{"/foo/bar"},
			want:   leaf("/foo/bar"),
		},
		{
			// Two different routes in tree.
			routes: []string{"/foo", "/bar"},
			want:   branch("/", leaf("foo"), leaf("bar")),
		},
		{
			// Three routes in tree.
			routes: []string{"/baz/foo", "/baz/bar", "/baz/foo/bar"},
			want: branch(
				"/baz/",
				leafBranch("foo", leaf("/bar")),
				leaf("bar"),
			),
		},
		{
			// Wild at the start
			routes: []string{"/*wild"},
			want:   branch("/", leaf("*wild")),
		},
	}

	for _, test := range tests {
		tree := loadTree(test.routes...)

		if !compareTrees(tree, test.want) {
			t.Errorf("node.insert(): got\n%s\n\nwanted:\n%s", tree, test.want)
		}
	}
}

func TestRouteInsertionFailures(t *testing.T) {
	tests := []struct {
		routes []string
		panic  string
	}{
		{
			routes: []string{"foo/bar"},
			panic:  "route 'foo/bar' must start with '/'",
		},
		{
			routes: []string{"/foo/bar", "/foo/bar"},
			panic:  "handler for route '/foo/bar' already exists",
		},
		{
			routes: []string{"/foo", "/:bar"},
			panic:  "handler for route '/:bar' already exists",
		},
		{
			routes: []string{"/:bar", "/foo"},
			panic:  "handler for route '/foo' already exists",
		},
		{
			routes: []string{"/foo/baz", "/:bar/baz"},
			panic:  "handler for route '/:bar/baz' already exists",
		},
        {
            routes: []string{"/foo/", "/foo/bar", "/foo/*baz"},
            panic:  "handler for route '/foo/baz already exists",
        },
        {
            routes: []string{"/foo/bar", "/foo/:baz"},
            panic:  "handler for route '/foo/:baz already exists",
        },
        {
            routes: []string{"/foo/*bar/baz"},
            panic:  "wildcard parameter must be the last element of the route '/foo/*bar/baz'",
        },
	}

	for _, test := range tests {
		func() {
			defer func() {
				r := recover()
                //@TODO: Fix the error handling
				//if r != test.panic {
				//	t.Errorf("node.insert(): got panic event \"%v\", wanted \"%s\"", r, test.panic)
				//}
				if r == nil {
                    t.Errorf(fmt.Sprintf("node.insert('%s'): expected panic, got none", test.routes))
                }
                //else {
                //    fmt.Printf("recieved panic event: %s\n", r)
                //}
			}()

			tree := loadTree(test.routes...)
            fmt.Print(tree)
		}()
	}
}

func TestNodeCanSplit(t *testing.T) {
	tests := []struct {
		tree *node
		want bool
	}{
		{
			tree: &node{
				label:   "/",
				handler: emptyHandler,
			},
			want: false,
		},
		{
			tree: &node{
				label:   "/foo",
				handler: emptyHandler,
			},
			want: true,
		},
	}

	for _, test := range tests {
		canSplit := test.tree.canSplit()

		if canSplit != test.want {
			t.Errorf("node.canSplit(): %v, wanted %v", canSplit, test.want)
		}
	}
}

func TestAddSingleRoute(t *testing.T) {
	tests := []struct {
        route string
        want  bool
	}{
		{"/api/v1/hello/world", true},
		{"/route/404", false},
	}

    tree := loadTree("/api/v1/hello/world")

	for _, test := range tests {
        handler := tree.getHandler(test.route)

        if (handler != nil) != test.want {
            t.Errorf("node.getHandler('%s'): %v, wanded %v", test.route, handler != nil, test.want)
        }
	}
}

func loadTree(routes ...string) *node {
	tree := &node{}

	for _, route := range routes {
		tree.insert(route, emptyHandler)
	}

	return tree
}

func compareTrees(a, b *node) bool {
	return a.String() == b.String()
}

func leaf(label string) *node {
	return &node{
		label:    label,
		handler:  emptyHandler,
		children: []*node{},
	}
}

func branch(label string, children ...*node) *node {
	return &node{
		label:    label,
		handler:  nil,
		children: children,
	}
}

func leafBranch(label string, children ...*node) *node {
	return &node{
		label:    label,
		handler:  emptyHandler,
		children: children,
	}
}
