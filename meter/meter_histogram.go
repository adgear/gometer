// Copyright (c) 2014 Datacratic. All rights reserved.

package meter

import (
	"math"
	"math/rand"
	"sort"
	"sync"
	"time"
)

// DefaultHistogramSize is used if Size is not set in Histogram.
const DefaultHistogramSize = 1000

// Histogram aggregates metrics over a histogram of values.
//
// Record will add up to a maximum of Size elements after which new elements
// will be randomly sampled with a probability that depends on the number of
// elements recorded. This schemes ensures that a histogram has a constant
// memory footprint and doesn't need to allocate for calls to Record.
//
// ReadMeter will compute percentiles over the sampled histogram and the min
// and max value seen over the entire histogram.
//
// Histogram is completely go-routine safe.
type Histogram struct {

	// Size is maximum number of elements that the histogram can hold. Above
	// this amount, new values are sampled.
	Size int

	// SamplingSeed is the initial seed for the RNG used during sampling.
	SamplingSeed int64

	mutex sync.Mutex
	state *histogram
}

// Record adds the given value to the histogram with a probability based on
// the number of elements recorded since the last call to ReadMeter.
func (dist *Histogram) Record(value float64) {
	dist.mutex.Lock()

	if dist.state == nil {
		dist.state = newHistogram(dist.getSize(), dist.getSeed())
	}
	dist.state.Record(value)

	dist.mutex.Unlock()
}

// RecordDuration similar to Record but with time.Duration values.
func (dist *Histogram) RecordDuration(duration time.Duration) {
	dist.Record(float64(duration) / float64(time.Second))
}

// RecordSince records a duration elapsed since the given time.
func (dist *Histogram) RecordSince(t0 time.Time) {
	dist.RecordDuration(time.Since(t0))
}

// ReadMeter computes various statistic over the sampled histogram (50th,
// 90th and 99th percentile) and the count, min and max over the entire
// histogram. All recorded elements are then discarded from the histogram.
func (dist *Histogram) ReadMeter(_ time.Duration) map[string]float64 {
	dist.mutex.Lock()

	oldState := dist.state
	dist.state = newHistogram(dist.getSize(), dist.getSeed())

	dist.mutex.Unlock()

	if oldState == nil {
		return make(map[string]float64)
	}

	return oldState.Read()
}

func (dist *Histogram) getSize() int {
	if dist.Size == 0 {
		return DefaultHistogramSize
	}
	return dist.Size
}

func (dist *Histogram) getSeed() int64 {
	dist.SamplingSeed++
	return dist.SamplingSeed
}

type histogram struct {
	items    []float64
	count    int
	min, max float64
	sum      float64

	rand *rand.Rand
}

func newHistogram(size int, seed int64) *histogram {
	return &histogram{
		items: make([]float64, size),
		min:   math.MaxFloat64,

		rand: rand.New(rand.NewSource(seed)),
	}
}

func (dist *histogram) Record(value float64) {
	dist.count++
	dist.sum += value

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

func (dist *histogram) Read() map[string]float64 {
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

	return map[string]float64{
		"count": float64(dist.count),
		"min":   dist.min,
		"max":   dist.max,
		"avg":   dist.sum / float64(dist.count),
		"p50":   percentile(50),
		"p90":   percentile(90),
		"p99":   percentile(99),
	}
}

// GetHistogram returns the histogram registered with the given key or creates a
// new one and registers it.
func GetHistogram(prefix string) *Histogram {
	return GetOrAdd(prefix, new(Histogram)).(*Histogram)
}
