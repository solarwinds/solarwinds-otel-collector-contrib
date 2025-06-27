package solarwindsprocessor

import "fmt"

type Config struct {
	// Extension identifies a Solarwinds Extension to
	// use for obtaining required configuration.
	ExtensionName string `mapstructure:"extension"`
	// Limit for maximum size (in MiB) of the serialized signal payload.
	// When maximum size is set greater to zero, limit check on serialized content is
	// performed. When the serialized content exceeds the limit, it is reported to log.
	// When maximum size is set to zero, no limit check is performed.
	MaxSizeMib int `mapstructure:"max_size_mib,omitempty"`
	// Resource attributes to be added to the processed signals.
	ResourceAttributes map[string]string `mapstructure:"resource,omitempty"`
}

func (c *Config) Validate() error {
	if c.ExtensionName == "" {
		return fmt.Errorf("%s", "invalid configuration: 'extension' must be set")
	}
	if c.MaxSizeMib < 0 {
		return fmt.Errorf("%s: %d", "invalid configuration: 'max_size_mib' must be greater than or equal to zero", c.MaxSizeMib)
	}
	return nil
}
