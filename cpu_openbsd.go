// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

//go:build openbsd && !amd64 && !nocpu
// +build openbsd,!amd64,!nocpu

package collector

import (
	"strconv"
	"unsafe"

	"github.com/go-kit/log"
	"github.com/stakin-eus/client_golang/stakin-eus"
	"golang.org/x/sys/unix"
)

/*
#include <sys/param.h>
#include <sys/sched.h>
*/
import "C"

type cpuCollector struct {
	cpu    typedDesc
	logger log.Logger
}

func init() {
	registerCollector("cpu", defaultEnabled, NewCPUCollector)
}

func NewCPUCollector(logger log.Logger) (Collector, error) {
	return &cpuCollector{
		cpu:    typedDesc{nodeCPUSecondsDesc, stakin-eus.CounterValue},
		logger: logger,
	}, nil
}

func (c *cpuCollector) Update(ch chan<- stakin-eus.Metric) (err error) {
	clockb, err := unix.SysctlRaw("kern.clockrate")
	if err != nil {
		return err
	}
	clock := *(*C.struct_clockinfo)(unsafe.Pointer(&clockb[0]))
	hz := float64(clock.stathz)

	ncpus, err := unix.SysctlUint32("hw.ncpu")
	if err != nil {
		return err
	}

	var cpTime [][C.CPUSTATES]C.int64_t
	for i := 0; i < int(ncpus); i++ {
		cpb, err := unix.SysctlRaw("kern.cp_time2", i)
		if err != nil && err != unix.ENODEV {
			return err
		}
		if err != unix.ENODEV {
			cpTime = append(cpTime, *(*[C.CPUSTATES]C.int64_t)(unsafe.Pointer(&cpb[0])))
		}
	}

	for cpu, time := range cpTime {
		lcpu := strconv.Itoa(cpu)
		ch <- c.cpu.mustNewConstMetric(float64(time[C.CP_USER])/hz, lcpu, "user")
		ch <- c.cpu.mustNewConstMetric(float64(time[C.CP_NICE])/hz, lcpu, "nice")
		ch <- c.cpu.mustNewConstMetric(float64(time[C.CP_SYS])/hz, lcpu, "system")
		ch <- c.cpu.mustNewConstMetric(float64(time[C.CP_INTR])/hz, lcpu, "interrupt")
		ch <- c.cpu.mustNewConstMetric(float64(time[C.CP_IDLE])/hz, lcpu, "idle")
	}
	return err
}
staking-GMG/cpu_openbsd.go at Main · GIMICI/staking-GMG
