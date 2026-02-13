package componentcommunication

type Config struct {
	// ID of OpAMP extension to send events to.
	OpAMPExtensionID string `mapstructure:"opamp_extension"`
	// List of custom capabilities.
	Capabilities []string `mapstructure:"custom_capabilities"`
}
