package hyper

import (
	"context"
	"net/http"
	"testing"
)

func testHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func TestTreeAddAndGet(t *testing.T) {
	tree := treeNode{}

	routes := [...]string{
		"/api/v1/test",
		"/api/v1/foo/bar",
		"/api/v1/example",
		"/api/v1/foo",
	}

	for _, route := range routes {
		tree.insertNode(route, HandlerFunc(testHandler))
	}

	tests := []struct {
		route             string
		shouldHaveHandler bool
	}{
		{"/api/v1/test", true},
		{"/api/v1/example", true},
		{"/hello/world", false},
	}

	for _, test := range tests {
		handler := tree.getHandler(test.route)

		hasHandler := handler != nil

		if hasHandler != test.shouldHaveHandler {
			t.Errorf("tree.getHandler(%s): %v, wanted %v", test.route, hasHandler, test.shouldHaveHandler)
		}
	}
}

func TestFindPrefixSize(t *testing.T) {
	tests := []struct {
		first, second string
		wanted        int
	}{
		{"foo bar", "foo bar baz", 7},
		{"foo", "Foo", 0},
		{"foobar", "foo bar", 3},
	}

	for _, test := range tests {
		got := findPrefixLength(test.first, test.second)

		if got != test.wanted {
			t.Errorf(`findPrefix("%s", "%s") got "%s", wanted "%s"`, test.first, test.second, got, test.wanted)
		}
	}
}
