// Copyright 2026 SolarWinds Worldwide, LLC. All rights reserved.
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
	"context"
	"errors"
	"strings"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/pdata/ptrace"
	"go.opentelemetry.io/collector/pdata/xpdata/entity"
	"go.uber.org/zap"
)

// errReadOnlyBatch is returned when the processor receives a read-only batch.
// Mutating a read-only batch panics; returning an error lets the pipeline drop
// the batch instead of crashing. Read-only batches occur when a fanout connector
// distributes a batch to a sibling branch with MutatesData: false — see README.
var errReadOnlyBatch = errors.New("swootelentityref: received a read-only batch; ensure all sibling pipeline branches declare MutatesData: true")

type swootelentityrefprocessor struct {
	logger *zap.Logger
	cfg    *Config
}

func (p *swootelentityrefprocessor) Start(_ context.Context, _ component.Host) error {

	p.logger.Info("Starting swootelentityrefprocessor processor")
	return nil
}

func (p *swootelentityrefprocessor) processLogs(_ context.Context, ld plog.Logs) (plog.Logs, error) {
	if ld.IsReadOnly() {
		return ld, errReadOnlyBatch
	}
	for _, rl := range ld.ResourceLogs().All() {
		p.processResource(rl.Resource())
	}
	return ld, nil
}

func (p *swootelentityrefprocessor) processMetrics(_ context.Context, md pmetric.Metrics) (pmetric.Metrics, error) {
	if md.IsReadOnly() {
		return md, errReadOnlyBatch
	}
	for _, rm := range md.ResourceMetrics().All() {
		p.processResource(rm.Resource())
	}
	return md, nil
}

func (p *swootelentityrefprocessor) processTraces(_ context.Context, td ptrace.Traces) (ptrace.Traces, error) {
	if td.IsReadOnly() {
		return td, errReadOnlyBatch
	}
	for _, rs := range td.ResourceSpans().All() {
		p.processResource(rs.Resource())
	}
	return td, nil
}

func (p *swootelentityrefprocessor) processResource(resource pcommon.Resource) {
	refs := entity.ResourceEntityRefs(resource)
	switch p.cfg.Action {
	case ActionInsert:
		// Snapshot existing ref types before the loop to avoid O(N×M) rescanning
		// and to eliminate any dependence on iterator re-evaluation between AppendEmpty calls.
		// Skipping a type that already exists preserves the caller's ref intact; it does not
		// update id_keys even if the current config lists different keys for that type.
		existingTypes := make(map[string]struct{}, refs.Len())
		for _, ref := range refs.All() {
			existingTypes[strings.ToLower(ref.Type())] = struct{}{}
		}
		attrs := resource.Attributes()
		for _, ercfg := range p.cfg.EntityRefs {
			if !allKeysPresent(attrs, ercfg.IDKeys) {
				continue
			}
			if _, exists := existingTypes[ercfg.normalizedType]; exists {
				continue
			}
			ref := refs.AppendEmpty()
			ref.SetType(ercfg.Type)
			idKeys := ref.IdKeys()
			for _, k := range ercfg.IDKeys {
				idKeys.Append(k)
			}
			existingTypes[ercfg.normalizedType] = struct{}{}
			p.logger.Debug("added entity ref", zap.String("type", ercfg.Type), zap.Any("idKeys", ercfg.IDKeys))
		}
	case ActionRemoveAll:
		// EntityRefSlice has no RemoveAll; RemoveIf with a tautological predicate is the standard clear pattern.
		refs.RemoveIf(func(_ entity.EntityRef) bool { return true })
	}
}

func allKeysPresent(attrs pcommon.Map, keys []string) bool {
	if len(keys) == 0 {
		return false
	}
	for _, k := range keys {
		v, ok := attrs.Get(k)
		if !ok || v.Type() != pcommon.ValueTypeStr || v.Str() == "" {
			return false
		}
	}
	return true
}
