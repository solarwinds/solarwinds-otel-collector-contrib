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

package swootelentityrefprocessor

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/pdata/ptrace"
	"go.opentelemetry.io/collector/pdata/xpdata/entity"
	"go.uber.org/zap"
)

// buildLogs creates a plog.Logs with a single ResourceLogs. The caller
// receives the Resource so it can populate attributes before processing.
func buildLogs(setAttrs func(pcommon.Map)) plog.Logs {
	ld := plog.NewLogs()
	rl := ld.ResourceLogs().AppendEmpty()
	setAttrs(rl.Resource().Attributes())
	return ld
}

func buildTestProcessor(cfg component.Config) (*swootelentityrefprocessor, error) {
	logger := zap.NewNop()
	return buildProcessor(cfg, logger)
}

// extractEntityRefs extracts the EntityRefs from a Resource into a plain slice for assertions.
func extractEntityRefs(resource pcommon.Resource) []struct {
	typ    string
	idKeys []string
} {
	refs := entity.ResourceEntityRefs(resource)
	out := make([]struct {
		typ    string
		idKeys []string
	}, refs.Len())
	for i, ref := range refs.All() {
		out[i] = struct {
			typ    string
			idKeys []string
		}{typ: ref.Type(), idKeys: ref.IdKeys().AsRaw()}
	}
	return out
}

func collectEntityRefs(ld plog.Logs) []struct {
	typ    string
	idKeys []string
} {
	if ld.ResourceLogs().Len() == 0 {
		return nil
	}
	return extractEntityRefs(ld.ResourceLogs().At(0).Resource())
}

// --- Logs tests ---

func TestInsertLogs_AllIDKeysPresent(t *testing.T) {
	cfg := &Config{
		Action: ActionInsert,
		EntityRefs: []EntityRefConfig{
			{Type: "KubernetesContainer", IDKeys: []string{"sw.k8s.cluster.uid", "k8s.namespace.name", "k8s.pod.name", "k8s.container.name"}},
		},
	}
	ld := buildLogs(func(m pcommon.Map) {
		m.PutStr("sw.k8s.cluster.uid", "cluster-1")
		m.PutStr("k8s.namespace.name", "default")
		m.PutStr("k8s.pod.name", "my-pod")
		m.PutStr("k8s.container.name", "my-container")
	})

	p, err := buildTestProcessor(cfg)
	require.NoError(t, err)
	result, err := p.processLogs(context.Background(), ld)
	require.NoError(t, err)

	refs := collectEntityRefs(result)
	require.Len(t, refs, 1)
	assert.Equal(t, "KubernetesContainer", refs[0].typ)
	assert.Equal(t, []string{"sw.k8s.cluster.uid", "k8s.namespace.name", "k8s.pod.name", "k8s.container.name"}, refs[0].idKeys)
}

func TestInsertLogs_OneIdKeyMissing(t *testing.T) {
	cfg := &Config{
		Action: ActionInsert,
		EntityRefs: []EntityRefConfig{
			{Type: "KubernetesContainer", IDKeys: []string{"sw.k8s.cluster.uid", "k8s.namespace.name", "k8s.pod.name", "k8s.container.name"}},
		},
	}
	ld := buildLogs(func(m pcommon.Map) {
		m.PutStr("sw.k8s.cluster.uid", "cluster-1")
		m.PutStr("k8s.namespace.name", "default")
		m.PutStr("k8s.pod.name", "my-pod")
		// k8s.container.name intentionally absent
	})

	p, err := buildTestProcessor(cfg)
	require.NoError(t, err)
	result, err := p.processLogs(context.Background(), ld)
	require.NoError(t, err)

	refs := collectEntityRefs(result)
	assert.Empty(t, refs)
}

// TestInsertLogs_MultipleEntries_PartialMatch: 8-entry k8s config, only mandatory
// attributes present → Container, Pod, Node inserted; all workload types skipped.
func TestInsertLogs_MultipleEntries_PartialMatch(t *testing.T) {
	cfg := &Config{
		Action: ActionInsert,
		EntityRefs: []EntityRefConfig{
			{Type: "KubernetesContainer", IDKeys: []string{"sw.k8s.cluster.uid", "k8s.namespace.name", "k8s.pod.name", "k8s.container.name"}},
			{Type: "KubernetesPod", IDKeys: []string{"sw.k8s.cluster.uid", "k8s.namespace.name", "k8s.pod.name"}},
			{Type: "KubernetesNode", IDKeys: []string{"sw.k8s.cluster.uid", "k8s.node.name"}},
			{Type: "KubernetesDeployment", IDKeys: []string{"sw.k8s.cluster.uid", "k8s.namespace.name", "k8s.deployment.name"}},
			{Type: "KubernetesStatefulSet", IDKeys: []string{"sw.k8s.cluster.uid", "k8s.namespace.name", "k8s.statefulset.name"}},
			{Type: "KubernetesDaemonSet", IDKeys: []string{"sw.k8s.cluster.uid", "k8s.namespace.name", "k8s.daemonset.name"}},
			{Type: "KubernetesJob", IDKeys: []string{"sw.k8s.cluster.uid", "k8s.namespace.name", "k8s.job.name"}},
			{Type: "KubernetesCronJob", IDKeys: []string{"sw.k8s.cluster.uid", "k8s.namespace.name", "k8s.cronjob.name"}},
		},
	}
	ld := buildLogs(func(m pcommon.Map) {
		m.PutStr("sw.k8s.cluster.uid", "cluster-1")
		m.PutStr("k8s.namespace.name", "default")
		m.PutStr("k8s.pod.name", "standalone-pod")
		m.PutStr("k8s.container.name", "app")
		m.PutStr("k8s.node.name", "node-1")
		// No workload-specific attributes
	})

	p, err := buildTestProcessor(cfg)
	require.NoError(t, err)
	result, err := p.processLogs(context.Background(), ld)
	require.NoError(t, err)

	refs := collectEntityRefs(result)
	require.Len(t, refs, 3)
	types := make([]string, len(refs))
	for i, r := range refs {
		types[i] = r.typ
	}
	assert.Contains(t, types, "KubernetesContainer")
	assert.Contains(t, types, "KubernetesPod")
	assert.Contains(t, types, "KubernetesNode")
}

