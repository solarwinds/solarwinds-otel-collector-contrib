package internal

type HostDecoration struct {
	Enabled  bool   `mapstructure:"enabled"`
	ClientId string `mapstructure:"client_id,omitempty"`
}
