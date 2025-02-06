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
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

const procStatContents = `cpu  27843473 836 8380015 1412260133 457893 0 1307738 903476 0 0
cpu0 14125191 388 4344332 706380870 331668 0 65189 451504 0 0
cpu1 13718282 448 4035683 705879263 126224 0 1242548 451971 0 0
intr 9570161979 6 0 0 0 76 0 0 0 1 0 0 0 1 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 3000045416 72509530 0 0 681601920 0 3065590190 54733739 0 0 585911412 0 262 10192992 704670616 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0
ctxt 14558051079
btime 1723558363
processes 7124433
procs_running 1
procs_blocked 0
softirq 2798882917 0 444499156 444213944 430715330 10166792 0 7185056 998638904 16706 463447029`

type cpuStatsReaderMock struct {
	err error
}

func (c *cpuStatsReaderMock) Read() (io.ReadCloser, error) {
	return io.NopCloser(strings.NewReader(procStatContents)), c.err
}

func Test_provide_Success(t *testing.T) {
	reader := &cpuStatsReaderMock{}
	p := &provider{cpuStatsReader: reader}

	p.readCPUFileStats()

	container := <-p.Provide()

	assert.EqualValues(
		t,
		[]WorkDetail{
			{AttrValue: "ctxt", Value: 14558051079},
		},
		container.WorkDetails["ctxt"],
	)
	assert.EqualValues(
		t,
		[]WorkDetail{
			{AttrName: "state", AttrValue: "procs_running", Value: 1},
			{AttrName: "state", AttrValue: "procs_blocked", Value: 0},
		},
		container.WorkDetails["current_procs"],
	)
	assert.EqualValues(
		t,
		[]WorkDetail{
			{AttrValue: "intr", Value: 9570161979},
		},
		container.WorkDetails["intr"],
	)
	assert.EqualValues(
		t,
		[]WorkDetail{
			{Value: 2},
		},
		container.WorkDetails["numcores"],
	)
	assert.EqualValues(
		t,
		[]WorkDetail{
			{AttrValue: "processes", Value: 7124433},
		},
		container.WorkDetails["processes"],
	)
	assert.EqualValues(
		t,
		[]WorkDetail{
			{AttrName: "mode", AttrValue: "user", Value: 278434730000},
			{AttrName: "mode", AttrValue: "nice", Value: 8360000},
			{AttrName: "mode", AttrValue: "system", Value: 83800150000},
			{AttrName: "mode", AttrValue: "idle", Value: 14122601330000},
			{AttrName: "mode", AttrValue: "io_waits", Value: 4578930000},
			{AttrName: "mode", AttrValue: "irq", Value: 0},
			{AttrName: "mode", AttrValue: "softirq", Value: 13077380000},
			{AttrName: "mode", AttrValue: "steal", Value: 9034760000},
			{AttrName: "mode", AttrValue: "guest", Value: 0},
			{AttrName: "mode", AttrValue: "guest_nice", Value: 0},
			{AttrName: "mode", AttrValue: "total", Value: 388934310000},
		},
		container.WorkDetails["cpu_time"],
	)
}
