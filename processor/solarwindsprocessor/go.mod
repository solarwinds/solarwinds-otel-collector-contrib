module github.com/solarwinds/solarwinds-otel-collector-contrib/processor/solarwindsprocessor

go 1.24.5

require (
	github.com/google/go-cmp v0.7.0
	github.com/solarwinds/solarwinds-otel-collector-contrib/extension/solarwindsextension v0.127.8
	github.com/solarwinds/solarwinds-otel-collector-contrib/pkg/attributesdecorator v0.127.8
	github.com/solarwinds/solarwinds-otel-collector-contrib/pkg/extensionfinder v0.127.8
	github.com/stretchr/testify v1.10.0
	go.opentelemetry.io/collector/component v1.33.0
	go.opentelemetry.io/collector/component/componenttest v0.127.0
	go.opentelemetry.io/collector/confmap v1.33.0
	go.opentelemetry.io/collector/consumer v1.33.0
	go.opentelemetry.io/collector/consumer/consumertest v0.127.0
	go.opentelemetry.io/collector/pdata v1.33.0
	go.opentelemetry.io/collector/processor v1.33.0
	go.opentelemetry.io/collector/processor/processorhelper v0.127.0
	go.opentelemetry.io/collector/processor/processortest v0.127.0
	go.uber.org/goleak v1.3.0
	go.uber.org/zap v1.27.0
)

require (
	github.com/cenkalti/backoff/v5 v5.0.2 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/foxboron/go-tpm-keyfiles v0.0.0-20250520203025-c3c3a4ec1653 // indirect
	github.com/fsnotify/fsnotify v1.9.0 // indirect
	github.com/go-logr/logr v1.4.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-viper/mapstructure/v2 v2.3.0 // indirect
	github.com/gobwas/glob v0.2.3 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/snappy v1.0.0 // indirect
	github.com/google/go-tpm v0.9.5 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/hashicorp/go-version v1.7.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/klauspost/compress v1.18.0 // indirect
	github.com/knadh/koanf/maps v0.1.2 // indirect
	github.com/knadh/koanf/providers/confmap v1.0.0 // indirect
	github.com/knadh/koanf/v2 v2.2.0 // indirect
	github.com/mitchellh/copystructure v1.2.0 // indirect
	github.com/mitchellh/reflectwalk v1.0.2 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/mostynb/go-grpc-compression v1.2.3 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/solarwinds/solarwinds-otel-collector-contrib/pkg/version v0.127.8 // indirect
	go.opentelemetry.io/auto/sdk v1.1.0 // indirect
	go.opentelemetry.io/collector v0.127.0 // indirect
	go.opentelemetry.io/collector/client v1.33.0 // indirect
	go.opentelemetry.io/collector/component/componentstatus v0.127.0 // indirect
	go.opentelemetry.io/collector/config/configauth v0.127.0 // indirect
	go.opentelemetry.io/collector/config/configcompression v1.33.0 // indirect
	go.opentelemetry.io/collector/config/configgrpc v0.127.0 // indirect
	go.opentelemetry.io/collector/config/configmiddleware v0.127.0 // indirect
	go.opentelemetry.io/collector/config/confignet v1.33.0 // indirect
	go.opentelemetry.io/collector/config/configopaque v1.33.0 // indirect
	go.opentelemetry.io/collector/config/configretry v1.33.0 // indirect
	go.opentelemetry.io/collector/config/configtls v1.33.0 // indirect
	go.opentelemetry.io/collector/consumer/consumererror v0.127.0 // indirect
	go.opentelemetry.io/collector/consumer/consumererror/xconsumererror v0.127.0 // indirect
	go.opentelemetry.io/collector/consumer/xconsumer v0.127.0 // indirect
	go.opentelemetry.io/collector/exporter v0.127.0 // indirect
	go.opentelemetry.io/collector/exporter/exporterhelper/xexporterhelper v0.127.0 // indirect
	go.opentelemetry.io/collector/exporter/otlpexporter v0.127.0 // indirect
	go.opentelemetry.io/collector/exporter/xexporter v0.127.0 // indirect
	go.opentelemetry.io/collector/extension v1.33.0 // indirect
	go.opentelemetry.io/collector/extension/extensionauth v1.33.0 // indirect
	go.opentelemetry.io/collector/extension/extensionmiddleware v0.127.0 // indirect
	go.opentelemetry.io/collector/extension/xextension v0.127.0 // indirect
	go.opentelemetry.io/collector/featuregate v1.33.0 // indirect
	go.opentelemetry.io/collector/internal/telemetry v0.127.0 // indirect
	go.opentelemetry.io/collector/pdata/pprofile v0.127.0 // indirect
	go.opentelemetry.io/collector/pdata/testdata v0.127.0 // indirect
	go.opentelemetry.io/collector/pipeline v0.127.0 // indirect
	go.opentelemetry.io/collector/pipeline/xpipeline v0.127.0 // indirect
	go.opentelemetry.io/collector/processor/xprocessor v0.127.0 // indirect
	go.opentelemetry.io/contrib/bridges/otelzap v0.11.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.61.0 // indirect
	go.opentelemetry.io/otel v1.36.0 // indirect
	go.opentelemetry.io/otel/log v0.12.2 // indirect
	go.opentelemetry.io/otel/metric v1.36.0 // indirect
	go.opentelemetry.io/otel/sdk v1.36.0 // indirect
	go.opentelemetry.io/otel/sdk/metric v1.36.0 // indirect
	go.opentelemetry.io/otel/trace v1.36.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/crypto v0.38.0 // indirect
	golang.org/x/net v0.40.0 // indirect
	golang.org/x/sys v0.33.0 // indirect
	golang.org/x/text v0.25.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250528174236-200df99c418a // indirect
	google.golang.org/grpc v1.72.2 // indirect
	google.golang.org/protobuf v1.36.6 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	sigs.k8s.io/yaml v1.4.0 // indirect
)

replace (
	github.com/solarwinds/solarwinds-otel-collector-contrib/extension/solarwindsextension => ../../extension/solarwindsextension
	github.com/solarwinds/solarwinds-otel-collector-contrib/pkg/attributesdecorator => ../../pkg/attributesdecorator
	github.com/solarwinds/solarwinds-otel-collector-contrib/pkg/extensionfinder => ../../pkg/extensionfinder
	github.com/solarwinds/solarwinds-otel-collector-contrib/pkg/testutil => ../../pkg/testutil
	github.com/solarwinds/solarwinds-otel-collector-contrib/pkg/version => ../../pkg/version
)
