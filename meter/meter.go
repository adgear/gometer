// Copyright (c) 2014 Datacratic. All rights reserved.

package meter

import (
	"time"
)

type Meter interface {
	ReadMeter(time.Duration) map[string]float64
}

var DefaultPoller Poller

func Add(key string, meter Meter) {
	DefaultPoller.Add(key, meter)
}

func Remove(key string) {
	DefaultPoller.Remove(key)
}

func Handle(handler Handler) {
	DefaultPoller.Handle(handler)
}
