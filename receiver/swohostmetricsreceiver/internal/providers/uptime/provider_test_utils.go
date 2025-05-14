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

package uptime

type succeedingUptimeWrapper struct {
	uptime uint64
}

var _ Wrapper = (*succeedingUptimeWrapper)(nil)

func CreateSuceedingUptimeWrapper(
	uptime uint64,
) Wrapper {
	return &succeedingUptimeWrapper{
		uptime: uptime,
	}
}

// GetUptime implements UptimeWrapper.
func (w *succeedingUptimeWrapper) GetUptime() (uint64, error) {
	return w.uptime, nil
}

type failingUptimeWrapper struct {
	err error
}

var _ Wrapper = (*failingUptimeWrapper)(nil)

func CreateFailingUptimeWrapper(
	err error,
) Wrapper {
	return &failingUptimeWrapper{
		err: err,
	}
}

// GetUptime implements UptimeWrapper.
func (w *failingUptimeWrapper) GetUptime() (uint64, error) {
	return 0, w.err
}
