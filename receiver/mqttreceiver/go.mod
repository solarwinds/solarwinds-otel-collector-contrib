module github.com/solarwinds/solarwinds-otel-collector-releases/receiver/mqttreceiver

go 1.24.2

require (
	github.com/eclipse/paho.mqtt.golang v1.5.0
	go.opentelemetry.io/collector/component v1.29.0
	go.opentelemetry.io/collector/consumer v1.29.0
	go.opentelemetry.io/collector/pdata v1.29.0
	go.opentelemetry.io/collector/receiver v1.29.0
	go.uber.org/zap v1.27.0
)

require (
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/go-logr/logr v1.4.2 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/gorilla/websocket v1.5.3 // indirect
	github.com/hashicorp/go-version v1.7.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	go.opentelemetry.io/auto/sdk v1.1.0 // indirect
	go.opentelemetry.io/collector/featuregate v1.29.0 // indirect
	go.opentelemetry.io/collector/internal/telemetry v0.123.0 // indirect
	go.opentelemetry.io/collector/pipeline v0.123.0 // indirect
	go.opentelemetry.io/contrib/bridges/otelzap v0.10.0 // indirect
	go.opentelemetry.io/otel v1.35.0 // indirect
	go.opentelemetry.io/otel/log v0.11.0 // indirect
	go.opentelemetry.io/otel/metric v1.35.0 // indirect
	go.opentelemetry.io/otel/sdk v1.35.0 // indirect
	go.opentelemetry.io/otel/sdk/metric v1.35.0 // indirect
	go.opentelemetry.io/otel/trace v1.35.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/net v0.38.0 // indirect
	golang.org/x/sync v0.12.0 // indirect
	golang.org/x/sys v0.32.0 // indirect
	golang.org/x/text v0.23.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250218202821-56aae31c358a // indirect
	google.golang.org/grpc v1.71.0 // indirect
	google.golang.org/protobuf v1.36.6 // indirect
)

replace github.com/solarwinds/solarwinds-otel-collector-releases/internal/k8sconfig => ../../internal/k8sconfig

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
