package hyper

import (
	"fmt"
	"net/http"
)

// Router is a http.Handler that is responsible for
// registering and dispatching other handlers to correct routes.
type Router struct {
	handlerTrees map[string]*node
}

// NewRouter return the an empty Router.
func NewRouter() *Router {
	return &Router{}
}

// Get is a shortcut to the router.Handle(http.MethodGet, path, handler) method.
func (r *Router) Get(path string, handler http.Handler) {
	r.Handle(http.MethodGet, path, handler)
}

// Get is a shortcut to the router.Handle(http.MethodHead, path, handler) method.
func (r *Router) Head(path string, handler http.Handler) {
	r.Handle(http.MethodHead, path, handler)
}

// Get is a shortcut to the router.Handle(http.MethodOptions, path, handler) method.
func (r *Router) Options(path string, handler http.Handler) {
	r.Handle(http.MethodOptions, path, handler)
}

// Get is a shortcut to the router.Handle(http.MethodPost, path, handler) method.
func (r *Router) Post(path string, handler http.Handler) {
	r.Handle(http.MethodPost, path, handler)
}

// Get is a shortcut to the router.Handle(http.MethodPut, path, handler) method.
func (r *Router) Put(path string, handler http.Handler) {
	r.Handle(http.MethodPut, path, handler)
}

// Get is a shortcut to the router.Handle(http.MethodPatch, path, handler) method.
func (r *Router) Patch(path string, handler http.Handler) {
	r.Handle(http.MethodPatch, path, handler)
}

// Get is a shortcut to the router.Handle(http.MethodDelete, path, handler) method.
func (r *Router) Delete(path string, handler http.Handler) {
	r.Handle(http.MethodDelete, path, handler)
}

// Handle adds a new route to the Router for the specified method and path.
func (r *Router) Handle(method string, path string, handler http.Handler) {
	if path[0] != '/' {
		panic(fmt.Sprintf("path must start with '/' in '%s'", path))
	}

	// If no routes are defined yet, create a new tree.
	if r.handlerTrees == nil {
		r.handlerTrees = make(map[string]*node)
	}

	// If tree does not exist for `method`, create a new one.
	root, ok := r.handlerTrees[method]
	if !ok {
		root = new(node)
		r.handlerTrees[method] = root
	}

	root.insert(nodeLabel(path), handler)
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	path := req.URL.Path

	if root, ok := r.handlerTrees[req.Method]; ok {
		if handler, ctx := root.getHandler(req.Context(), nodeLabel(path)); handler != nil {
			handler.ServeHTTP(w, req.WithContext(ctx))
			return
		}
	}

	http.NotFound(w, req)
}