func TestInsertLogs_EmptyConfig(t *testing.T) {
	cfg := &Config{
		Action:     ActionInsert,
		EntityRefs: nil,
	}
	ld := buildLogs(func(m pcommon.Map) {
		m.PutStr("sw.k8s.cluster.uid", "cluster-1")
	})

	p, err := buildTestProcessor(cfg)
	require.NoError(t, err)
	result, err := p.processLogs(context.Background(), ld)
	require.NoError(t, err)

	refs := collectEntityRefs(result)
	assert.Empty(t, refs)
}

func TestRemoveLogs_WithEntityRefs(t *testing.T) {
	cfg := &Config{Action: ActionRemove}
	ld := buildLogs(func(m pcommon.Map) {})

	// Pre-populate EntityRefs on the resource
	resource := ld.ResourceLogs().At(0).Resource()
	refs := entity.ResourceEntityRefs(resource)
	r1 := refs.AppendEmpty()
	r1.SetType("KubernetesPod")
	r1.IdKeys().Append("sw.k8s.cluster.uid")
	r2 := refs.AppendEmpty()
	r2.SetType("KubernetesNode")
	r2.IdKeys().Append("sw.k8s.cluster.uid")

	require.Equal(t, 2, entity.ResourceEntityRefs(resource).Len())

	p, err := buildTestProcessor(cfg)
	require.NoError(t, err)
	result, err := p.processLogs(context.Background(), ld)
	require.NoError(t, err)

	assert.Equal(t, 0, entity.ResourceEntityRefs(result.ResourceLogs().At(0).Resource()).Len())
}

func TestRemoveLogs_NoEntityRefs(t *testing.T) {
	cfg := &Config{Action: ActionRemove}
	ld := buildLogs(func(m pcommon.Map) {
		m.PutStr("host.name", "my-host")
	})

	p, err := buildTestProcessor(cfg)
	require.NoError(t, err)
	result, err := p.processLogs(context.Background(), ld)
	require.NoError(t, err)

	assert.Equal(t, 0, entity.ResourceEntityRefs(result.ResourceLogs().At(0).Resource()).Len())
}

// TestInsertLogs_NoInferenceFromNonIdKeyAttrs: host.name / service.name are NOT
// added to any EntityRef — only statically configured IDKeys are referenced.
func TestInsertLogs_NoInferenceFromNonIdKeyAttrs(t *testing.T) {
	cfg := &Config{
		Action: ActionInsert,
		EntityRefs: []EntityRefConfig{
			{Type: "KubernetesContainer", IDKeys: []string{"sw.k8s.cluster.uid", "k8s.namespace.name", "k8s.pod.name", "k8s.container.name"}},
		},
	}
	ld := buildLogs(func(m pcommon.Map) {
		m.PutStr("sw.k8s.cluster.uid", "cluster-1")
		m.PutStr("k8s.namespace.name", "default")
		m.PutStr("k8s.pod.name", "my-pod")
		m.PutStr("k8s.container.name", "app")
		m.PutStr("host.name", "node-1")
		m.PutStr("service.name", "my-service")
	})

	p, err := buildTestProcessor(cfg)
	require.NoError(t, err)
	result, err := p.processLogs(context.Background(), ld)
	require.NoError(t, err)

	refs := collectEntityRefs(result)
	require.Len(t, refs, 1)
	assert.Equal(t, "KubernetesContainer", refs[0].typ)
	// Verify host.name and service.name do not appear in any EntityRef id_keys
	for _, ref := range refs {
		for _, k := range ref.idKeys {
			assert.NotEqual(t, "host.name", k)
			assert.NotEqual(t, "service.name", k)
		}
	}
}

