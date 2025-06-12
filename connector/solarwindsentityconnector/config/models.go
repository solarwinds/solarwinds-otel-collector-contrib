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

package config

type Entity struct {
	Type       string   `mapstructure:"entity"`
	IDs        []string `mapstructure:"id"`
	Attributes []string `mapstructure:"attributes"`
}

type Events struct {
	Relationships []RelationshipEvent `mapstructure:"relationships"`
	Entities      []EntityEvent       `mapstructure:"entities"`
}

type Relationship struct {
	Type        string `mapstructure:"type"`
	Source      string `mapstructure:"source_entity"`
	Destination string `mapstructure:"destination_entity"`
}

type RelationshipEvent struct {
	Type        string   `mapstructure:"type"`
	Source      string   `mapstructure:"source_entity"`
	Destination string   `mapstructure:"destination_entity"`
	Attributes  []string `mapstructure:"attributes"`
	Conditions  []string `mapstructure:"conditions"`
	Context     string   `mapstructure:"context"`
}

type EntityEvent struct {
	Context    string   `mapstructure:"context"`
	Conditions []string `mapstructure:"conditions"`
	Type       string   `mapstructure:"type"`
}
