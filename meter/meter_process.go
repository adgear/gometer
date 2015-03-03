// Copyright (c) 2014 Datacratic. All rights reserved.

package meter

import (
	"github.com/datacratic/goklog/klog"

	"fmt"
	"io/ioutil"
	"runtime"
	"syscall"
	"time"
)

type process struct {
	Boot    *Counter
	Running *Gauge

	Load *Gauge

	Golang struct {
		Threads    *Gauge
		Goroutines *Gauge
	}

	Processor struct {
		UserTime   *Gauge
		SystemTime *Gauge

		ContextSwitchesV *Gauge
		ContextSwitchesI *Gauge
	}

	Memory struct {
		Resident *Gauge
		Virtual  *Gauge
		Shared   *Gauge

		MinorFaults *Gauge
		MajorFaults *Gauge
		Swaps       *Gauge
	}

	Block struct {
		InOps  *Gauge
		OutOps *Gauge
	}

	lastRusage syscall.Rusage
}

func ProcessStats(prefix string) {
	meter := &process{}
	Load(meter, Join(prefix, "process"))

	go func() {
		meter.Boot.Hit()
		meter.Running.Change(1)
		meter.lastRusage = meter.rusage()

		tickC := time.Tick(1 * time.Second)
		for {
			<-tickC
			meter.update()
		}
	}()
}

func (meter *process) rusage() (result syscall.Rusage) {
	if err := syscall.Getrusage(syscall.RUSAGE_SELF, &result); err != nil {
		klog.KFatalf("meter.process.rusage.error", err.Error())
	}
	return
}

func (meter *process) update() {
	rusage := meter.rusage()

	meter.Golang.Goroutines.Change(float64(runtime.NumGoroutine()))

	threads := runtime.GOMAXPROCS(0)
	meter.Golang.Threads.Change(float64(threads))

	utime := time.Duration(rusage.Utime.Nano() - meter.lastRusage.Utime.Nano())
	meter.Processor.UserTime.ChangeDuration(utime)

	stime := time.Duration(rusage.Stime.Nano() - meter.lastRusage.Stime.Nano())
	meter.Processor.SystemTime.ChangeDuration(stime)

	meter.Load.Change(float64(utime + stime) / float64(time.Duration(threads) * time.Second))

	meter.Processor.ContextSwitchesI.Change(float64(rusage.Nvcsw - meter.lastRusage.Nvcsw))
	meter.Processor.ContextSwitchesV.Change(float64(rusage.Nivcsw - meter.lastRusage.Nivcsw))

	meter.Memory.MinorFaults.Change(float64(rusage.Minflt - meter.lastRusage.Minflt))
	meter.Memory.MajorFaults.Change(float64(rusage.Majflt - meter.lastRusage.Majflt))
	meter.Memory.Swaps.Change(float64(rusage.Nswap))

	meter.Block.InOps.Change(float64(rusage.Inblock - meter.lastRusage.Inblock))
	meter.Block.OutOps.Change(float64(rusage.Oublock - meter.lastRusage.Oublock))

	meter.lastRusage = rusage
	
	meter.statm()
}

func (meter *process) statm() {
	body, err := ioutil.ReadFile("/proc/self/statm")
	if err != nil {
		klog.KFatalf("meter.process.statm.error", err.Error())
	}

	var virt, rss, shared uint64
	if _, err := fmt.Sscanf(string(body), "%d %d %d", &virt, &rss, &shared); err != nil {
		klog.KFatalf("meter.process.statm.parse.error", err.Error())
	}

	meter.Memory.Resident.Change(float64(rss))
	meter.Memory.Virtual.Change(float64(virt))
	meter.Memory.Shared.Change(float64(shared))
}