// TestInsertLogs_Idempotent_CaseInsensitive: a pre-existing EntityRef whose Type differs only in case
// must not cause a second EntityRef to be inserted.
func TestInsertLogs_Idempotent_CaseInsensitive(t *testing.T) {
	cfg := &Config{
		Action: ActionInsert,
		EntityRefs: []EntityRefConfig{
			{Type: "KubernetesPod", IDKeys: []string{"sw.k8s.cluster.uid", "k8s.pod.name"}},
		},
	}
	ld := buildLogs(func(m pcommon.Map) {
		m.PutStr("sw.k8s.cluster.uid", "cluster-1")
		m.PutStr("k8s.pod.name", "my-pod")
	})

	// Pre-populate an EntityRef with the same type but in uppercase to simulate
	// a case difference from a prior pipeline stage.
	resource := ld.ResourceLogs().At(0).Resource()
	existing := entity.ResourceEntityRefs(resource)
	r := existing.AppendEmpty()
	r.SetType("KUBERNETESPOD")
	r.IdKeys().Append("sw.k8s.cluster.uid")

	p, err := buildTestProcessor(cfg)
	require.NoError(t, err)
	result, err := p.processLogs(context.Background(), ld)
	require.NoError(t, err)

	refs := collectEntityRefs(result)
	require.Len(t, refs, 1, "case-insensitive duplicate must not be inserted")
	// The pre-existing ref (KUBERNETESPOD) is kept; no new one added.
	assert.Equal(t, "KUBERNETESPOD", refs[0].typ)
}

// TestInsertLogs_Idempotent: processing the same logs twice must not duplicate EntityRefs.
func TestInsertLogs_Idempotent(t *testing.T) {
	cfg := &Config{
		Action: ActionInsert,
		EntityRefs: []EntityRefConfig{
			{Type: "KubernetesPod", IDKeys: []string{"sw.k8s.cluster.uid", "k8s.pod.name"}},
		},
	}
	ld := buildLogs(func(m pcommon.Map) {
		m.PutStr("sw.k8s.cluster.uid", "cluster-1")
		m.PutStr("k8s.pod.name", "my-pod")
	})

	p, err := buildTestProcessor(cfg)
	require.NoError(t, err)
	result, err := p.processLogs(context.Background(), ld)
	require.NoError(t, err)

	// Process the same (already-annotated) logs a second time.
	result, err = p.processLogs(context.Background(), result)
	require.NoError(t, err)

	refs := collectEntityRefs(result)
	require.Len(t, refs, 1, "duplicate EntityRef must not be inserted on second pass")
	assert.Equal(t, "KubernetesPod", refs[0].typ)
}

// TestInsertLogs_MultipleResources: batch with 2 ResourceLogs — first has all IDKeys
// present (EntityRef inserted), second is missing one key (EntityRef skipped).
func TestInsertLogs_MultipleResources(t *testing.T) {
	cfg := &Config{
		Action: ActionInsert,
		EntityRefs: []EntityRefConfig{
			{Type: "KubernetesPod", IDKeys: []string{"sw.k8s.cluster.uid", "k8s.namespace.name", "k8s.pod.name"}},
		},
	}
	ld := plog.NewLogs()

	// Resource 0: all IDKeys present → EntityRef should be inserted.
	rl0 := ld.ResourceLogs().AppendEmpty()
	rl0.Resource().Attributes().PutStr("sw.k8s.cluster.uid", "cluster-1")
	rl0.Resource().Attributes().PutStr("k8s.namespace.name", "default")
	rl0.Resource().Attributes().PutStr("k8s.pod.name", "my-pod")

	// Resource 1: missing k8s.pod.name → EntityRef should NOT be inserted.
	rl1 := ld.ResourceLogs().AppendEmpty()
	rl1.Resource().Attributes().PutStr("sw.k8s.cluster.uid", "cluster-1")
	rl1.Resource().Attributes().PutStr("k8s.namespace.name", "default")

	p, err := buildTestProcessor(cfg)
	require.NoError(t, err)
	result, err := p.processLogs(context.Background(), ld)
	require.NoError(t, err)

	refs0 := entity.ResourceEntityRefs(result.ResourceLogs().At(0).Resource())
	require.Equal(t, 1, refs0.Len(), "resource 0: expected 1 EntityRef")
	assert.Equal(t, "KubernetesPod", refs0.At(0).Type())

	refs1 := entity.ResourceEntityRefs(result.ResourceLogs().At(1).Resource())
	assert.Equal(t, 0, refs1.Len(), "resource 1: expected 0 EntityRefs")
}

// TestInsertLogs_EmptyStringAttribute: an id_key attribute present with an empty string value
// must NOT cause an EntityRef to be inserted — empty-string identity keys are invalid.
func TestInsertLogs_EmptyStringAttribute(t *testing.T) {
	cfg := &Config{
		Action: ActionInsert,
		EntityRefs: []EntityRefConfig{
			{Type: "KubernetesPod", IDKeys: []string{"sw.k8s.cluster.uid", "k8s.pod.name"}},
		},
	}
	ld := buildLogs(func(m pcommon.Map) {
		m.PutStr("sw.k8s.cluster.uid", "cluster-1")
		m.PutStr("k8s.pod.name", "") // present but empty
	})

	p, err := buildTestProcessor(cfg)
	require.NoError(t, err)
	result, err := p.processLogs(context.Background(), ld)
	require.NoError(t, err)

	refs := collectEntityRefs(result)
	assert.Empty(t, refs, "empty-string id_key attribute must not produce an EntityRef")
}

