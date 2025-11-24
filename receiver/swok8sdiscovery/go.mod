module github.com/solarwinds/solarwinds-otel-collector-contrib/receiver/swok8sdiscovery

go 1.25.3

require (
	github.com/solarwinds/solarwinds-otel-collector-contrib/internal/k8sconfig v0.136.7
	github.com/stretchr/testify v1.11.1
	go.opentelemetry.io/collector/component v1.42.0
	go.opentelemetry.io/collector/component/componenttest v0.136.0
	go.opentelemetry.io/collector/confmap v1.42.0
	go.opentelemetry.io/collector/confmap/xconfmap v0.136.0
	go.opentelemetry.io/collector/consumer v1.42.0
	go.opentelemetry.io/collector/consumer/consumertest v0.136.0
	go.opentelemetry.io/collector/pdata v1.42.0
	go.opentelemetry.io/collector/receiver v1.42.0
	go.opentelemetry.io/collector/receiver/receivertest v0.136.0
	go.uber.org/goleak v1.3.0
	go.uber.org/zap v1.27.1
	k8s.io/api v0.33.4
	k8s.io/apimachinery v0.33.4
	k8s.io/client-go v0.33.4
)

require (
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/emicklei/go-restful/v3 v3.12.2 // indirect
	github.com/fxamacker/cbor/v2 v2.8.0 // indirect
	github.com/go-logr/logr v1.4.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-openapi/jsonpointer v0.21.1 // indirect
	github.com/go-openapi/jsonreference v0.21.0 // indirect
	github.com/go-openapi/swag v0.23.1 // indirect
	github.com/go-viper/mapstructure/v2 v2.4.0 // indirect
	github.com/gobwas/glob v0.2.3 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/google/gnostic-models v0.6.9 // indirect
	github.com/google/go-cmp v0.7.0 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/hashicorp/go-version v1.7.0 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/knadh/koanf/maps v0.1.2 // indirect
	github.com/knadh/koanf/providers/confmap v1.0.0 // indirect
	github.com/knadh/koanf/v2 v2.3.0 // indirect
	github.com/mailru/easyjson v0.9.0 // indirect
	github.com/mitchellh/copystructure v1.2.0 // indirect
	github.com/mitchellh/reflectwalk v1.0.2 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.3-0.20250322232337-35a7c28c31ee // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/openshift/api v3.9.0+incompatible // indirect
	github.com/openshift/client-go v0.0.0-20250513150353-9ea84fa6431b // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/spf13/pflag v1.0.6 // indirect
	github.com/x448/float16 v0.8.4 // indirect
	go.opentelemetry.io/auto/sdk v1.2.1 // indirect
	go.opentelemetry.io/collector/consumer/consumererror v0.136.0 // indirect
	go.opentelemetry.io/collector/consumer/xconsumer v0.136.0 // indirect
	go.opentelemetry.io/collector/featuregate v1.42.0 // indirect
	go.opentelemetry.io/collector/internal/telemetry v0.136.0 // indirect
	go.opentelemetry.io/collector/pdata/pprofile v0.136.0 // indirect
	go.opentelemetry.io/collector/pipeline v1.42.0 // indirect
	go.opentelemetry.io/collector/receiver/xreceiver v0.136.0 // indirect
	go.opentelemetry.io/contrib/bridges/otelzap v0.13.0 // indirect
	go.opentelemetry.io/otel v1.38.0 // indirect
	go.opentelemetry.io/otel/log v0.14.0 // indirect
	go.opentelemetry.io/otel/metric v1.38.0 // indirect
	go.opentelemetry.io/otel/sdk v1.38.0 // indirect
	go.opentelemetry.io/otel/sdk/metric v1.38.0 // indirect
	go.opentelemetry.io/otel/trace v1.38.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.yaml.in/yaml/v3 v3.0.4 // indirect
	golang.org/x/net v0.42.0 // indirect
	golang.org/x/oauth2 v0.30.0 // indirect
	golang.org/x/sys v0.36.0 // indirect
	golang.org/x/term v0.33.0 // indirect
	golang.org/x/text v0.27.0 // indirect
	golang.org/x/time v0.11.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250707201910-8d1bb00bc6a7 // indirect
	google.golang.org/grpc v1.75.1 // indirect
	google.golang.org/protobuf v1.36.9 // indirect
	gopkg.in/evanphx/json-patch.v4 v4.12.0 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	k8s.io/klog/v2 v2.130.1 // indirect
	k8s.io/kube-openapi v0.0.0-20250318190949-c8a335a9a2ff // indirect
	k8s.io/utils v0.0.0-20250502105355-0f33e8f1c979 // indirect
	sigs.k8s.io/json v0.0.0-20241014173422-cfa47c3a1cc8 // indirect
	sigs.k8s.io/randfill v1.0.0 // indirect
	sigs.k8s.io/structured-merge-diff/v4 v4.7.0 // indirect
	sigs.k8s.io/yaml v1.4.0 // indirect
)

replace github.com/solarwinds/solarwinds-otel-collector-contrib/internal/k8sconfig => ../../internal/k8sconfig

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
