// Copyright (c) 2014 Datacratic. All rights reserved.

package meter

import (
	"sync/atomic"
	"time"
)

type Counter struct{ value uint64 }

func (counter *Counter) RecordHit() {
	atomic.AddUint64(&counter.value, 1)
}

func (counter *Counter) RecordCount(count uint64) {
	atomic.AddUint64(&counter.value, count)
}

func (counter *Counter) ReadMeter(delta time.Duration) map[string]float64 {
	if value := atomic.SwapUint64(&counter.value, 0); value > 0 {
		normalized := float64(value) * (float64(time.Second) / float64(delta))
		return map[string]float64{"": normalized}
	}

	return map[string]float64{}
}