// TestInsertLogs_NonStringAttribute: an id_key attribute present with a non-string type (int)
// is treated as absent — only string-typed attribute values satisfy id_key presence.
func TestInsertLogs_NonStringAttribute(t *testing.T) {
	cfg := &Config{
		Action: ActionInsert,
		EntityRefs: []EntityRefConfig{
			{Type: "KubernetesPod", IDKeys: []string{"sw.k8s.cluster.uid", "k8s.pod.name"}},
		},
	}
	ld := buildLogs(func(m pcommon.Map) {
		m.PutStr("sw.k8s.cluster.uid", "cluster-1")
		m.PutInt("k8s.pod.name", 42) // present but wrong type
	})

	p, err := buildTestProcessor(cfg)
	require.NoError(t, err)
	result, err := p.processLogs(context.Background(), ld)
	require.NoError(t, err)

	refs := collectEntityRefs(result)
	assert.Empty(t, refs, "non-string id_key attribute must not produce an EntityRef")
}

// TestInsertLogs_ReadOnly: a read-only batch must return errReadOnlyBatch, not panic.
func TestInsertLogs_ReadOnly(t *testing.T) {
	cfg := &Config{
		Action: ActionInsert,
		EntityRefs: []EntityRefConfig{
			{Type: "KubernetesPod", IDKeys: []string{"sw.k8s.cluster.uid", "k8s.pod.name"}},
		},
	}
	ld := buildLogs(func(m pcommon.Map) {
		m.PutStr("sw.k8s.cluster.uid", "cluster-1")
		m.PutStr("k8s.pod.name", "my-pod")
	})
	ld.MarkReadOnly()

	p, err := buildTestProcessor(cfg)
	require.NoError(t, err)
	_, err = p.processLogs(context.Background(), ld)
	require.ErrorIs(t, err, errReadOnlyBatch)
}

// TestInsert_EmptyBatch_AllSignals: empty signals (no resources) must return without error.
func TestInsert_EmptyBatch_AllSignals(t *testing.T) {
	cfg := &Config{
		Action: ActionInsert,
		EntityRefs: []EntityRefConfig{
			{Type: "KubernetesPod", IDKeys: []string{"sw.k8s.cluster.uid", "k8s.pod.name"}},
		},
	}
	p, err := buildTestProcessor(cfg)
	require.NoError(t, err)

	_, err = p.processLogs(context.Background(), plog.NewLogs())
	require.NoError(t, err)

	_, err = p.processMetrics(context.Background(), pmetric.NewMetrics())
	require.NoError(t, err)

	_, err = p.processTraces(context.Background(), ptrace.NewTraces())
	require.NoError(t, err)
}

// --- Metrics helpers and tests ---

func buildMetrics(setAttrs func(pcommon.Map)) pmetric.Metrics {
	md := pmetric.NewMetrics()
	rm := md.ResourceMetrics().AppendEmpty()
	setAttrs(rm.Resource().Attributes())
	return md
}

func collectMetricEntityRefs(md pmetric.Metrics) []struct {
	typ    string
	idKeys []string
} {
	if md.ResourceMetrics().Len() == 0 {
		return nil
	}
	return extractEntityRefs(md.ResourceMetrics().At(0).Resource())
}

func TestInsertMetrics_AllIDKeysPresent(t *testing.T) {
	cfg := &Config{
		Action: ActionInsert,
		EntityRefs: []EntityRefConfig{
			{Type: "KubernetesContainer", IDKeys: []string{"sw.k8s.cluster.uid", "k8s.namespace.name", "k8s.pod.name", "k8s.container.name"}},
		},
	}
	md := buildMetrics(func(m pcommon.Map) {
		m.PutStr("sw.k8s.cluster.uid", "cluster-1")
		m.PutStr("k8s.namespace.name", "default")
		m.PutStr("k8s.pod.name", "my-pod")
		m.PutStr("k8s.container.name", "my-container")
	})

	p, err := buildTestProcessor(cfg)
	require.NoError(t, err)
	result, err := p.processMetrics(context.Background(), md)
	require.NoError(t, err)

	refs := collectMetricEntityRefs(result)
	require.Len(t, refs, 1)
	assert.Equal(t, "KubernetesContainer", refs[0].typ)
	assert.Equal(t, []string{"sw.k8s.cluster.uid", "k8s.namespace.name", "k8s.pod.name", "k8s.container.name"}, refs[0].idKeys)
}

func TestInsertMetrics_OneIdKeyMissing(t *testing.T) {
	cfg := &Config{
		Action: ActionInsert,
		EntityRefs: []EntityRefConfig{
			{Type: "KubernetesContainer", IDKeys: []string{"sw.k8s.cluster.uid", "k8s.namespace.name", "k8s.pod.name", "k8s.container.name"}},
		},
	}
	md := buildMetrics(func(m pcommon.Map) {
		m.PutStr("sw.k8s.cluster.uid", "cluster-1")
		m.PutStr("k8s.namespace.name", "default")
		m.PutStr("k8s.pod.name", "my-pod")
		// k8s.container.name absent
	})

	p, err := buildTestProcessor(cfg)
	require.NoError(t, err)
	result, err := p.processMetrics(context.Background(), md)
	require.NoError(t, err)

	refs := collectMetricEntityRefs(result)
	assert.Empty(t, refs)
}

