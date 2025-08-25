package internal

type HostDecoration struct {
	Enabled        bool   `mapstructure:"enabled"`
	FallbackHostID string `mapstructure:"fallback_host_id,omitempty"`
}
