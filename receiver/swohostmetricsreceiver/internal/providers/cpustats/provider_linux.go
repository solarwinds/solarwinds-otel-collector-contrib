// Copyright 2025 SolarWinds Worldwide, LLC. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cpustats

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/solarwinds/solarwinds-otel-collector/receiver/swohostmetricsreceiver/internal/providers"
)

const tickerInterval = 5 * time.Second

var (
	//nolint:gochecknoglobals // package private
	once sync.Once

	//nolint:gochecknoglobals // package private
	singleton *provider
)

type cpuStatsReader interface {
	Read() (io.ReadCloser, error)
}

type fileCPUStatsReader struct{}

func (f *fileCPUStatsReader) Read() (io.ReadCloser, error) {
	return os.Open("/proc/stat")
}

func CreateProvider() providers.Provider[Container] {
	return createProvider(&fileCPUStatsReader{})
}

func createProvider(reader cpuStatsReader) *provider {
	once.Do(func() {
		singleton = &provider{
			cpuStatsReader: reader,
		}

		go singleton.loop()
	})

	return singleton
}

type provider struct {
	cpuStatsReader
	container Container
	mx        sync.RWMutex
}

func (p *provider) loop() {
	ticker := time.NewTicker(tickerInterval)
	defer ticker.Stop()

	// Do a read immediately after start
	p.readCPUFileStats()

	for {
		<-ticker.C
		p.readCPUFileStats()
	}
}

func (p *provider) readCPUFileStats() {
	p.mx.Lock()
	defer p.mx.Unlock()

	file, err := p.cpuStatsReader.Read()
	if err != nil {
		p.container = Container{Error: err}
		return
	}

	defer file.Close()

	// 32KB should be enough for 128 CPUs
	buf := bufio.NewReaderSize(file, 32*1024)

	cpuStats := p.parseCPUStatsLinux(buf)
	p.container = cpuStats
}

func (p *provider) Provide() <-chan Container {
	ch := make(chan Container)

	go func() {
		defer close(ch)

		p.mx.RLock()
		defer p.mx.RUnlock()

		ch <- p.container
	}()

	return ch
}

func (p *provider) parseCPUStatsLinux(buf *bufio.Reader) Container {
	cpuStats := Container{
		WorkDetails: make(map[string][]WorkDetail),
	}

	for line, err := buf.ReadString('\n'); err == nil; line, err = buf.ReadString('\n') {
		parts := strings.Fields(line)

		if len(parts) == 0 {
			continue
		}

		if len(parts) < 2 {
			cpuStats.Error = fmt.Errorf("unexpected number of elements in line from /proc/stat: %d", len(parts))
			return cpuStats
		}

		p.processCPUStatsLine(parts, &cpuStats)
		if cpuStats.Error != nil {
			return cpuStats
		}
	}

	cpuStats.WorkDetails[FieldTypeCPUTime] = append(
		cpuStats.WorkDetails[FieldTypeCPUTime],
		WorkDetail{
			AttrName:  "mode",
			AttrValue: TotalTime,
			Value:     cpuStats.totalCPUTime,
		},
	)
	cpuStats.WorkDetails[FieldTypeNumCores] = append(
		cpuStats.WorkDetails[FieldTypeNumCores],
		WorkDetail{
			Value: float64(cpuStats.numCPUs),
		},
	)

	return cpuStats
}

func (p *provider) processCPUStatsLine(parts []string, cpuStats *Container) {
	// Multiply by this to convert USER_HZ ticks into microseconds
	factorToUs := float64(1000000) / UserHz

	rowtype := parts[0]
	fields := parts[1:]

	switch rowtype {
	case "cpu":
		for i, valueStr := range fields {
			value, err := strconv.ParseFloat(valueStr, 64)
			if err != nil {
				cpuStats.Error = fmt.Errorf("invalid value for field in cpu row: %w", err)
				return
			}

			cpuStats.WorkDetails[FieldTypeCPUTime] = append(
				cpuStats.WorkDetails[FieldTypeCPUTime],
				WorkDetail{
					AttrName:  "mode",
					AttrValue: getCPULineFormat()[i],
					Value:     value * factorToUs,
				},
			)

			// computes aggregate usage for cpu excluding "idle" (i=3)
			if i != 3 {
				cpuStats.totalCPUTime += value * factorToUs
			}
		}
	case "ctxt", "processes", "intr", "procs_running", "procs_blocked":
		value, err := strconv.ParseFloat(fields[0], 64)
		if err != nil {
			cpuStats.Error = fmt.Errorf("invalid value for rowtype %s: %w", rowtype, err)
			return
		}

		attrName := ""
		if rowtype == "procs_running" || rowtype == "procs_blocked" {
			attrName = "state"
		}

		cpuStats.WorkDetails[getFieldMapping()[rowtype]] = append(
			cpuStats.WorkDetails[getFieldMapping()[rowtype]],
			WorkDetail{
				AttrName:  attrName,
				AttrValue: rowtype,
				Value:     value,
			},
		)
	default:
		if _, err := strconv.Atoi(strings.TrimPrefix(rowtype, "cpu")); err == nil {
			// This line was of the form cpux, so it's reporting per-core stats.
			cpuStats.numCPUs++
		}
	}
}

var _ providers.Provider[Container] = (*provider)(nil)

func getCPULineFormat() []string {
	return []string{
		UserMode,
		NiceProc,
		SystemMode,
		IdleState,
		IOWait,
		IRQ,
		SoftIRQ,
		StealTime,
		GuestTime,
		GuestNiceProc,
	}
}

func getFieldMapping() map[string]string {
	return map[string]string{
		"ctxt":          FieldTypeCtxt,
		"processes":     FieldTypeProcesses,
		"intr":          FieldTypeIntr,
		"procs_running": FieldTypeCurrentProcs,
		"procs_blocked": FieldTypeCurrentProcs,
	}
}
