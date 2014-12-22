// Copyright (c) 2014 Datacratic. All rights reserved.
//
// All the multi counters follow the CounterMulti implementation so we'll write
// one test for them all here because counter is the lowest overhead meter which
// will reduce the noise in the benchmarks.

package meter

import (
	"strconv"
	"testing"
	"time"
)

func TestCounterMulti_Seq(t *testing.T) {
	var multi MultiCounter

	for i := 0; i < 3; i++ {
		multi.RecordHit("a")
		multi.RecordHit("b")
		multi.RecordHit("a")
		multi.RecordHit("a")
		multi.RecordHit("c")
		multi.RecordHit("b")
		multi.RecordHit("b")

		values := multi.ReadMeter(1 * time.Second)
		CheckValues(t, "seq", values, map[string]float64{"a": 3, "b": 3, "c": 1})
	}
}

func TestCounterMulti_Para(t *testing.T) {
	keys := 1000
	increments := 100
	workers := 10

	var multi MultiCounter

	stopReaderC := make(chan int)
	resultC := make(chan map[string]float64)

	go func() {
		tick := time.Tick(1 * time.Millisecond)
		result := make(map[string]float64)

		for {
			select {
			case <-tick:
				ReadValuesInto(t, &multi, result)

			case <-stopReaderC:
				ReadValuesInto(t, &multi, result)
				resultC <- result
				return
			}
		}
	}()

	workerDoneC := make(chan int)

	worker := func() {
		for i := 0; i < increments; i++ {
			for j := 0; j < keys; j++ {
				multi.RecordHit(strconv.Itoa(j))
			}
		}
		workerDoneC <- 1
	}

	for i := 0; i < workers; i++ {
		go worker()
	}

	for i := 0; i < workers; i++ {
		<-workerDoneC
	}

	stopReaderC <- 1
	result := <-resultC

	exp := make(map[string]float64)
	for i := 0; i < keys; i++ {
		exp[strconv.Itoa(i)] = float64(increments * workers)
	}
	CheckValues(t, "para", result, exp)
}

func ReadValuesInto(t *testing.T, multi *MultiCounter, values map[string]float64) {
	for key, value := range multi.ReadMeter(1 * time.Second) {
		values[key] += value
	}
}

func BenchmarkMultiCounter_ControlSeq(b *testing.B) {
	for i := 0; i < b.N; i++ {
		strconv.Itoa(i)
	}
}

func BenchMultiCounterSeq(b *testing.B, keys int) {
	var multi MultiCounter

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		multi.RecordHit(strconv.Itoa(i % keys))
	}
}

func BenchmarkMultiCounter_1KeySeq(b *testing.B)    { BenchMultiCounterSeq(b, 1) }
func BenchmarkMultiCounter_10KeySeq(b *testing.B)   { BenchMultiCounterSeq(b, 10) }
func BenchmarkMultiCounter_100KeySeq(b *testing.B)  { BenchMultiCounterSeq(b, 100) }
func BenchmarkMultiCounter_1000KeySeq(b *testing.B) { BenchMultiCounterSeq(b, 1000) }

func BenchmarkMultiCounter_ControlPara(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for i := 0; pb.Next(); i++ {
			strconv.Itoa(i)
		}
	})
}

func BenchMultiCounterPara(b *testing.B, keys int) {
	var multi MultiCounter

	b.RunParallel(func(pb *testing.PB) {
		for i := 0; pb.Next(); i++ {
			multi.RecordHit(strconv.Itoa(i % keys))
		}
	})
}

func BenchmarkMultiCounter_1KeyPara(b *testing.B)    { BenchMultiCounterPara(b, 1) }
func BenchmarkMultiCounter_10KeyPara(b *testing.B)   { BenchMultiCounterPara(b, 10) }
func BenchmarkMultiCounter_100KeyPara(b *testing.B)  { BenchMultiCounterPara(b, 100) }
func BenchmarkMultiCounter_1000KeyPara(b *testing.B) { BenchMultiCounterPara(b, 1000) }