func TestInsertMetrics_MultipleEntries_PartialMatch(t *testing.T) {
	cfg := &Config{
		Action: ActionInsert,
		EntityRefs: []EntityRefConfig{
			{Type: "KubernetesContainer", IDKeys: []string{"sw.k8s.cluster.uid", "k8s.namespace.name", "k8s.pod.name", "k8s.container.name"}},
			{Type: "KubernetesPod", IDKeys: []string{"sw.k8s.cluster.uid", "k8s.namespace.name", "k8s.pod.name"}},
			{Type: "KubernetesNode", IDKeys: []string{"sw.k8s.cluster.uid", "k8s.node.name"}},
			{Type: "KubernetesDeployment", IDKeys: []string{"sw.k8s.cluster.uid", "k8s.namespace.name", "k8s.deployment.name"}},
		},
	}
	md := buildMetrics(func(m pcommon.Map) {
		m.PutStr("sw.k8s.cluster.uid", "cluster-1")
		m.PutStr("k8s.namespace.name", "default")
		m.PutStr("k8s.pod.name", "standalone-pod")
		m.PutStr("k8s.container.name", "app")
		m.PutStr("k8s.node.name", "node-1")
		// No workload-specific attributes
	})

	p, err := buildTestProcessor(cfg)
	require.NoError(t, err)
	result, err := p.processMetrics(context.Background(), md)
	require.NoError(t, err)

	refs := collectMetricEntityRefs(result)
	require.Len(t, refs, 3)
	types := make([]string, len(refs))
	for i, r := range refs {
		types[i] = r.typ
	}
	assert.Contains(t, types, "KubernetesContainer")
	assert.Contains(t, types, "KubernetesPod")
	assert.Contains(t, types, "KubernetesNode")
}

func TestInsertMetrics_EmptyConfig(t *testing.T) {
	cfg := &Config{
		Action:     ActionInsert,
		EntityRefs: nil,
	}
	md := buildMetrics(func(m pcommon.Map) {
		m.PutStr("sw.k8s.cluster.uid", "cluster-1")
	})

	p, err := buildTestProcessor(cfg)
	require.NoError(t, err)
	result, err := p.processMetrics(context.Background(), md)
	require.NoError(t, err)

	refs := collectMetricEntityRefs(result)
	assert.Empty(t, refs)
}

func TestInsertMetrics_EmptyStringAttribute(t *testing.T) {
	cfg := &Config{
		Action: ActionInsert,
		EntityRefs: []EntityRefConfig{
			{Type: "KubernetesPod", IDKeys: []string{"sw.k8s.cluster.uid", "k8s.pod.name"}},
		},
	}
	md := buildMetrics(func(m pcommon.Map) {
		m.PutStr("sw.k8s.cluster.uid", "cluster-1")
		m.PutStr("k8s.pod.name", "") // present but empty
	})

	p, err := buildTestProcessor(cfg)
	require.NoError(t, err)
	result, err := p.processMetrics(context.Background(), md)
	require.NoError(t, err)

	refs := collectMetricEntityRefs(result)
	assert.Empty(t, refs, "empty-string id_key attribute must not produce an EntityRef")
}

func TestRemoveMetrics_WithEntityRefs(t *testing.T) {
	cfg := &Config{Action: ActionRemove}
	md := buildMetrics(func(m pcommon.Map) {})

	resource := md.ResourceMetrics().At(0).Resource()
	refs := entity.ResourceEntityRefs(resource)
	r1 := refs.AppendEmpty()
	r1.SetType("KubernetesPod")
	r1.IdKeys().Append("sw.k8s.cluster.uid")
	r2 := refs.AppendEmpty()
	r2.SetType("KubernetesNode")
	r2.IdKeys().Append("sw.k8s.cluster.uid")

	require.Equal(t, 2, entity.ResourceEntityRefs(resource).Len())

	p, err := buildTestProcessor(cfg)
	require.NoError(t, err)
	result, err := p.processMetrics(context.Background(), md)
	require.NoError(t, err)

	assert.Equal(t, 0, entity.ResourceEntityRefs(result.ResourceMetrics().At(0).Resource()).Len())
}

// must not panic and must leave the Resource unchanged.
func TestRemoveMetrics_NoOp(t *testing.T) {
	cfg := &Config{Action: ActionRemove}
	md := buildMetrics(func(m pcommon.Map) {
		m.PutStr("host.name", "my-host")
	})

	p, err := buildTestProcessor(cfg)
	require.NoError(t, err)
	result, err := p.processMetrics(context.Background(), md)
	require.NoError(t, err)

	assert.Equal(t, 0, entity.ResourceEntityRefs(result.ResourceMetrics().At(0).Resource()).Len())
}

func TestInsertMetrics_ReadOnly(t *testing.T) {
	cfg := &Config{Action: ActionInsert, EntityRefs: []EntityRefConfig{
		{Type: "KubernetesPod", IDKeys: []string{"sw.k8s.cluster.uid", "k8s.pod.name"}},
	}}
	md := buildMetrics(func(m pcommon.Map) {})
	md.MarkReadOnly()

	p, err := buildTestProcessor(cfg)
	require.NoError(t, err)
	_, err = p.processMetrics(context.Background(), md)
	require.ErrorIs(t, err, errReadOnlyBatch)
}

