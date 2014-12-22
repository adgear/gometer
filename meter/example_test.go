// Copyright (c) 2014 Datacratic. All rights reserved.

package meter_test

import (
	"github.com/datacratic/gometer/meter"

	"fmt"
	"sort"
)

func ExampleMeter() {

	var counter meter.Counter
	var dist meter.Distribution
	var level meter.Level
	var multi meter.MultiCounter

	meter.Add("meters.hits", &counter)
	meter.Add("meters.latency", &dist)
	meter.Add("meters.level", &level)
	meter.Add("meters.result", &multi)

	resultC := make(chan map[string]float64)
	handler := func(values map[string]float64) { resultC <- values }
	meter.Handle(meter.HandlerFunc(handler))

	counter.RecordHit()
	counter.RecordCount(10)

	for i := 0; i < 100; i++ {
		dist.Record(float64(i))
	}

	level.Record(5)

	multi.RecordHit("err")
	multi.RecordCount("ok", 10)

	SortAndPrint(<-resultC)

	// Output:
	// meters.hits: 11.000000
	// meters.latency.count: 100.000000
	// meters.latency.p00: 0.000000
	// meters.latency.p50: 50.000000
	// meters.latency.p90: 90.000000
	// meters.latency.p99: 99.000000
	// meters.latency.pmx: 99.000000
	// meters.level: 5.000000
	// meters.result.err: 1.000000
	// meters.result.ok: 10.000000
}

func SortAndPrint(values map[string]float64) {
	var keys []string

	for key := range values {
		keys = append(keys, key)
	}

	sort.Strings(keys)

	for _, key := range keys {
		fmt.Printf("%s: %f\n", key, values[key])
	}

}
