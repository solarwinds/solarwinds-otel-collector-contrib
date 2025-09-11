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

package mqttreceiver

type Metric struct {
	Name         string `mapstructure:"name"`
	Type         string `mapstructure:"type"`
	Topic        string `mapstructure:"topic"`
	Unit         string `mapstructure:"unit"`
	Desc         string `mapstructure:"description"`
	JsonProperty string `mapstructure:"json_property"`
}

type Sensor struct {
	Name    string    `mapstructure:"name"`
	Metrics []*Metric `mapstructure:"metrics"`
}

type Broker struct {
	Name     string    `mapstructure:"name"`
	Protocol string    `mapstructure:"protocol"`
	Host     string    `mapstructure:"host"`
	Port     int       `mapstructure:"port"`
	User     string    `mapstructure:"user"`
	Password string    `mapstructure:"password"`
	Sensors  []*Sensor `mapstructure:"sensors"`
}

type Config struct {
	Brokers []*Broker `mapstructure:"brokers"`
}

func (c *Config) Validate() error {
	return nil
}
