// Copyright (c) 2014 Datacratic. All rights reserved.

package meter

import ()

type Handler interface {
	HandleMeters(map[string]float64)
}

type HandlerFunc func(map[string]float64)

func (fn HandlerFunc) HandleMeters(values map[string]float64) {
	fn(values)
}
