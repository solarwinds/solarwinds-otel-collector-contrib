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

const (
	// Event action types
	EventUpdateAction = "update"
	EventDeleteAction = "delete"
)

type Entity struct {
	Entity     string   `mapstructure:"entity"`
	IDs        []string `mapstructure:"id"`
	Attributes []string `mapstructure:"attributes"`
}

// By implementing the xconfmap.Validator, we ensure it's validated by the collector automatically
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

type Event struct {
	Action     string   `mapstructure:"action"`
	Context    string   `mapstructure:"context"`
	Conditions []string `mapstructure:"conditions"`
}

func (e *Event) validateActionAndContext() error {
	var errs error

	if e.Action == "" {
		errs = errors.Join(errs, fmt.Errorf("action is mandatory"))
	} else if e.Action != EventUpdateAction && e.Action != EventDeleteAction {
		errs = errors.Join(errs, fmt.Errorf("action must be '%s' or '%s', got '%s'",
			EventUpdateAction, EventDeleteAction, e.Action))
	}

	if e.Context == "" {
		errs = errors.Join(errs, fmt.Errorf("context is mandatory"))
	} else if e.Context != ottllog.ContextName && e.Context != ottlmetric.ContextName {
		errs = errors.Join(errs, fmt.Errorf("context must be '%s' or '%s', got '%s'",
			ottllog.ContextName, ottlmetric.ContextName, e.Context))
	}
	return errs
}

type RelationshipEvent struct {
	Event
	Type        string   `mapstructure:"type"`
	Source      string   `mapstructure:"source_entity"`
	Destination string   `mapstructure:"destination_entity"`
	Attributes  []string `mapstructure:"attributes"`
}

// By implementing the xconfmap.Validator, we ensure it's validated by the collector automatically
var _ xconfmap.Validator = (*RelationshipEvent)(nil)

func (e *RelationshipEvent) Validate() error {
	errs := e.validateActionAndContext()

	if e.Type == "" {
		errs = errors.Join(errs, fmt.Errorf("type is mandatory"))
	}

	if e.Source == "" {
		errs = errors.Join(errs, fmt.Errorf("source_entity is mandatory"))
	}

	if e.Destination == "" {
		errs = errors.Join(errs, fmt.Errorf("destination_entity is mandatory"))
	}

	return errs
}

type EntityEvent struct {
	Event
	Entity string `mapstructure:"entity"`
}

// By implementing the xconfmap.Validator, we ensure it's validated by the collector automatically
var _ xconfmap.Validator = (*EntityEvent)(nil)

func (e *EntityEvent) Validate() error {
	errs := e.validateActionAndContext()

	if e.Entity == "" {
		errs = errors.Join(errs, fmt.Errorf("entity is mandatory"))
	}

	return errs
}
