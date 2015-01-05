// Copyright (c) 2014 Datacratic. All rights reserved.

package meter

import (
	"math"
	"math/rand"
	"sort"
	"sync"
	"time"
)

// DefaultDistributionSize is used if Size is not set in Distribution.
const DefaultDistributionSize = 1000

// Distribution aggregates metrics over a distribution of values.
//
// Record will add up to a maximum of Size elements after which new elements
// will be randomly sampled with a probability that depends on the number of
// elements recorded. This schemes ensures that a distribution has a constant
// memory footprint and doesn't need to allocate for calls to Record.
//
// ReadMeter will compute percentiles over the sampled distribution and the min
// and max value seen over the entire distribution.
//
// Distribution is completely go-routine safe.
type Distribution struct {

	// Size is maximum number of elements that the distribution can hold. Above
	// this amount, new values are sampled.
	Size int

	// SamplingSeed is the initial seed for the RNG used during sampling.
	SamplingSeed int64

	mutex sync.Mutex
	state *distribution
}

// Record adds the given value to the distribution with a probability based on
// the number of elements recorded since the last call to ReadMeter.
func (dist *Distribution) Record(value float64) {
	dist.mutex.Lock()

	if dist.state == nil {
		dist.state = newDistribution(dist.getSize(), dist.getSeed())
	}
	dist.state.Record(value)

	dist.mutex.Unlock()
}

// RecordDuration similar to Record but with time.Duration values.
func (dist *Distribution) RecordDuration(duration time.Duration) {
	dist.Record(float64(duration))
}

// ReadMeter computes various statistic over the sampled distribution (50th,
// 90th and 99th percentile) and the count, min and max over the entire
// distribution. All recorded elements are then discarded from the distribution.
func (dist *Distribution) ReadMeter(delta time.Duration) map[string]float64 {
	dist.mutex.Lock()

	oldState := dist.state
	dist.state = newDistribution(dist.getSize(), dist.getSeed())

	dist.mutex.Unlock()

	if oldState == nil {
		return make(map[string]float64)
	}

	return oldState.ReadMeter(delta)
}

func (dist *Distribution) getSize() int {
	if dist.Size == 0 {
		return DefaultDistributionSize
	}
	return dist.Size
}

func (dist *Distribution) getSeed() int64 {
	dist.SamplingSeed++
	return dist.SamplingSeed
}

type distribution struct {
	items    []float64
	count    int
	min, max float64

	rand *rand.Rand
}

func newDistribution(size int, seed int64) *distribution {
	return &distribution{
		items: make([]float64, size),
		min:   math.MaxFloat64,

		rand: rand.New(rand.NewSource(seed)),
	}
}

func (dist *distribution) Record(value float64) {
	dist.count++

	if dist.count <= len(dist.items) {
		dist.items[dist.count-1] = value

	} else if i := dist.rand.Int63n(int64(dist.count)); int(i) < len(dist.items) {
		dist.items[i] = value
	}

	if value < dist.min {
		dist.min = value
	}

	if value > dist.max {
		dist.max = value
	}
}

type float64Array []float64

func (array float64Array) Len() int           { return len(array) }
func (array float64Array) Swap(i, j int)      { array[i], array[j] = array[j], array[i] }
func (array float64Array) Less(i, j int) bool { return array[i] < array[j] }

func (dist *distribution) ReadMeter(_ time.Duration) map[string]float64 {
	if dist.count == 0 {
		return map[string]float64{}
	}

	items := make([]float64, len(dist.items))
	for i := 0; i < len(dist.items); i++ {
		items[i] = dist.items[i]
	}

	n := dist.count
	if dist.count > len(items) {
		n = len(items)
	}

	sort.Sort(float64Array(items[:n]))

	percentile := func(p int) float64 {
		index := float32(n) / 100 * float32(p)
		return items[int(index)]
	}

	result := map[string]float64{
		"count": float64(dist.count),
		"p00":   dist.min,
		"p50":   percentile(50),
		"p90":   percentile(90),
		"p99":   percentile(99),
		"pmx":   dist.max,
	}

	dist.count = 0
	dist.min = 0
	dist.max = 0

	return result
}
