// Copyright (c) 2014 Datacratic. All rights reserved.

package meter

import (
	"reflect"
)

var (
	counterType      = reflect.TypeOf((*Counter)(nil))
	counterMultiType = reflect.TypeOf((*MultiCounter)(nil))

	gaugeType      = reflect.TypeOf((*Gauge)(nil))
	gaugeMultiType = reflect.TypeOf((*MultiGauge)(nil))

	histogramType      = reflect.TypeOf((*Histogram)(nil))
	histogramMultiType = reflect.TypeOf((*MultiHistogram)(nil))

	stateType = reflect.TypeOf((*State)(nil))
)

// Load crawls the given object to register and instantiate any pointer to
// meters that it finds. The meter key will be derived from the name of the
// field that contains the meter and will be prefixed by the given prefix
// string. If a structure is encoutered then it will be recursively crawled with
// the name of the field appended to the given prefix.
func Load(obj interface{}, prefix string) {

	forEachMeter(reflect.ValueOf(obj), prefix, func(field reflect.Value, name string) {
		switch field.Type() {

		case counterType:
			field.Set(reflect.ValueOf(GetCounter(name)))
		case counterMultiType:
			field.Set(reflect.ValueOf(GetMultiCounter(name)))

		case gaugeType:
			field.Set(reflect.ValueOf(GetGauge(name)))
		case gaugeMultiType:
			field.Set(reflect.ValueOf(GetMultiGauge(name)))

		case histogramType:
			field.Set(reflect.ValueOf(GetHistogram(name)))
		case histogramMultiType:
			field.Set(reflect.ValueOf(GetMultiHistogram(name)))

		case stateType:
			field.Set(reflect.ValueOf(GetState(name)))
		}
	})
}

// Unload crawls the given object and deregisters any pointer to meters that it
// finds. See Load for more details about the crawling and naming behaviour.
func Unload(obj interface{}, prefix string) {
	forEachMeter(reflect.ValueOf(obj), prefix, func(field reflect.Value, name string) {
		switch field.Type() {

		case counterType:
		case counterMultiType:

		case gaugeType:
		case gaugeMultiType:

		case histogramType:
		case histogramMultiType:

		case stateType:

		default:
			return
		}

		Remove(name)
	})
}

func forEachMeter(value reflect.Value, prefix string, fn func(reflect.Value, string)) {
	if value.Kind() == reflect.Ptr {
		value = value.Elem()
	}

	typ := value.Type()

	for i := 0; i < typ.NumField(); i++ {
		field := value.Field(i)
		fieldEntry := typ.Field(i)

		name := Join(prefix, fieldEntry.Name)

		if field.Kind() == reflect.Struct {
			forEachMeter(field, name, fn)
		} else {
			fn(field, name)
		}
	}
}
