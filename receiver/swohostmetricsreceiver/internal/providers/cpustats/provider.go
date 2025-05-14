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

const UserHz = 100

const (
	UserMode      = "user"
	NiceProc      = "nice"
	SystemMode    = "system"
	IdleState     = "idle"
	IOWait        = "io_waits"
	IRQ           = "irq"
	SoftIRQ       = "softirq"
	StealTime     = "steal"
	GuestTime     = "guest"
	GuestNiceProc = "guest_nice"
	TotalTime     = "total"

	FieldTypeCPUTime      = "cpu_time"
	FieldTypeProcesses    = "processes"
	FieldTypeCurrentProcs = "current_procs"
	FieldTypeIntr         = "intr"
	FieldTypeCtxt         = "ctxt"
	FieldTypeNumCores     = "numcores"
)

type WorkDetail struct {
	AttrName  string
	AttrValue string
	Value     float64
}

type Container struct {
	WorkDetails  map[string][]WorkDetail
	Error        error
	totalCPUTime float64
	numCPUs      float64
}