func TestInsertMetrics_Idempotent(t *testing.T) {
	cfg := &Config{
		Action: ActionInsert,
		EntityRefs: []EntityRefConfig{
			{Type: "KubernetesPod", IDKeys: []string{"sw.k8s.cluster.uid", "k8s.pod.name"}},
		},
	}
	md := buildMetrics(func(m pcommon.Map) {
		m.PutStr("sw.k8s.cluster.uid", "cluster-1")
		m.PutStr("k8s.pod.name", "my-pod")
	})

	p, err := buildTestProcessor(cfg)
	require.NoError(t, err)
	result, err := p.processMetrics(context.Background(), md)
	require.NoError(t, err)
	result, err = p.processMetrics(context.Background(), result)
	require.NoError(t, err)

	refs := collectMetricEntityRefs(result)
	require.Len(t, refs, 1, "duplicate EntityRef must not be inserted on second pass")
	assert.Equal(t, "KubernetesPod", refs[0].typ)
}

// TestInsertMetrics_MultipleResources: batch with 2 ResourceMetrics — first has all IDKeys
// present (EntityRef inserted), second is missing one key (EntityRef skipped).
func TestInsertMetrics_MultipleResources(t *testing.T) {
	cfg := &Config{
		Action: ActionInsert,
		EntityRefs: []EntityRefConfig{
			{Type: "KubernetesPod", IDKeys: []string{"sw.k8s.cluster.uid", "k8s.namespace.name", "k8s.pod.name"}},
		},
	}
	md := pmetric.NewMetrics()
	rm0 := md.ResourceMetrics().AppendEmpty()
	rm0.Resource().Attributes().PutStr("sw.k8s.cluster.uid", "cluster-1")
	rm0.Resource().Attributes().PutStr("k8s.namespace.name", "default")
	rm0.Resource().Attributes().PutStr("k8s.pod.name", "my-pod")
	rm1 := md.ResourceMetrics().AppendEmpty()
	rm1.Resource().Attributes().PutStr("sw.k8s.cluster.uid", "cluster-1")
	rm1.Resource().Attributes().PutStr("k8s.namespace.name", "default")

	p, err := buildTestProcessor(cfg)
	require.NoError(t, err)
	result, err := p.processMetrics(context.Background(), md)
	require.NoError(t, err)

	refs0 := entity.ResourceEntityRefs(result.ResourceMetrics().At(0).Resource())
	require.Equal(t, 1, refs0.Len(), "resource 0: expected 1 EntityRef")
	assert.Equal(t, "KubernetesPod", refs0.At(0).Type())

	refs1 := entity.ResourceEntityRefs(result.ResourceMetrics().At(1).Resource())
	assert.Equal(t, 0, refs1.Len(), "resource 1: expected 0 EntityRefs")
}

// TestInsertMetrics_Idempotent_CaseInsensitive: a pre-existing metric EntityRef whose
// Type differs only in case must not cause a second EntityRef to be inserted.
func TestInsertMetrics_Idempotent_CaseInsensitive(t *testing.T) {
	cfg := &Config{
		Action: ActionInsert,
		EntityRefs: []EntityRefConfig{
			{Type: "KubernetesPod", IDKeys: []string{"sw.k8s.cluster.uid", "k8s.pod.name"}},
		},
	}
	md := buildMetrics(func(m pcommon.Map) {
		m.PutStr("sw.k8s.cluster.uid", "cluster-1")
		m.PutStr("k8s.pod.name", "my-pod")
	})

	resource := md.ResourceMetrics().At(0).Resource()
	existing := entity.ResourceEntityRefs(resource)
	r := existing.AppendEmpty()
	r.SetType("KUBERNETESPOD")
	r.IdKeys().Append("sw.k8s.cluster.uid")

	p, err := buildTestProcessor(cfg)
	require.NoError(t, err)
	result, err := p.processMetrics(context.Background(), md)
	require.NoError(t, err)

	refs := collectMetricEntityRefs(result)
	require.Len(t, refs, 1, "case-insensitive duplicate must not be inserted")
	assert.Equal(t, "KUBERNETESPOD", refs[0].typ)
}

// --- Traces helpers and tests ---

func buildTraces(setAttrs func(pcommon.Map)) ptrace.Traces {
	td := ptrace.NewTraces()
	rs := td.ResourceSpans().AppendEmpty()
	setAttrs(rs.Resource().Attributes())
	return td
}

func collectTraceEntityRefs(td ptrace.Traces) []struct {
	typ    string
	idKeys []string
} {
	if td.ResourceSpans().Len() == 0 {
		return nil
	}
	return extractEntityRefs(td.ResourceSpans().At(0).Resource())
}

func TestInsertTraces_AllIDKeysPresent(t *testing.T) {
	cfg := &Config{
		Action: ActionInsert,
		EntityRefs: []EntityRefConfig{
			{Type: "KubernetesContainer", IDKeys: []string{"sw.k8s.cluster.uid", "k8s.namespace.name", "k8s.pod.name", "k8s.container.name"}},
		},
	}
	td := buildTraces(func(m pcommon.Map) {
		m.PutStr("sw.k8s.cluster.uid", "cluster-1")
		m.PutStr("k8s.namespace.name", "default")
		m.PutStr("k8s.pod.name", "my-pod")
		m.PutStr("k8s.container.name", "my-container")
	})

	p, err := buildTestProcessor(cfg)
	require.NoError(t, err)
	result, err := p.processTraces(context.Background(), td)
	require.NoError(t, err)

	refs := collectTraceEntityRefs(result)
	require.Len(t, refs, 1)
	assert.Equal(t, "KubernetesContainer", refs[0].typ)
	assert.Equal(t, []string{"sw.k8s.cluster.uid", "k8s.namespace.name", "k8s.pod.name", "k8s.container.name"}, refs[0].idKeys)
}

