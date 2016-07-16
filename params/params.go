package params

import "context"

const ctxParams = "hyper.params"

// Param is a single URL parameter, consisting of a key and a value.
type Param struct {
	Key   string
	Value string
}

// Params is a Param-slice, as returned by the router.
// The slice is ordered, the first URL parameter is also the first slice value.
// It is therefore safe to read values by the index.
type Params []Param

// ByName returns the value of the first Param which key matches the given name.
// If no matching Param is found, ok = false is returned
func (ps Params) ByName(name string) (string, bool) {
	for i := range ps {
		if ps[i].Key == name {
			return ps[i].Value, true
		}
	}
	return "", false
}

// NewContext returns a new context.Context with Params object passed as value.
func NewContext(ctx context.Context, params Params) context.Context {
	return context.WithValue(ctx, ctxParams, params)
}

// Extracts params from a given context.
func FromContext(ctx context.Context) (Params, bool) {
	p, ok := ctx.Value(ctxParams).(Params)

	return p, ok
}
