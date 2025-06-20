package extensionfinder

import (
	"fmt"

	"go.opentelemetry.io/collector/component"
	"go.uber.org/zap"
)

func FindExtension[E any](
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
