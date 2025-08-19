package solarwindsprocessor

import (
	"fmt"

	"github.com/solarwinds/solarwinds-otel-collector-contrib/extension/solarwindsextension"
	"github.com/solarwinds/solarwinds-otel-collector-contrib/pkg/extensionfinder"
	"go.opentelemetry.io/collector/component"
	"go.uber.org/zap"
)

type ExtensionProvider interface {
	SetExtension(logger *zap.Logger, extensionName string, host component.Host) (*solarwindsextension.SolarwindsExtension, error)
	SetAttributes(resourceAttributes *map[string]string) error
}

type extensionProvider struct {
	extension *solarwindsextension.SolarwindsExtension
}

var _ ExtensionProvider = (*extensionProvider)(nil)

// NewExtensionProvider creates a new instance of ExtensionProvider.
func NewExtensionProvider() ExtensionProvider {
	return &extensionProvider{}
}

func (ep *extensionProvider) SetExtension(logger *zap.Logger, extensionName string, host component.Host) (*solarwindsextension.SolarwindsExtension, error) {
	if ep.extension == nil {
		swiExtension, err := extensionfinder.FindExtension[*solarwindsextension.SolarwindsExtension](
			logger,
			extensionName,
			host,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to find Solarwinds Extension %q", extensionName)
		}

		ep.extension = swiExtension
	}

	return ep.extension, nil
}

func (ep *extensionProvider) SetAttributes(resourceAttributes *map[string]string) error {
	if ep.extension == nil {
		return fmt.Errorf("solarwinds extension is nil")
	}

	cfg := ep.extension.GetCommonConfig()
	if cfg == nil {
		return fmt.Errorf("failed to get config from solarwinds extension")
	}

	attrs := *resourceAttributes
	// Send entity creation flag if the extension is configured to infer collector entity.
	if !cfg.WithoutEntity() {
		attrs[solarwindsextension.EntityCreation] = solarwindsextension.EntityCreationValue
	}

	// Set collector name.
	attrs[solarwindsextension.CollectorNameAttribute] = cfg.CollectorName()

	return nil
}
