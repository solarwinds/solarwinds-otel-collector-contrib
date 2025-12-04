module github.com/solarwinds/solarwinds-otel-collector-contrib/processor/solarwindsprocessor

go 1.25.5

require (
	github.com/google/go-cmp v0.7.0
	github.com/solarwinds/solarwinds-otel-collector-contrib/extension/solarwindsextension v0.140.1
	github.com/solarwinds/solarwinds-otel-collector-contrib/pkg/attributesdecorator v0.140.1
	github.com/solarwinds/solarwinds-otel-collector-contrib/pkg/container v0.140.1
	github.com/solarwinds/solarwinds-otel-collector-contrib/pkg/extensionfinder v0.140.1
	github.com/stretchr/testify v1.11.1
	go.opentelemetry.io/collector/component v1.46.0
	go.opentelemetry.io/collector/component/componenttest v0.140.0
	go.opentelemetry.io/collector/confmap v1.46.0
	go.opentelemetry.io/collector/consumer v1.46.0
	go.opentelemetry.io/collector/consumer/consumertest v0.140.0
	go.opentelemetry.io/collector/pdata v1.46.0
	go.opentelemetry.io/collector/processor v1.46.0
	go.opentelemetry.io/collector/processor/processorhelper v0.140.0
	go.opentelemetry.io/collector/processor/processortest v0.140.0
	go.uber.org/goleak v1.3.0
	go.uber.org/zap v1.27.1
)

require (
	github.com/cenkalti/backoff/v5 v5.0.3 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/foxboron/go-tpm-keyfiles v0.0.0-20250903184740-5d135037bd4d // indirect
	github.com/fsnotify/fsnotify v1.9.0 // indirect
	github.com/go-logr/logr v1.4.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-viper/mapstructure/v2 v2.4.0 // indirect
	github.com/gobwas/glob v0.2.3 // indirect
	github.com/golang/snappy v1.0.0 // indirect
	github.com/google/go-tpm v0.9.7 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/hashicorp/go-version v1.7.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/klauspost/compress v1.18.0 // indirect
	github.com/knadh/koanf/maps v0.1.2 // indirect
	github.com/knadh/koanf/providers/confmap v1.0.0 // indirect
	github.com/knadh/koanf/v2 v2.3.0 // indirect
	github.com/mitchellh/copystructure v1.2.0 // indirect
	github.com/mitchellh/reflectwalk v1.0.2 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.3-0.20250322232337-35a7c28c31ee // indirect
	github.com/mostynb/go-grpc-compression v1.2.3 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/solarwinds/solarwinds-otel-collector-contrib/pkg/version v0.140.1 // indirect
	go.opentelemetry.io/auto/sdk v1.2.1 // indirect
	go.opentelemetry.io/collector v0.140.0 // indirect
	go.opentelemetry.io/collector/client v1.46.0 // indirect
	go.opentelemetry.io/collector/component/componentstatus v0.140.0 // indirect
	go.opentelemetry.io/collector/config/configauth v1.46.0 // indirect
	go.opentelemetry.io/collector/config/configcompression v1.46.0 // indirect
	go.opentelemetry.io/collector/config/configgrpc v0.140.0 // indirect
	go.opentelemetry.io/collector/config/configmiddleware v1.46.0 // indirect
	go.opentelemetry.io/collector/config/confignet v1.46.0 // indirect
	go.opentelemetry.io/collector/config/configopaque v1.46.0 // indirect
	go.opentelemetry.io/collector/config/configoptional v1.46.0 // indirect
	go.opentelemetry.io/collector/config/configretry v1.46.0 // indirect
	go.opentelemetry.io/collector/config/configtls v1.46.0 // indirect
	go.opentelemetry.io/collector/confmap/xconfmap v0.140.0 // indirect
	go.opentelemetry.io/collector/consumer/consumererror v0.140.0 // indirect
	go.opentelemetry.io/collector/consumer/consumererror/xconsumererror v0.140.0 // indirect
	go.opentelemetry.io/collector/consumer/xconsumer v0.140.0 // indirect
	go.opentelemetry.io/collector/exporter v1.46.0 // indirect
	go.opentelemetry.io/collector/exporter/exporterhelper v0.140.0 // indirect
	go.opentelemetry.io/collector/exporter/exporterhelper/xexporterhelper v0.140.0 // indirect
	go.opentelemetry.io/collector/exporter/otlpexporter v0.140.0 // indirect
	go.opentelemetry.io/collector/exporter/xexporter v0.140.0 // indirect
	go.opentelemetry.io/collector/extension v1.46.0 // indirect
	go.opentelemetry.io/collector/extension/extensionauth v1.46.0 // indirect
	go.opentelemetry.io/collector/extension/extensionmiddleware v0.140.0 // indirect
	go.opentelemetry.io/collector/extension/xextension v0.140.0 // indirect
	go.opentelemetry.io/collector/featuregate v1.46.0 // indirect
	go.opentelemetry.io/collector/pdata/pprofile v0.140.0 // indirect
	go.opentelemetry.io/collector/pdata/testdata v0.140.0 // indirect
	go.opentelemetry.io/collector/pdata/xpdata v0.140.0 // indirect
	go.opentelemetry.io/collector/pipeline v1.46.0 // indirect
	go.opentelemetry.io/collector/pipeline/xpipeline v0.140.0 // indirect
	go.opentelemetry.io/collector/processor/xprocessor v0.140.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.63.0 // indirect
	go.opentelemetry.io/otel v1.38.0 // indirect
	go.opentelemetry.io/otel/metric v1.38.0 // indirect
	go.opentelemetry.io/otel/sdk v1.38.0 // indirect
	go.opentelemetry.io/otel/sdk/metric v1.38.0 // indirect
	go.opentelemetry.io/otel/trace v1.38.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.yaml.in/yaml/v3 v3.0.4 // indirect
	golang.org/x/crypto v0.45.0 // indirect
	golang.org/x/net v0.47.0 // indirect
	golang.org/x/sys v0.38.0 // indirect
	golang.org/x/text v0.31.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250825161204-c5933d9347a5 // indirect
	google.golang.org/grpc v1.76.0 // indirect
	google.golang.org/protobuf v1.36.10 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace (
	github.com/solarwinds/solarwinds-otel-collector-contrib/extension/solarwindsextension => ../../extension/solarwindsextension
	github.com/solarwinds/solarwinds-otel-collector-contrib/pkg/attributesdecorator => ../../pkg/attributesdecorator
	github.com/solarwinds/solarwinds-otel-collector-contrib/pkg/container => ../../pkg/container
	github.com/solarwinds/solarwinds-otel-collector-contrib/pkg/extensionfinder => ../../pkg/extensionfinder
	github.com/solarwinds/solarwinds-otel-collector-contrib/pkg/testutil => ../../pkg/testutil
	github.com/solarwinds/solarwinds-otel-collector-contrib/pkg/version => ../../pkg/version
)
