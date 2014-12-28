// Copyright (c) 2014 Datacratic. All rights reserved.

package meter

import (
	"fmt"
	"testing"
	"time"
)

type Metrics struct {
	Bool     bool
	Int      int
	Float    float64
	String   string
	Duration time.Duration
	Map      map[string]string
}

func TestStruct(t *testing.T) {
	var meter Struct

	meter.Record(&Metrics{
		Bool:     true,
		Int:      12,
		Float:    10,
		String:   "foo",
		Duration: 1 * time.Second,
		Map:      map[string]string{"a": "b", "b": "c", "c": "d"},
	})

	meter.Record(&Metrics{
		Int:      21,
		Float:    10,
		String:   "foo",
		Duration: 1 * time.Second,
		Map:      map[string]string{"a": "c", "b": "c", "d": "e"},
	})

	CheckValues(t, "struct", meter.ReadMeter(time.Second), map[string]float64{
		"Metrics.Bool":           1,
		"Metrics.Int":            33,
		"Metrics.Float.count":    2,
		"Metrics.Float.p00":      10,
		"Metrics.Float.p50":      10,
		"Metrics.Float.p90":      10,
		"Metrics.Float.p99":      10,
		"Metrics.Float.pmx":      10,
		"Metrics.String.foo":     2,
		"Metrics.Duration.count": 2,
		"Metrics.Duration.p00":   float64(1 * time.Second),
		"Metrics.Duration.p50":   float64(1 * time.Second),
		"Metrics.Duration.p90":   float64(1 * time.Second),
		"Metrics.Duration.p99":   float64(1 * time.Second),
		"Metrics.Duration.pmx":   float64(1 * time.Second),
		"Metrics.Map.a.b":        1,
		"Metrics.Map.a.c":        1,
		"Metrics.Map.b.c":        2,
		"Metrics.Map.c.d":        1,
		"Metrics.Map.d.e":        1,
	})
}

func BenchStruct(b *testing.B, obj interface{}) {
	var meter Struct

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		meter.RecordPrefix("", obj)
	}
}

func BenchmarkStruct_Bool(b *testing.B)     { BenchStruct(b, true) }
func BenchmarkStruct_Int(b *testing.B)      { BenchStruct(b, 10) }
func BenchmarkStruct_Float64(b *testing.B)  { BenchStruct(b, 1.2) }
func BenchmarkStruct_String(b *testing.B)   { BenchStruct(b, "foo") }
func BenchmarkStruct_Duration(b *testing.B) { BenchStruct(b, time.Second) }

type S1 struct{ K0 bool }
type S2 struct{ K0, K1 bool }
type S4 struct{ K0, K1, K2, K3 bool }
type S8 struct{ K0, K1, K2, K3, K4, K5, K6, K7 bool }

func BenchmarkStruct_1Struct(b *testing.B) { BenchStruct(b, new(S1)) }
func BenchmarkStruct_2Struct(b *testing.B) { BenchStruct(b, new(S2)) }
func BenchmarkStruct_4Struct(b *testing.B) { BenchStruct(b, new(S4)) }
func BenchmarkStruct_8Struct(b *testing.B) { BenchStruct(b, new(S8)) }

func BenchStructMap(b *testing.B, n int) {
	var meter Struct

	value := make(map[string]bool)
	for i := 0; i < n; i++ {
		value[fmt.Sprintf("k%d", i)] = true
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		meter.RecordPrefix("", value)
	}
}

func BenchmarkStruct_1Map(b *testing.B)   { BenchStructMap(b, 1) }
func BenchmarkStruct_10Map(b *testing.B)  { BenchStructMap(b, 10) }
func BenchmarkStruct_100Map(b *testing.B) { BenchStructMap(b, 100) }
