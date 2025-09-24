module github.com/solarwinds/solarwinds-otel-collector-contrib/receiver/mqttreceiver

go 1.25.1

require (
	github.com/eclipse/paho.mqtt.golang v1.5.1
	github.com/google/go-cmp v0.7.0
	github.com/stretchr/testify v1.11.1
	go.opentelemetry.io/collector/component v1.41.0
	go.opentelemetry.io/collector/component/componenttest v0.135.0
	go.opentelemetry.io/collector/confmap v1.41.0
	go.opentelemetry.io/collector/consumer v1.41.0
	go.opentelemetry.io/collector/consumer/consumertest v0.135.0
	go.opentelemetry.io/collector/filter v0.135.0
	go.opentelemetry.io/collector/pdata v1.41.0
	go.opentelemetry.io/collector/receiver v1.41.0
	go.opentelemetry.io/collector/receiver/receivertest v0.135.0
	go.uber.org/goleak v1.3.0
	go.uber.org/zap v1.27.0
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/go-logr/logr v1.4.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-viper/mapstructure/v2 v2.4.0 // indirect
	github.com/gobwas/glob v0.2.3 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/gorilla/websocket v1.5.3 // indirect
	github.com/hashicorp/go-version v1.7.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/knadh/koanf/maps v0.1.2 // indirect
	github.com/knadh/koanf/providers/confmap v1.0.0 // indirect
	github.com/knadh/koanf/v2 v2.3.0 // indirect
	github.com/mitchellh/copystructure v1.2.0 // indirect
	github.com/mitchellh/reflectwalk v1.0.2 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.3-0.20250322232337-35a7c28c31ee // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	go.opentelemetry.io/auto/sdk v1.2.1 // indirect
	go.opentelemetry.io/collector/consumer/consumererror v0.135.0 // indirect
	go.opentelemetry.io/collector/consumer/xconsumer v0.135.0 // indirect
	go.opentelemetry.io/collector/featuregate v1.41.0 // indirect
	go.opentelemetry.io/collector/internal/telemetry v0.135.0 // indirect
	go.opentelemetry.io/collector/pdata/pprofile v0.135.0 // indirect
	go.opentelemetry.io/collector/pipeline v1.41.0 // indirect
	go.opentelemetry.io/collector/receiver/xreceiver v0.135.0 // indirect
	go.opentelemetry.io/contrib/bridges/otelzap v0.13.0 // indirect
	go.opentelemetry.io/otel v1.38.0 // indirect
	go.opentelemetry.io/otel/log v0.14.0 // indirect
	go.opentelemetry.io/otel/metric v1.38.0 // indirect
	go.opentelemetry.io/otel/sdk v1.38.0 // indirect
	go.opentelemetry.io/otel/sdk/metric v1.38.0 // indirect
	go.opentelemetry.io/otel/trace v1.38.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.yaml.in/yaml/v3 v3.0.4 // indirect
	golang.org/x/exp v0.0.0-20250911091902-df9299821621 // indirect
	golang.org/x/net v0.44.0 // indirect
	golang.org/x/sync v0.17.0 // indirect
	golang.org/x/sys v0.36.0 // indirect
	golang.org/x/text v0.29.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250908214217-97024824d090 // indirect
	google.golang.org/grpc v1.75.1 // indirect
	google.golang.org/protobuf v1.36.9 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

// Dependabot fails on this indirect dependency.
// openshift removed all tags from their repo, use the pseudoversion from the release-4.4, first with go.mod
replace github.com/openshift/api v3.9.0+incompatible => github.com/openshift/api v0.0.0-20200618202633-7192180f496a

// Dependabot fails on this indirect dependency.
// https://gonum.org/v1/gonum?go-get=1 returns 404, and dependabot gives up and fails.
// Using the latest version from go.sum or go.mod when run without this replace.
replace gonum.org/v1/gonum => github.com/gonum/gonum v0.15.1

// Also breaks dependabot, dependency of the one github.com/openshift/api.
// Using the latest version from go.sum or go.mod when run without this replace.
replace gonum.org/v1/netlib => github.com/gonum/netlib v0.0.0-20190331212654-76723241ea4e

retract (
	v0.76.2
	v0.76.1
	v0.65.0
)
