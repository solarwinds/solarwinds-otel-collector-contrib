package mqttreceiver

type Metric struct {
	Name         string `mapstructure:"name"`
	Type         string `mapstructure:"type"`
	Topic        string `mapstructure:"topic"`
	Unit         string `mapstructure:"unit"`
	Desc         string `mapstructure:"description"`
	JsonProperty string `mapstructure:"jsonProperty"`
}

type Sensor struct {
	Name     string    `mapstructure:"name"`
	Category string    `mapstructure:"category"`
	Metrics  []*Metric `mapstructure:"metrics"`
}

type Broker struct {
	Name     string    `mapstructure:"name"`
	Protocol string    `mapstructure:"protocol"`
	Server   string    `mapstructure:"server"`
	Port     int       `mapstructure:"port"`
	User     string    `mapstructure:"user"`
	Password string    `mapstructure:"password"`
	Sensors  []*Sensor `mapstructure:"sensors"`
}

type Config struct {
	Brokers []*Broker `mapstructure:"brokers"`
}

func (c *Config) Validate() error {
	return nil
}
