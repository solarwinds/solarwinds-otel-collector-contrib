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

type Uptime struct {
	Uptime uint64
	Error  error
}

// Provider for host uptime.
type Provider interface {
	// Provides channel capable of delivering uptime.
	GetUptime() <-chan Uptime
}

type hostProvider struct {
	Wrapper
}

var _ Provider = (*hostProvider)(nil)

func CreateUptimeProvider(
	w Wrapper,
) Provider {
	return &hostProvider{
		Wrapper: w,
	}
}

// GetUptime implements Provider.
func (p *hostProvider) GetUptime() <-chan Uptime {
	ch := make(chan Uptime)
	go p.getUptimeInternal(ch)
	return ch
}

func (p *hostProvider) getUptimeInternal(ch chan Uptime) {
	defer close(ch)

	uptime, err := p.Wrapper.GetUptime()
	d := Uptime{
		Uptime: uptime,
		Error:  err,
	}

	ch <- d
}
