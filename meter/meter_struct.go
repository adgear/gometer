// Copyright (c) 2014 Datacratic. All rights reserved.

package meter

import (
	"github.com/datacratic/goklog/klog"

	"reflect"
	"sync"
	"time"
)

var structMeter Struct
var initializeStructMeter sync.Once

func initStructMeter() {
	initializeStructMeter.Do(func() {
		Add("structs", &structMeter)
	})
}

// RecordMetrics uses the given object to record various metrics based on the
// type of the fields. This function uses the type name of the given object as a
// prefix to all the keys.
func RecordMetrics(obj interface{}) {
	initStructMeter()
	structMeter.Record(obj)
}

// RecordMetricsPrefix same as record except that prefix is used to prefix all the keys
// instead of the type of the object. This is mostly a compatibility option for
// older code.
func RecordMetricsPrefix(prefix string, obj interface{}) {
	initStructMeter()
	structMeter.RecordPrefix(prefix, obj)
}

// Struct uses a struct to aggregate a various metrics. Acts as a backwards
// compatibility layer to gometrics and is not recommended for new code.
type Struct struct {
	mutex  sync.Mutex
	values map[string]Meter
}

// ReadMeter returns the current aggregate metrics.
func (meter *Struct) ReadMeter(delta time.Duration) map[string]float64 {
	result := make(map[string]float64)

	meter.mutex.Lock()

	if meter.values == nil {
		meter.values = make(map[string]Meter)
	}

	for prefix, m := range meter.values {
		for suffix, value := range m.ReadMeter(delta) {
			result[Join(prefix, suffix)] = value
		}
	}

	meter.mutex.Unlock()

	return result
}

// Record uses the given object to record various metrics based on the type of
// the fields. This function uses the type name of the given object as a prefix
// to all the keys.
func (meter *Struct) Record(obj interface{}) {
	meter.mutex.Lock()

	if meter.values == nil {
		meter.values = make(map[string]Meter)
	}

	value := reflect.ValueOf(obj)
	meter.record(meter.name(value), value)

	meter.mutex.Unlock()
}

func (meter *Struct) name(obj reflect.Value) string {
	if obj.Type().Kind() == reflect.Ptr {
		return meter.name(obj.Elem())
	}

	return obj.Type().Name()
}

// RecordPrefix same as record except that prefix is used to prefix all the keys
// instead of the type of the object. This is mostly a compatibility option for
// older code.
func (meter *Struct) RecordPrefix(prefix string, obj interface{}) {
	meter.mutex.Lock()

	if meter.values == nil {
		meter.values = make(map[string]Meter)
	}

	meter.record(prefix, reflect.ValueOf(obj))

	meter.mutex.Unlock()
}

func (meter *Struct) record(prefix string, obj reflect.Value) {
	typ := obj.Type()

	switch typ.Kind() {

	case reflect.Ptr:
		meter.record(prefix, obj.Elem())

	case reflect.Map:
		for _, k := range obj.MapKeys() {
			key := k.Interface().(string)
			meter.record(Join(prefix, key), obj.MapIndex(k))
		}

	case reflect.Struct:
		for index := 0; index < typ.NumField(); index++ {
			field := typ.Field(index)
			meter.record(Join(prefix, field.Name), obj.Field(index))
		}

	default:
		i := obj.Interface()

		switch i.(type) {

		case bool:
			meter.recordBool(prefix, i.(bool))

		case int:
			meter.recordInt(prefix, i.(int))

		case float64:
			meter.recordFloat(prefix, i.(float64))

		case string:
			meter.recordString(prefix, i.(string))

		case time.Duration:
			meter.recordDuration(prefix, i.(time.Duration))

		default:
			klog.KFatalf("meter.struct.error", "unknown kind '%s' for key '%s'", typ.Kind(), prefix)
		}
	}
}

func (meter *Struct) recordBool(key string, value bool) {
	if !value {
		return
	}

	counter, ok := meter.values[key]

	if !ok {
		counter = new(Counter)
		meter.values[key] = counter
	}

	counter.(*Counter).RecordHit()
}

func (meter *Struct) recordInt(key string, value int) {
	if value == 0 {
		return
	}

	counter, ok := meter.values[key]

	if !ok {
		counter = new(Counter)
		meter.values[key] = counter
	}

	counter.(*Counter).RecordCount(uint64(value))
}

func (meter *Struct) recordString(key string, value string) {
	if value == "" {
		return
	}

	multi, ok := meter.values[key]

	if !ok {
		multi = new(MultiCounter)
		meter.values[key] = multi
	}

	multi.(*MultiCounter).RecordHit(value)
}

func (meter *Struct) recordFloat(key string, value float64) {
	dist, ok := meter.values[key]

	if !ok {
		dist = new(Distribution)
		meter.values[key] = dist
	}

	dist.(*Distribution).Record(value)
}

func (meter *Struct) recordDuration(key string, value time.Duration) {
	dist, ok := meter.values[key]

	if !ok {
		dist = new(Distribution)
		meter.values[key] = dist
	}

	dist.(*Distribution).RecordDuration(value)
}
