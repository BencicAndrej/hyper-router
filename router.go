package hyper

import (
	"context"
	"fmt"
	"net/http"
)

// Handler interface is an extension of the http.Handler with an additional
// context provided with every request.
//
// The context object is used to propagate contextual information
// through the request.
type Handler interface {
	ServeHTTP(context.Context, http.ResponseWriter, *http.Request)
}

// HandlerFunc is an extension of the http.HandlerFunc with an additional
// context provided with every request
//
// The context object is used to propagate contextual information
// through the request.
type HandlerFunc func(context.Context, http.ResponseWriter, *http.Request)

// ServeHTTP method calls the parent HandleFunc in order to satisfy
// the Handler interface.
func (f HandlerFunc) ServeHTTP(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	f(ctx, w, r)
}

// Router is a http.Handler that is responsible for
// registering and dispatching other handlers to correct routes.
type Router struct {
	handlerTrees map[string]*treeNode
}

// NewRouter return the an empty Router.
func NewRouter() *Router {
	return &Router{}
}

// Get is a shortcut to the router.Handle(http.MethodGet, path, handler) method.
func (r *Router) Get(path string, handler Handler) {
	r.Handle(http.MethodGet, path, handler)
}

// Get is a shortcut to the router.Handle(http.MethodHead, path, handler) method.
func (r *Router) Head(path string, handler Handler) {
	r.Handle(http.MethodHead, path, handler)
}

// Get is a shortcut to the router.Handle(http.MethodOptions, path, handler) method.
func (r *Router) Options(path string, handler Handler) {
	r.Handle(http.MethodOptions, path, handler)
}

// Get is a shortcut to the router.Handle(http.MethodPost, path, handler) method.
func (r *Router) Post(path string, handler Handler) {
	r.Handle(http.MethodPost, path, handler)
}

// Get is a shortcut to the router.Handle(http.MethodPut, path, handler) method.
func (r *Router) Put(path string, handler Handler) {
	r.Handle(http.MethodPut, path, handler)
}

// Get is a shortcut to the router.Handle(http.MethodPatch, path, handler) method.
func (r *Router) Patch(path string, handler Handler) {
	r.Handle(http.MethodPatch, path, handler)
}

// Get is a shortcut to the router.Handle(http.MethodDelete, path, handler) method.
func (r *Router) Delete(path string, handler Handler) {
	r.Handle(http.MethodDelete, path, handler)
}

// Handle adds a new route to the Router for the specified method and path.
func (r *Router) Handle(method string, path string, handler Handler) {
	if path[0] != '/' {
		panic(fmt.Sprintf("path must start with '/' in '%s'", path))
	}

	// If no routes are defined yet, create a new tree.
	if r.handlerTrees == nil {
		r.handlerTrees = make(map[string]*treeNode)
	}

	// If tree does not exist for `method`, create a new one.
	root, ok := r.handlerTrees[method]
	if !ok {
		root = new(treeNode)
		r.handlerTrees[method] = root
	}

	root.insertNode(path, handler)
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	path := req.URL.Path

	if root, ok := r.handlerTrees[req.Method]; ok {
		if handler := root.getHandler(path); handler != nil {
			handler.ServeHTTP(context.Background(), w, req)
			return
		}
	}

	http.NotFound(w, req)
}
