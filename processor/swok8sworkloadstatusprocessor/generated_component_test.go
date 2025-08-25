package swok8sworkloadstatusprocessor

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/consumer/consumertest"
	"go.opentelemetry.io/collector/processor/processortest"
)

func TestGeneratedComponent(t *testing.T) {
	factory := NewFactory()
	procType := component.MustNewType("swok8sworkloadstatus")

	t.Run("CreateDefaultConfig", func(t *testing.T) {
		cfg := factory.CreateDefaultConfig()
		require.NotNil(t, cfg)
		require.NoError(t, componenttest.CheckConfigStruct(cfg))
	})

	t.Run("CreateLogsProcessor", func(t *testing.T) {
		cfg := factory.CreateDefaultConfig()
		processor, err := factory.CreateLogs(context.Background(), processortest.NewNopSettings(procType), cfg, consumertest.NewNop())
		require.NoError(t, err)
		require.NotNil(t, processor)
	})
}
