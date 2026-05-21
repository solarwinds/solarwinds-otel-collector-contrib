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

package swootelentityrefprocessor // import "github.com/solarwinds/solarwinds-otel-collector-contrib/processor/swootelentityrefprocessor"

import (
	"fmt"
	"strings"
)

// Action defines the behavior of the processor when processing signals.
type Action string

const (
	// ActionInsert appends EntityRefs to the Resource of each incoming signal.
	ActionInsert Action = "insert"
	// ActionRemoveAll clears all EntityRefs from the Resource of each incoming signal.
	ActionRemoveAll Action = "remove_all"
)

// EntityRefConfig holds the static configuration for a single EntityRef entry.
type EntityRefConfig struct {
	// Type is the entity type name (e.g. "KubernetesContainer").
	Type string `mapstructure:"type"`
	// IDKeys is the ordered list of Resource attribute keys that together
	// uniquely identify the entity. An EntityRef is only inserted when ALL
	// listed keys are present in the Resource attributes.
	IDKeys []string `mapstructure:"id_keys"`

	// normalizedType holds strings.ToLower(Type), pre-computed by Validate
	// so processResource avoids per-signal allocations.
	normalizedType string
}

// Config is the configuration for the swootelentityref processor.
type Config struct {
	// Action is the operation to perform: "insert" or "remove_all".
	Action Action `mapstructure:"action"`
	// EntityRefs is the list of EntityRef entries to evaluate when Action is "insert".
	EntityRefs []EntityRefConfig `mapstructure:"entity_refs"`
}

// Validate returns an error if the configuration is invalid.
func (c *Config) Validate() error {
	switch c.Action {
	case ActionInsert, ActionRemoveAll:
	default:
		return fmt.Errorf("invalid action %q: must be %q or %q", c.Action, ActionInsert, ActionRemoveAll)
	}
	if c.Action == ActionRemoveAll && len(c.EntityRefs) > 0 {
		return fmt.Errorf("entity_refs must be empty when action is %q", ActionRemoveAll)
	}
	seen := make(map[string]int, len(c.EntityRefs))
	for i := range c.EntityRefs {
		er := &c.EntityRefs[i]
		if er.Type == "" {
			return fmt.Errorf("entity_refs[%d]: type must not be empty", i)
		}
		if len(er.IDKeys) == 0 {
			return fmt.Errorf("entity_refs[%d]: id_keys must not be empty", i)
		}
		seenKeys := make(map[string]struct{}, len(er.IDKeys))
		for _, k := range er.IDKeys {
			if _, dup := seenKeys[k]; dup {
				return fmt.Errorf("entity_refs[%d]: duplicate id_key %q", i, k)
			}
			seenKeys[k] = struct{}{}
		}
		key := strings.ToLower(er.Type)
		if prev, dup := seen[key]; dup {
			return fmt.Errorf("entity_refs[%d]: duplicate type %q (already declared at index %d)", i, er.Type, prev)
		}
		seen[key] = i
		er.normalizedType = key
	}
	return nil
}
