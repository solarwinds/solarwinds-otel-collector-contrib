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

package internal

import (
	"encoding/json"
	"fmt"
	"hash/fnv"

	"github.com/solarwinds/solarwinds-otel-collector-contrib/connector/solarwindsentityconnector/internal"
)

// KeyBuilder provides methods for generating consistent keys for entities and relationships
type KeyBuilder interface {
	// BuildEntityKey constructs a unique key for an entity
	BuildEntityKey(entity Entity) (string, error)

	// BuildRelationshipKey constructs a unique key for a relationship using the relationship type
	// and the hashed keys of the source and destination entities
	BuildRelationshipKey(relationshipType string, sourceHash string, destHash string) (string, error)
}

// defaultKeyBuilder provides the default implementation of KeyBuilder interface
type defaultKeyBuilder struct{}

// NewKeyBuilder creates a new instance of the default key builder
func NewKeyBuilder() KeyBuilder {
	return &defaultKeyBuilder{}
}

// BuildEntityKey constructs a unique key for the entity referenced in the relationship.
// The key is composition of entity type and its ID attributes.
func (b *defaultKeyBuilder) BuildEntityKey(entity Entity) (string, error) {
	data := struct {
		Type string
		IDs  map[string]any
	}{
		Type: entity.Type,
		IDs:  entity.IDs.AsRaw(),
	}
	return hashObject(data)
}

// BuildRelationshipKey constructs a key using the relationship type and the hashed keys
// of the source and destination entities. Returns an error if any input is empty.
func (b *defaultKeyBuilder) BuildRelationshipKey(relationshipType string, sourceHash string, destHash string) (string, error) {
	if relationshipType == "" {
		return "", fmt.Errorf("relationshipType cannot be empty")
	}
	if sourceHash == "" {
		return "", fmt.Errorf("sourceHash cannot be empty")
	}
	if destHash == "" {
		return "", fmt.Errorf("destHash cannot be empty")
	}
	return fmt.Sprintf("%s:%s:%s", relationshipType, sourceHash, destHash), nil
}

func hashObject(data interface{}) (string, error) {
	bytes, err := json.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("failed to encode: %w", err)
	}

	// Create hash from the marshaled bytes
	h := fnv.New64a()
	_, err = h.Write(bytes)
	if err != nil {
		return "", fmt.Errorf("failed to write bytes to hash: %w", err)
	}
	return fmt.Sprintf("%x", h.Sum64()), nil
}
