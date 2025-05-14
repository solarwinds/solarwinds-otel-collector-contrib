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
	"context"
	"testing"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/stretchr/testify/assert"
)

type mockCpuStatsProvider struct{}

func (m *mockCpuStatsProvider) TimesWithContext(ctx context.Context, percpu bool) ([]cpu.TimesStat, error) {
	return []cpu.TimesStat{
		{
			CPU:    "cpu-total",
			System: 3.0,
			Idle:   0.2,
			User:   4.0,
		},
	}, nil
}

func Test_provide_Success(t *testing.T) {
	p := &provider{cpuStatsProvider: &mockCpuStatsProvider{}}
	container := <-p.Provide()

	assert.EqualValues(
		t,
		[]WorkDetail{
			{AttrName: "mode", AttrValue: "user", Value: 4000000},
			{AttrName: "mode", AttrValue: "system", Value: 3000000},
			{AttrName: "mode", AttrValue: "idle", Value: 200000},
			{AttrName: "mode", AttrValue: "total", Value: 7000000},
		},
		container.WorkDetails["cpu_time"],
	)
}
