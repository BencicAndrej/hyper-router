package hyper

import "net/http"

// Middleware is a function that takes a handler,
// is expected to do some work before or after provided
// handler and returns its handler.
type Middleware func(next http.Handler) http.Handler

// MiddlewareStack is a list of Middleware.
type MiddlewareStack struct {
	middleware []Middleware
}

// NewStack creates a MiddlewareStack with arbitrary
// number or middleware.
func NewStack(middleware ...Middleware) MiddlewareStack {
	return MiddlewareStack{
		middleware: append([]Middleware{}, middleware...),
	}
}

// Do adds the last Handler and returns the final,
// net/http compatible http.Handler.
func (stack MiddlewareStack) Do(h http.Handler) http.Handler {
	// Loop through all middleware backwards and construct
	// the final middleware.
	for i := range stack.middleware {
		h = stack.middleware[len(stack.middleware)-1-i](h)
	}

	return h
}

// A shorthand of stack.Do(stack.HandlerFunc(f))
func (stack MiddlewareStack) DoFunc(f http.HandlerFunc) http.Handler {
	return stack.Do(http.HandlerFunc(f))
}

// Append extends a stack, adding the specified middleware
// as the last ones in the request flow.
//
// Append returns a new stack, leaving the original one untouched.
func (stack MiddlewareStack) Append(middleware ...Middleware) MiddlewareStack {
	newMiddleware := make([]Middleware, len(stack.middleware)+len(middleware))
	copy(newMiddleware, stack.middleware)
	copy(newMiddleware[len(stack.middleware):], middleware)

	return NewStack(newMiddleware...)
}

// Extend extends a stack by adding the specified stack
// as the last one in the request flow.
//
// Extend returns a new stack, leaving the original one untouched.
func (stack MiddlewareStack) Extend(newStack MiddlewareStack) MiddlewareStack {
	return stack.Append(newStack.middleware...)
}