func TestInsertTraces_OneIdKeyMissing(t *testing.T) {
	cfg := &Config{
		Action: ActionInsert,
		EntityRefs: []EntityRefConfig{
			{Type: "KubernetesContainer", IDKeys: []string{"sw.k8s.cluster.uid", "k8s.namespace.name", "k8s.pod.name", "k8s.container.name"}},
		},
	}
	td := buildTraces(func(m pcommon.Map) {
		m.PutStr("sw.k8s.cluster.uid", "cluster-1")
		m.PutStr("k8s.namespace.name", "default")
		m.PutStr("k8s.pod.name", "my-pod")
		// k8s.container.name absent
	})

	p, err := buildTestProcessor(cfg)
	require.NoError(t, err)
	result, err := p.processTraces(context.Background(), td)
	require.NoError(t, err)

	refs := collectTraceEntityRefs(result)
	assert.Empty(t, refs)
}

func TestInsertTraces_MultipleEntries_PartialMatch(t *testing.T) {
	cfg := &Config{
		Action: ActionInsert,
		EntityRefs: []EntityRefConfig{
			{Type: "KubernetesContainer", IDKeys: []string{"sw.k8s.cluster.uid", "k8s.namespace.name", "k8s.pod.name", "k8s.container.name"}},
			{Type: "KubernetesPod", IDKeys: []string{"sw.k8s.cluster.uid", "k8s.namespace.name", "k8s.pod.name"}},
			{Type: "KubernetesNode", IDKeys: []string{"sw.k8s.cluster.uid", "k8s.node.name"}},
			{Type: "KubernetesDeployment", IDKeys: []string{"sw.k8s.cluster.uid", "k8s.namespace.name", "k8s.deployment.name"}},
		},
	}
	td := buildTraces(func(m pcommon.Map) {
		m.PutStr("sw.k8s.cluster.uid", "cluster-1")
		m.PutStr("k8s.namespace.name", "default")
		m.PutStr("k8s.pod.name", "standalone-pod")
		m.PutStr("k8s.container.name", "app")
		m.PutStr("k8s.node.name", "node-1")
		// No workload-specific attributes
	})

	p, err := buildTestProcessor(cfg)
	require.NoError(t, err)
	result, err := p.processTraces(context.Background(), td)
	require.NoError(t, err)

	refs := collectTraceEntityRefs(result)
	require.Len(t, refs, 3)
	types := make([]string, len(refs))
	for i, r := range refs {
		types[i] = r.typ
	}
	assert.Contains(t, types, "KubernetesContainer")
	assert.Contains(t, types, "KubernetesPod")
	assert.Contains(t, types, "KubernetesNode")
}

func TestInsertTraces_EmptyConfig(t *testing.T) {
	cfg := &Config{
		Action:     ActionInsert,
		EntityRefs: nil,
	}
	td := buildTraces(func(m pcommon.Map) {
		m.PutStr("sw.k8s.cluster.uid", "cluster-1")
	})

	p, err := buildTestProcessor(cfg)
	require.NoError(t, err)
	result, err := p.processTraces(context.Background(), td)
	require.NoError(t, err)

	refs := collectTraceEntityRefs(result)
	assert.Empty(t, refs)
}

func TestInsertTraces_EmptyStringAttribute(t *testing.T) {
	cfg := &Config{
		Action: ActionInsert,
		EntityRefs: []EntityRefConfig{
			{Type: "KubernetesPod", IDKeys: []string{"sw.k8s.cluster.uid", "k8s.pod.name"}},
		},
	}
	td := buildTraces(func(m pcommon.Map) {
		m.PutStr("sw.k8s.cluster.uid", "cluster-1")
		m.PutStr("k8s.pod.name", "") // present but empty
	})

	p, err := buildTestProcessor(cfg)
	require.NoError(t, err)
	result, err := p.processTraces(context.Background(), td)
	require.NoError(t, err)

	refs := collectTraceEntityRefs(result)
	assert.Empty(t, refs, "empty-string id_key attribute must not produce an EntityRef")
}

func TestRemoveTraces_WithEntityRefs(t *testing.T) {
	cfg := &Config{Action: ActionRemove}
	td := buildTraces(func(m pcommon.Map) {})

	resource := td.ResourceSpans().At(0).Resource()
	refs := entity.ResourceEntityRefs(resource)
	r1 := refs.AppendEmpty()
	r1.SetType("KubernetesPod")
	r1.IdKeys().Append("sw.k8s.cluster.uid")
	r2 := refs.AppendEmpty()
	r2.SetType("KubernetesNode")
	r2.IdKeys().Append("sw.k8s.cluster.uid")

	require.Equal(t, 2, entity.ResourceEntityRefs(resource).Len())

	p, err := buildTestProcessor(cfg)
	require.NoError(t, err)
	result, err := p.processTraces(context.Background(), td)
	require.NoError(t, err)

	assert.Equal(t, 0, entity.ResourceEntityRefs(result.ResourceSpans().At(0).Resource()).Len())
}

