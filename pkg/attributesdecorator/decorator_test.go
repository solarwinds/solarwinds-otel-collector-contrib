package attributesdecorator

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/pdata/ptrace"
)

func TestDecorateResourceAttributesForMetrics(t *testing.T) {
	additionalAttributes := map[string]string{
		"service.name": "test-service",
		"environment":  "test-env",
	}

	ms := pmetric.NewMetrics()
	ms.ResourceMetrics().AppendEmpty().Resource().Attributes().PutStr("existing.attribute", "test-value")

	DecorateResourceAttributes(ms.ResourceMetrics(), additionalAttributes)

	expectedAttributes := pcommon.NewMap()
	err := expectedAttributes.FromRaw(map[string]any{
		"existing.attribute": "test-value",
		"service.name":       "test-service",
		"environment":        "test-env",
	})
	assert.NoError(t, err)

	actualAttributes := ms.ResourceMetrics().At(0).Resource().Attributes()
	assert.Equal(t, expectedAttributes, actualAttributes)
}

func TestDecorateResourceAttributesForLogs(t *testing.T) {
	additionalAttributes := map[string]string{
		"service.name": "test-service",
		"environment":  "test-env",
	}

	ls := plog.NewLogs()
	ls.ResourceLogs().AppendEmpty().Resource().Attributes().PutStr("existing.attribute", "test-value")

	DecorateResourceAttributes(ls.ResourceLogs(), additionalAttributes)

	expectedAttributes := pcommon.NewMap()
	err := expectedAttributes.FromRaw(map[string]any{
		"existing.attribute": "test-value",
		"service.name":       "test-service",
		"environment":        "test-env",
	})
	assert.NoError(t, err)

	actualAttributes := ls.ResourceLogs().At(0).Resource().Attributes()
	assert.Equal(t, expectedAttributes, actualAttributes)
}

func TestDecorateResourceAttributesForTraces(t *testing.T) {
	additionalAttributes := map[string]string{
		"service.name": "test-service",
		"environment":  "test-env",
	}

	ts := ptrace.NewTraces()
	ts.ResourceSpans().AppendEmpty().Resource().Attributes().PutStr("existing.attribute", "test-value")

	DecorateResourceAttributes(ts.ResourceSpans(), additionalAttributes)

	expectedAttributes := pcommon.NewMap()
	err := expectedAttributes.FromRaw(map[string]any{
		"existing.attribute": "test-value",
		"service.name":       "test-service",
		"environment":        "test-env",
	})
	assert.NoError(t, err)

	actualAttributes := ts.ResourceSpans().At(0).Resource().Attributes()
	assert.Equal(t, expectedAttributes, actualAttributes)
}
