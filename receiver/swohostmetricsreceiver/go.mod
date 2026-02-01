module github.com/solarwinds/solarwinds-otel-collector-contrib/receiver/swohostmetricsreceiver

go 1.25.6

require (
	github.com/go-ole/go-ole v1.3.0
	github.com/google/go-cmp v0.7.0
	github.com/shirou/gopsutil/v4 v4.26.1
	github.com/solarwinds/solarwinds-otel-collector-contrib/pkg/registry v0.140.8
	github.com/solarwinds/solarwinds-otel-collector-contrib/pkg/testutil v0.140.8
	github.com/solarwinds/solarwinds-otel-collector-contrib/pkg/version v0.140.8
	github.com/solarwinds/solarwinds-otel-collector-contrib/pkg/wmi v0.140.8
	github.com/stretchr/testify v1.11.1
	go.opentelemetry.io/collector/component v1.46.0
	go.opentelemetry.io/collector/component/componenttest v0.140.0
	go.opentelemetry.io/collector/confmap v1.46.0
	go.opentelemetry.io/collector/consumer v1.46.0
	go.opentelemetry.io/collector/consumer/consumertest v0.140.0
	go.opentelemetry.io/collector/pdata v1.46.0
	go.opentelemetry.io/collector/receiver v1.46.0
	go.opentelemetry.io/collector/receiver/receivertest v0.140.0
	go.opentelemetry.io/collector/scraper/scrapertest v0.140.0
	go.opentelemetry.io/otel v1.38.0
	go.uber.org/goleak v1.3.0
	go.uber.org/zap v1.27.1
)

require (
	github.com/ebitengine/purego v0.9.1 // indirect
	github.com/gobwas/glob v0.2.3 // indirect
	github.com/hashicorp/go-version v1.7.0 // indirect
	github.com/yusufpapurcu/wmi v1.2.4 // indirect
	go.opentelemetry.io/collector/featuregate v1.46.0 // indirect
	go.opentelemetry.io/collector/receiver/receiverhelper v0.140.0 // indirect
	go.yaml.in/yaml/v3 v3.0.4 // indirect
	golang.org/x/sys v0.40.0 // indirect
)

require (
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/go-logr/logr v1.4.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-viper/mapstructure/v2 v2.4.0 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/knadh/koanf/maps v0.1.2 // indirect
	github.com/knadh/koanf/providers/confmap v1.0.0 // indirect
	github.com/knadh/koanf/v2 v2.3.0 // indirect
	github.com/lufia/plan9stats v0.0.0-20250317134145-8bc96cf8fc35 // indirect
	github.com/mitchellh/copystructure v1.2.0 // indirect
	github.com/mitchellh/reflectwalk v1.0.2 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.3-0.20250322232337-35a7c28c31ee // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/power-devops/perfstat v0.0.0-20240221224432-82ca36839d55 // indirect
	github.com/tklauser/go-sysconf v0.3.16 // indirect
	github.com/tklauser/numcpus v0.11.0 // indirect
	go.opentelemetry.io/auto/sdk v1.2.1 // indirect
	go.opentelemetry.io/collector/consumer/consumererror v0.140.0 // indirect
	go.opentelemetry.io/collector/consumer/xconsumer v0.140.0 // indirect
	go.opentelemetry.io/collector/pdata/pprofile v0.140.0 // indirect
	go.opentelemetry.io/collector/pipeline v1.46.0 // indirect
	go.opentelemetry.io/collector/receiver/xreceiver v0.140.0 // indirect
	go.opentelemetry.io/collector/scraper v0.140.0
	go.opentelemetry.io/collector/scraper/scraperhelper v0.140.0
	go.opentelemetry.io/otel/metric v1.38.0 // indirect
	go.opentelemetry.io/otel/sdk v1.38.0 // indirect
	go.opentelemetry.io/otel/sdk/metric v1.38.0 // indirect
	go.opentelemetry.io/otel/trace v1.38.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250804133106-a7a43d27e69b // indirect
	google.golang.org/grpc v1.76.0 // indirect
	google.golang.org/protobuf v1.36.10 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/solarwinds/solarwinds-otel-collector-contrib/pkg/version => ../../pkg/version

replace github.com/solarwinds/solarwinds-otel-collector-contrib/pkg/testutil => ../../pkg/testutil

replace github.com/solarwinds/solarwinds-otel-collector-contrib/pkg/wmi => ./../../pkg/wmi

replace github.com/solarwinds/solarwinds-otel-collector-contrib/pkg/registry => ./../../pkg/registry
