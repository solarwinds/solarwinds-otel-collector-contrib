package mqttreceiver

import (
	"context"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/receiver"
)

// NewFactory creates a factory for the MQTT receiver.
// Serves as entry point for the receiver.
func NewFactory() receiver.Factory {
	return receiver.NewFactory(
		component.MustNewType("mqtt"),
		createDefaultConfig,
		receiver.WithMetrics(
			createMetricsReceiver,
			component.StabilityLevelStable,
		),
	)
}

func createDefaultConfig() component.Config {
	return &Config{}
}

func createMetricsReceiver(
	ctx context.Context,
	settings receiver.Settings,
	config component.Config,
	metrics consumer.Metrics,
) (receiver.Metrics, error) {
	settings.Logger.Info("Creating MQTT receiver")
	cfx := config.(*Config)
	return newReceiver(settings, cfx, ctx, metrics)
}
