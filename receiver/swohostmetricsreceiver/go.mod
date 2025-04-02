module github.com/solarwinds/solarwinds-otel-collector-releases/receiver/swohostmetricsreceiver

go 1.24.2

require (
	github.com/go-ole/go-ole v1.3.0
	github.com/google/go-cmp v0.7.0
	github.com/shirou/gopsutil/v3 v3.24.5
	github.com/solarwinds/solarwinds-otel-collector-releases/pkg/testutil v0.119.11
	github.com/solarwinds/solarwinds-otel-collector-releases/pkg/version v0.119.11
	github.com/stretchr/testify v1.10.0
	github.com/yusufpapurcu/wmi v1.2.4
	go.opentelemetry.io/collector/component v0.119.0
	go.opentelemetry.io/collector/component/componenttest v0.119.0
	go.opentelemetry.io/collector/confmap v1.25.0
	go.opentelemetry.io/collector/consumer v1.25.0
	go.opentelemetry.io/collector/consumer/consumertest v0.119.0
	go.opentelemetry.io/collector/pdata v1.25.0
	go.opentelemetry.io/collector/receiver v0.119.0
	go.opentelemetry.io/collector/receiver/receivertest v0.119.0
	go.opentelemetry.io/collector/scraper/scrapertest v0.119.0
	go.opentelemetry.io/collector/semconv v0.119.0
	go.opentelemetry.io/otel v1.34.0
	go.uber.org/goleak v1.3.0
	go.uber.org/zap v1.27.0
	golang.org/x/sys v0.31.0
)

require (
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/go-logr/logr v1.4.2 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-viper/mapstructure/v2 v2.2.1 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/knadh/koanf/maps v0.1.1 // indirect
	github.com/knadh/koanf/providers/confmap v0.1.0 // indirect
	github.com/knadh/koanf/v2 v2.1.2 // indirect
	github.com/lufia/plan9stats v0.0.0-20240909124753-873cd0166683 // indirect
	github.com/mitchellh/copystructure v1.2.0 // indirect
	github.com/mitchellh/reflectwalk v1.0.2 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/power-devops/perfstat v0.0.0-20240221224432-82ca36839d55 // indirect
	github.com/shoenig/go-m1cpu v0.1.6 // indirect
	github.com/tklauser/go-sysconf v0.3.14 // indirect
	github.com/tklauser/numcpus v0.9.0 // indirect
	go.opentelemetry.io/auto/sdk v1.1.0 // indirect
	go.opentelemetry.io/collector/config/configtelemetry v0.119.0 // indirect
	go.opentelemetry.io/collector/consumer/consumererror v0.119.0 // indirect
	go.opentelemetry.io/collector/consumer/xconsumer v0.119.0 // indirect
	go.opentelemetry.io/collector/pdata/pprofile v0.119.0 // indirect
	go.opentelemetry.io/collector/pipeline v0.119.0 // indirect
	go.opentelemetry.io/collector/receiver/xreceiver v0.119.0 // indirect
	go.opentelemetry.io/collector/scraper v0.119.0
	go.opentelemetry.io/collector/scraper/scraperhelper v0.119.0
	go.opentelemetry.io/otel/metric v1.34.0 // indirect
	go.opentelemetry.io/otel/sdk v1.34.0 // indirect
	go.opentelemetry.io/otel/sdk/metric v1.34.0 // indirect
	go.opentelemetry.io/otel/trace v1.34.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/net v0.36.0 // indirect
	golang.org/x/text v0.22.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20241202173237-19429a94021a // indirect
	google.golang.org/grpc v1.70.0 // indirect
	google.golang.org/protobuf v1.36.4 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/solarwinds/solarwinds-otel-collector-releases/pkg/version => ../../pkg/version

replace github.com/solarwinds/solarwinds-otel-collector-releases/pkg/testutil => ../../pkg/testutil
