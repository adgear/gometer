// Copyright (c) 2014 Datacratic. All rights reserved.

package meter

import ()

// Handler is used to periodically process the aggregated values of multiple
// meters.
type Handler interface {
	HandleMeters(map[string]float64)
}

// HandlerFunc is used to wrap a function as a Handler interface.
type HandlerFunc func(map[string]float64)

// HandleMeters forwards the call to the wrapped function.
func (fn HandlerFunc) HandleMeters(values map[string]float64) {
	fn(values)
}
