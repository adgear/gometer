// Copyright (c) 2014 Datacratic. All rights reserved.

package meter

import (
	"fmt"
	"sync"
	"time"
)

type MetricType int

const (
	CountMetric MetricType = iota
	DistMetric
	LevelMetric
)

func (typ MetricType) String() string {
	switch typ {

	case CountMetric:
		return "counter"

	case DistMetric:
		return "distribution"

	case LevelMetric:
		return "level"

	default:
		return "???"

	}
}

const DefaultQueueSize = 128

type Metric struct {
	Type  MetricType
	Key   string
	Value float64
}

type Queue struct{ metrics []Metric }

func (queue *Queue) RecordHit(key string) {
	queue.add(Metric{CountMetric, key, 1.0})
}

func (queue *Queue) RecordCount(key string, count float64) {
	queue.add(Metric{CountMetric, key, count})
}

func (queue *Queue) RecordValue(key string, value float64) {
	queue.add(Metric{DistMetric, key, value})
}

func (queue *Queue) RecordLevel(key string, level float64) {
	queue.add(Metric{LevelMetric, key, level})
}

func (queue *Queue) add(metric Metric) {
	queue.metrics = append(queue.metrics, metric)
}

type Queued struct {
	mutex  sync.Mutex
	meters map[string]Meter
	seed   int64
}

func (queued *Queued) New() Queue {
	return Queue{make([]Metric, DefaultQueueSize)[:0]}
}

func (queued *Queued) Record(queue Queue) error {
	queued.mutex.Lock()

	var errors []error
	errorf := func(format string, args ...interface{}) {
		errors = append(errors, fmt.Errorf(format, args...))
	}

	for _, metric := range queue.metrics {

		meter, ok := queued.meters[metric.Key]

		switch metric.Type {

		case CountMetric:
			if !ok {
				meter = new(counter)
			}

			if counter, ok := meter.(*counter); ok {
				counter.value += metric.Value
			} else {
				errorf("incompatible meter type '%s' != '%T' for key '%s'", metric.Type, meter, metric.Key)
			}

		case LevelMetric:
			if !ok {
				meter = new(level)
			}

			if level, ok := meter.(*level); ok {
				level.value = metric.Value
			} else {
				errorf("incompatible meter type '%s' != '%T' for key '%s'", metric.Type, meter, metric.Key)
			}

		case DistMetric:
			if !ok {
				queued.seed++
				meter = newDistribution(DefaultDistributionSize, queued.seed)
			}

			if dist, ok := meter.(*distribution); ok {
				dist.Record(metric.Value)

			} else {
				errorf("incompatible meter type '%s' != '%T' for key '%s'", metric.Type, meter, metric.Key)
			}

		default:
			errors = append(errors, fmt.Errorf("unknown metric type: %s", metric.Type))
			continue
		}

		if !ok {
			queued.meters[metric.Key] = meter
		}

	}

	queued.mutex.Unlock()

	return nil
}

func (queued *Queued) ReadMeter(delta time.Duration) map[string]float64 {
	result := make(map[string]float64)

	queued.mutex.Lock()

	for prefix, meter := range queued.meters {
		for suffix, value := range meter.ReadMeter(delta) {
			result[Join(prefix, suffix)] = value
		}
	}

	queued.mutex.Unlock()

	return result
}

type counter struct{ value float64 }

func (counter *counter) ReadMeter(delta time.Duration) map[string]float64 {
	result := make(map[string]float64)

	if counter.value > 0 {
		result[""] = float64(counter.value) * (float64(time.Second) / float64(delta))
		counter.value = 0
	}

	return result
}

type level struct{ value float64 }

func (level *level) ReadMeter(delta time.Duration) map[string]float64 {
	result := make(map[string]float64)

	if level.value > 0 {
		result[""] = level.value
	}

	return result
}
