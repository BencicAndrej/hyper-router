package hyper

import (
	"context"
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

type tree map[string]http.Handler

// Router is a http.Handler that is responsible for
// registering and dispatching other handlers to correct routes.
type Router struct {
	routes map[string]http.Handler
}
