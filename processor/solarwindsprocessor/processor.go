package solarwindsprocessor

import (
	"context"
	"fmt"

	"github.com/solarwinds/solarwinds-otel-collector-contrib/extension/solarwindsextension"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/pdata/plog/plogotlp"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/pdata/pmetric/pmetricotlp"
	"go.opentelemetry.io/collector/pdata/ptrace"
	"go.opentelemetry.io/collector/pdata/ptrace/ptraceotlp"
	"go.uber.org/zap"
)

type solarwindsprocessor struct {
	logger *zap.Logger
	cfg    *Config
}

func (p *solarwindsprocessor) start(ctx context.Context, host component.Host) error {
	swiExtension, err := findControllingExtension[*solarwindsextension.SolarwindsExtension](
		p.logger,
		p.cfg.ExtensionName,
		host,
	)
	if err != nil {
		msg := fmt.Sprintf("failed to find Solarwinds Extension %q", p.cfg.ExtensionName)
		p.logger.Error(msg, zap.Error(err))
		return fmt.Errorf("%s: %w", msg, err)
	}

	// Adjust configuration by fetching values from the extension.
	p.adjustConfigurationByExtension(swiExtension)

	return nil
}

func (p *solarwindsprocessor) adjustConfigurationByExtension(
	swiExtension *solarwindsextension.SolarwindsExtension,
) {
	extCfg := swiExtension.GetCommonConfig()

	// Without entity flag.
	if extCfg.WithoutEntity() {
		p.cfg.ResourceAttributes[solarwindsextension.EntityCreation] = solarwindsextension.EntityCreationValue
	}

	// Collector name.
	p.cfg.ResourceAttributes[solarwindsextension.CollectorNameAttribute] = extCfg.CollectorName()
	return
}

// TODO: to be exported out.
func findControllingExtension[E any](
	l *zap.Logger,
	extensionName string,
	host component.Host,
) (E, error) {
	extID := new(component.ID)
	if err := extID.UnmarshalText([]byte(extensionName)); err != nil {
		msg := fmt.Sprintf("failed to parse extension ID %q", extensionName)
		l.Error(msg, zap.Error(err))
		return *new(E), fmt.Errorf("%s: %w", msg, err)
	}

	ext, found := host.GetExtensions()[*extID]
	if !found {
		msg := fmt.Sprintf("extension %q not found", extensionName)
		l.Error(msg)
		return *new(E), fmt.Errorf("%s", msg)
	}

	castedExtension, castedOK := ext.(E)
	if !castedOK {
		msg := fmt.Sprintf("extension %q is not a %T", extensionName, (*E)(nil))
		l.Error(msg)
		return *new(E), fmt.Errorf("%s", msg)
	}

	return castedExtension, nil
}

// TODO: to be exported out.
type Resource interface {
	Resource() pcommon.Resource
}

type ResourceCollection[T Resource] interface {
	At(index int) T
	Len() int
}

func decorateResourceAttributes[T Resource](collection ResourceCollection[T], atts map[string]string) {
	collectionLen := collection.Len()
	if collectionLen == 0 {
		return
	}

	for i := 0; i < collectionLen; i++ {
		resource := collection.At(i).Resource()
		resourceAttributes := resource.Attributes()
		for key, value := range atts {
			resourceAttributes.PutStr(key, value)
		}
	}
}

// End of to be exported out.

// TODO: to be exported out.

func notifySignalSizeLimitExceeded(signal any, limit int, logger *zap.Logger) error {
	if limit <= 0 {
		return nil // No limit set, skip the check.
	}

	var err error
	var bs []byte
	var signalName string

	switch v := signal.(type) {
	case plog.Logs:
		signalName = "Logs"
		er := plogotlp.NewExportRequestFromLogs(v)
		bs, err = er.MarshalProto()
	case pmetric.Metrics:
		signalName = "Metrics"
		er := pmetricotlp.NewExportRequestFromMetrics(v)
		bs, err = er.MarshalProto()
	case ptrace.Traces:
		signalName = "Traces"
		er := ptraceotlp.NewExportRequestFromTraces(v)
		bs, err = er.MarshalProto()
	default:
		return fmt.Errorf("unsupported signal type: %T", v)
	}

	if err != nil {
		return fmt.Errorf("failed to marshal signal: %w", err)
	}

	bsLen := len(bs)
	if bsLen > limit {
		msg := fmt.Sprintf(
			"%s size %d bytes exceeds the limit of %d MiB",
			signalName, bsLen, limit,
		)
		logger.Warn(msg, zap.Int("size_bytes", bsLen), zap.Int("limit_mib", limit))
	}

	return nil
}

// End of to be exported out.

func (p *solarwindsprocessor) processLogs(
	ctx context.Context,
	logs plog.Logs,
) (plog.Logs, error) {
	decorateResourceAttributes(logs.ResourceLogs(), p.cfg.ResourceAttributes)
	err := notifySignalSizeLimitExceeded(logs, p.cfg.MaxSizeMib, p.logger)
	if err != nil {
		msg := fmt.Sprintf("failed to notify logs size limit exceeded")
		p.logger.Error(msg, zap.Error(err))
		return plog.Logs{}, fmt.Errorf("%s: %w", msg, err)
	}
	return logs, nil
}

func (p *solarwindsprocessor) processMetrics(
	ctx context.Context,
	metrics pmetric.Metrics,
) (pmetric.Metrics, error) {
	decorateResourceAttributes(metrics.ResourceMetrics(), p.cfg.ResourceAttributes)
	err := notifySignalSizeLimitExceeded(metrics, p.cfg.MaxSizeMib, p.logger)
	if err != nil {
		msg := fmt.Sprintf("failed to notify metrics size limit exceeded")
		p.logger.Error(msg, zap.Error(err))
		return pmetric.Metrics{}, fmt.Errorf("%s: %w", msg, err)
	}
	return metrics, nil
}

func (p *solarwindsprocessor) processTraces(
	ctx context.Context,
	traces ptrace.Traces,
) (ptrace.Traces, error) {
	decorateResourceAttributes(traces.ResourceSpans(), p.cfg.ResourceAttributes)
	err := notifySignalSizeLimitExceeded(traces, p.cfg.MaxSizeMib, p.logger)
	if err != nil {
		msg := fmt.Sprintf("failed to notify traces size limit exceeded")
		p.logger.Error(msg, zap.Error(err))
		return ptrace.Traces{}, fmt.Errorf("%s: %w", msg, err)
	}
	return traces, nil
}

func createProcessor(logger *zap.Logger, cfg component.Config) (*solarwindsprocessor, error) {
	c, err := checkConfig(cfg)
	if err != nil {
		return nil, err
	}
	return newProcessor(logger, c), nil
}

func checkConfig(cfg component.Config) (*Config, error) {
	c, ok := cfg.(*Config)
	if !ok {
		return nil, fmt.Errorf("invalid configuration type: %T, expected *solarwindsprocessor.Config", cfg)
	}
	if err := c.Validate(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}
	return c, nil
}

func newProcessor(logger *zap.Logger, cfg *Config) *solarwindsprocessor {
	return &solarwindsprocessor{
		logger: logger,
		cfg:    cfg,
	}
}