// must not panic and must leave the Resource unchanged.
func TestRemoveTraces_NoOp(t *testing.T) {
	cfg := &Config{Action: ActionRemove}
	td := buildTraces(func(m pcommon.Map) {
		m.PutStr("host.name", "my-host")
	})

	p, err := buildTestProcessor(cfg)
	require.NoError(t, err)
	result, err := p.processTraces(context.Background(), td)
	require.NoError(t, err)

	assert.Equal(t, 0, entity.ResourceEntityRefs(result.ResourceSpans().At(0).Resource()).Len())
}

func TestInsertTraces_ReadOnly(t *testing.T) {
	cfg := &Config{Action: ActionInsert, EntityRefs: []EntityRefConfig{
		{Type: "KubernetesPod", IDKeys: []string{"sw.k8s.cluster.uid", "k8s.pod.name"}},
	}}
	td := buildTraces(func(m pcommon.Map) {})
	td.MarkReadOnly()

	p, err := buildTestProcessor(cfg)
	require.NoError(t, err)
	_, err = p.processTraces(context.Background(), td)
	require.ErrorIs(t, err, errReadOnlyBatch)
}

func TestInsertTraces_Idempotent(t *testing.T) {
	cfg := &Config{
		Action: ActionInsert,
		EntityRefs: []EntityRefConfig{
			{Type: "KubernetesPod", IDKeys: []string{"sw.k8s.cluster.uid", "k8s.pod.name"}},
		},
	}
	td := buildTraces(func(m pcommon.Map) {
		m.PutStr("sw.k8s.cluster.uid", "cluster-1")
		m.PutStr("k8s.pod.name", "my-pod")
	})

	p, err := buildTestProcessor(cfg)
	require.NoError(t, err)
	result, err := p.processTraces(context.Background(), td)
	require.NoError(t, err)
	result, err = p.processTraces(context.Background(), result)
	require.NoError(t, err)

	refs := collectTraceEntityRefs(result)
	require.Len(t, refs, 1, "duplicate EntityRef must not be inserted on second pass")
	assert.Equal(t, "KubernetesPod", refs[0].typ)
}

// TestInsertTraces_MultipleResources: batch with 2 ResourceSpans — first has all IDKeys
// present (EntityRef inserted), second is missing one key (EntityRef skipped).
func TestInsertTraces_MultipleResources(t *testing.T) {
	cfg := &Config{
		Action: ActionInsert,
		EntityRefs: []EntityRefConfig{
			{Type: "KubernetesPod", IDKeys: []string{"sw.k8s.cluster.uid", "k8s.namespace.name", "k8s.pod.name"}},
		},
	}
	td := ptrace.NewTraces()
	rs0 := td.ResourceSpans().AppendEmpty()
	rs0.Resource().Attributes().PutStr("sw.k8s.cluster.uid", "cluster-1")
	rs0.Resource().Attributes().PutStr("k8s.namespace.name", "default")
	rs0.Resource().Attributes().PutStr("k8s.pod.name", "my-pod")
	rs1 := td.ResourceSpans().AppendEmpty()
	rs1.Resource().Attributes().PutStr("sw.k8s.cluster.uid", "cluster-1")
	rs1.Resource().Attributes().PutStr("k8s.namespace.name", "default")

	p, err := buildTestProcessor(cfg)
	require.NoError(t, err)
	result, err := p.processTraces(context.Background(), td)
	require.NoError(t, err)

	refs0 := entity.ResourceEntityRefs(result.ResourceSpans().At(0).Resource())
	require.Equal(t, 1, refs0.Len(), "resource 0: expected 1 EntityRef")
	assert.Equal(t, "KubernetesPod", refs0.At(0).Type())

	refs1 := entity.ResourceEntityRefs(result.ResourceSpans().At(1).Resource())
	assert.Equal(t, 0, refs1.Len(), "resource 1: expected 0 EntityRefs")
}

// TestInsertTraces_Idempotent_CaseInsensitive: a pre-existing trace EntityRef whose
// Type differs only in case must not cause a second EntityRef to be inserted.
func TestInsertTraces_Idempotent_CaseInsensitive(t *testing.T) {
	cfg := &Config{
		Action: ActionInsert,
		EntityRefs: []EntityRefConfig{
			{Type: "KubernetesPod", IDKeys: []string{"sw.k8s.cluster.uid", "k8s.pod.name"}},
		},
	}
	td := buildTraces(func(m pcommon.Map) {
		m.PutStr("sw.k8s.cluster.uid", "cluster-1")
		m.PutStr("k8s.pod.name", "my-pod")
	})

	resource := td.ResourceSpans().At(0).Resource()
	existing := entity.ResourceEntityRefs(resource)
	r := existing.AppendEmpty()
	r.SetType("KUBERNETESPOD")
	r.IdKeys().Append("sw.k8s.cluster.uid")

	p, err := buildTestProcessor(cfg)
	require.NoError(t, err)
	result, err := p.processTraces(context.Background(), td)
	require.NoError(t, err)

	refs := collectTraceEntityRefs(result)
	require.Len(t, refs, 1, "case-insensitive duplicate must not be inserted")
	assert.Equal(t, "KUBERNETESPOD", refs[0].typ)
}
