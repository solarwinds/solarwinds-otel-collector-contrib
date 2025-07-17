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

import (
	"errors"
	"fmt"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/ottl/contexts/ottllog"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/ottl/contexts/ottlmetric"
	"go.opentelemetry.io/collector/confmap/xconfmap"
)

type Entity struct {
	Entity     string   `mapstructure:"entity"`
	IDs        []string `mapstructure:"id"`
	Attributes []string `mapstructure:"attributes"`
}

// By implementing the xconfmap.Validator interface, we ensure it's validated by the collector automatically
var _ xconfmap.Validator = (*Entity)(nil)

func (e *Entity) Validate() error {
	var errs error

	if e.Entity == "" {
		errs = errors.Join(errs, fmt.Errorf("entity is mandatory"))
	}

	if len(e.IDs) == 0 {
		errs = errors.Join(errs, fmt.Errorf("id is mandatory and must contain at least 1 item"))
	}

	return errs
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
	Action      string   `mapstructure:"action"`
}

// By implementing the xconfmap.Validator interface, we ensure it's validated by the collector automatically
var _ xconfmap.Validator = (*RelationshipEvent)(nil)

func (r *RelationshipEvent) Validate() error {
	var errs error

	if r.Action == "" {
		errs = errors.Join(errs, fmt.Errorf("action is mandatory"))
	} else if r.Action != "update" && r.Action != "delete" {
		errs = errors.Join(errs, fmt.Errorf("action must be 'update' or 'delete', got '%s'", r.Action))
	}

	if r.Context == "" {
		errs = errors.Join(errs, fmt.Errorf("context is mandatory"))
	} else if r.Context != ottllog.ContextName && r.Context != ottlmetric.ContextName {
		errs = errors.Join(errs, fmt.Errorf("context must be '%s' or '%s', got '%s'",
			ottllog.ContextName, ottlmetric.ContextName, r.Context))
	}

	if r.Type == "" {
		errs = errors.Join(errs, fmt.Errorf("type is mandatory"))
	}

	if r.Source == "" {
		errs = errors.Join(errs, fmt.Errorf("source_entity is mandatory"))
	}

	if r.Destination == "" {
		errs = errors.Join(errs, fmt.Errorf("destination_entity is mandatory"))
	}

	return errs
}

type EntityEvent struct {
	Context    string   `mapstructure:"context"`
	Conditions []string `mapstructure:"conditions"`
	Entity     string   `mapstructure:"entity"`
	Action     string   `mapstructure:"action"`
}

// By implementing the xconfmap.Validator interface, we ensure it's validated by the collector automatically
var _ xconfmap.Validator = (*EntityEvent)(nil)

func (e *EntityEvent) Validate() error {
	var errs error

	if e.Action == "" {
		errs = errors.Join(errs, fmt.Errorf("action is mandatory"))
	} else if e.Action != "update" && e.Action != "delete" {
		errs = errors.Join(errs, fmt.Errorf("action must be 'update' or 'delete', got '%s'",
			e.Action))
	}

	if e.Context == "" {
		errs = errors.Join(errs, fmt.Errorf("context is mandatory"))
	} else if e.Context != ottllog.ContextName && e.Context != ottlmetric.ContextName {
		errs = errors.Join(errs, fmt.Errorf("context must be '%s' or '%s', got '%s'",
			ottllog.ContextName, ottlmetric.ContextName, e.Context))
	}

	if e.Entity == "" {
		errs = errors.Join(errs, fmt.Errorf("type is mandatory"))
	}

	return errs
}
