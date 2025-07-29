package config

import "github.com/open-telemetry/opentelemetry-collector-contrib/pkg/ottl"

// ParsedEventInterface represents a unified interface for parsed events.
// This interface allows both entity and relationship events to be stored and processed uniformly.
type ParsedEventInterface[C any] interface {
	IsEntityEvent() bool
	GetEntityEvent() *EntityEvent
	GetRelationshipEvent() *RelationshipEvent
	GetConditionSeq() ottl.ConditionSequence[C]
}

// EntityParsedEvent wraps an EntityEvent with its pre-parsed OTTL conditions.
type EntityParsedEvent[C any] struct {
	Definition   *EntityEvent
	ConditionSeq ottl.ConditionSequence[C]
}

var _ ParsedEventInterface[any] = (*EntityParsedEvent[any])(nil)

func (e EntityParsedEvent[C]) IsEntityEvent() bool                        { return true }
func (e EntityParsedEvent[C]) GetEntityEvent() *EntityEvent               { return e.Definition }
func (e EntityParsedEvent[C]) GetRelationshipEvent() *RelationshipEvent   { return nil }
func (e EntityParsedEvent[C]) GetConditionSeq() ottl.ConditionSequence[C] { return e.ConditionSeq }

// RelationshipParsedEvent wraps a RelationshipEvent with its pre-parsed OTTL conditions.
type RelationshipParsedEvent[C any] struct {
	Definition   *RelationshipEvent
	ConditionSeq ottl.ConditionSequence[C]
}

var _ ParsedEventInterface[any] = (*RelationshipParsedEvent[any])(nil)

func (r RelationshipParsedEvent[C]) IsEntityEvent() bool                      { return false }
func (r RelationshipParsedEvent[C]) GetEntityEvent() *EntityEvent             { return nil }
func (r RelationshipParsedEvent[C]) GetRelationshipEvent() *RelationshipEvent { return r.Definition }
func (r RelationshipParsedEvent[C]) GetConditionSeq() ottl.ConditionSequence[C] {
	return r.ConditionSeq
}
