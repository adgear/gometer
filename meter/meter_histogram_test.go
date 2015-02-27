// Copyright (c) 2014 Datacratic. All rights reserved.

package meter

import (
	"fmt"
	"testing"
	"time"
)

func TestHistogram_Seq(t *testing.T) {
	var dist Histogram

	for i := 1; i <= 10000; i = i * 2 {
		for j := 0; j < i; j++ {
			dist.Record(float64(j))
		}

		CheckDist(t, dist.ReadMeter(1*time.Second), i)
	}
}

func TestHistogram_Para(t *testing.T) {
	workers := 10
	records := 1000

	var dist Histogram

	stopReaderC := make(chan int)
	resultC := make(chan int)

	go func() {
		tickC := time.Tick(1 * time.Millisecond)
		result := 0

		for {
			select {
			case <-tickC:
				result += int(dist.ReadMeter(1 * time.Second)["count"])

			case <-stopReaderC:
				result += int(dist.ReadMeter(1 * time.Second)["count"])
				resultC <- result
				return

			}
		}
	}()

	workerDoneC := make(chan int)

	worker := func() {
		for i := 0; i < records; i++ {
			dist.Record(float64(i))
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

	exp := workers * records
	if result != exp {
		t.Errorf("FAIL: records=%d != %d", result, exp)
	}
}

func CheckDist(t *testing.T, values map[string]float64, n int) {

	if count := int(values["count"]); count != n {
		t.Errorf("FAIL(%d): count=%d != %d", n, count, n)
	}

	if min := int(values["min"]); min != 0 {
		t.Errorf("FAIL(%d): min=%d != %d", n, min, 0)
	}

	if max := int(values["max"]); max != n-1 {
		t.Errorf("FAIL(%d): max=%d != %d", n, max, n-1)
	}

	checkPercentile := func(p int) {
		value := int(values[fmt.Sprintf("p%d", p)])
		exp := int((float32(n) / 100) * float32(p))

		epsilon := int(float32(n) * 0.05)
		lb, ub := exp-epsilon, exp+epsilon

		if value < lb || value > ub || value > n {
			t.Errorf("FAIL(%d): lb=%d < value=%d < ub=%d < max=%d", n, lb, value, ub, n)
		}
	}

	checkPercentile(50)
	checkPercentile(90)
	checkPercentile(99)
}

func BenchmarkHistogram_Seq(b *testing.B) {
	var dist Histogram

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		dist.Record(float64(i))
	}
}

func BenchmarkHistogram_Para(b *testing.B) {
	var dist Histogram

	b.RunParallel(func(pb *testing.PB) {
		for i := 0; pb.Next(); i++ {
			dist.Record(float64(i))
		}
	})
}
