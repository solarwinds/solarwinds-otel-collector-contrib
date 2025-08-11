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

package processescount

import (
	"unsafe"

	"golang.org/x/sys/windows"
)

// From https://github.com/shirou/gopsutil/blob/ffcdc2b7662f84ead7114fbd5e841c83f95116ac/mem/mem_windows.go
var modPsapi = windows.NewLazySystemDLL("psapi.dll")
var procGetPerformanceInfo = modPsapi.NewProc("GetPerformanceInfo")

type performanceInformation struct {
	cb                uint32
	commitTotal       uint64
	commitLimit       uint64
	commitPeak        uint64
	physicalTotal     uint64
	physicalAvailable uint64
	systemCache       uint64
	kernelTotal       uint64
	kernelPaged       uint64
	kernelNonpaged    uint64
	pageSize          uint64
	handleCount       uint32
	processCount      uint32
	threadCount       uint32
}

func (*wrapper) GetCount() (int64, error) {
	var perfInfo performanceInformation
	perfInfo.cb = uint32(unsafe.Sizeof(perfInfo))
	// GetPerformanceInfo returns 0 for error, in which case we check err,
	// see https://pkg.go.dev/golang.org/x/sys/windows#LazyProc.Call
	mem, _, err := procGetPerformanceInfo.Call(uintptr(unsafe.Pointer(&perfInfo)), uintptr(perfInfo.cb))
	if mem == 0 {
		return 0, err
	}
	return int64(perfInfo.processCount), nil
}
