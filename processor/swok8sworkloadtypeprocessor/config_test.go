// Copyright 2025 SolarWinds Worldwide, LLC. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package swok8sworkloadtypeprocessor

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/confmap/confmaptest"
	"go.opentelemetry.io/collector/confmap/xconfmap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/solarwinds/solarwinds-otel-collector-contrib/internal/k8sconfig"
	"github.com/solarwinds/solarwinds-otel-collector-contrib/processor/swok8sworkloadtypeprocessor/internal/metadata"
)

func TestInvalidConfig(t *testing.T) {
	tests := []struct {
		id       component.ID
		expected *Config
	}{
		{
			id: component.NewIDWithName(metadata.Type, "invalid_api_config"),
		},
		{
			id: component.NewIDWithName(metadata.Type, "missing_expected_types"),
		},
		{
			id: component.NewIDWithName(metadata.Type, "missing_workload_type"),
		},
		{
			id: component.NewIDWithName(metadata.Type, "invalid_expected_types"),
		},
		{
			id: component.NewIDWithName(metadata.Type, "expected_type_has_empty_kind"),
		},
		{
			id: component.NewIDWithName(metadata.Type, "missing_workload_mappings"),
		},
		{
			id: component.NewIDWithName(metadata.Type, "inconsistent_attribute_context"),
		},
		{
			id: component.NewIDWithName(metadata.Type, "valid_config"),
			expected: &Config{
				APIConfig: k8sconfig.APIConfig{
					AuthType: k8sconfig.AuthTypeServiceAccount,
				},
				WorkloadMappings: []*K8sWorkloadMappingConfig{
					{
						NameAttr:         "source_workload",
						NamespaceAttr:    "source_workload_namespace",
						WorkloadTypeAttr: "source_workload_type",
						ExpectedTypes:    []string{"deployments"},
					},
					{
						NameAttr:         "dest_workload",
						NamespaceAttr:    "dest_workload_namespace",
						WorkloadTypeAttr: "dest_workload_type",
						ExpectedTypes:    []string{"services", "pods"},
					},
				},
				WatchSyncPeriod: time.Minute * 5,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.id.String(), func(t *testing.T) {
			mock, reset := MockKubeClient()
			defer reset()
			mock.MockedServerPreferredResources = []*metav1.APIResourceList{
				{
					GroupVersion: "v1",
					APIResources: []metav1.APIResource{
						{
							Name: "pods",
							Kind: "Pod",
						},
						{
							Name: "services",
							Kind: "Service",
						},
						{
							Name: "withinvalidkinds",
							Kind: "",
						},
					},
				},
				{
					GroupVersion: "apps/v1",
					APIResources: []metav1.APIResource{
						{
							Name: "deployments",
							Kind: "Deployment",
						},
					},
				},
			}

			cm, err := confmaptest.LoadConf(filepath.Join("testdata", "config.yaml"))
			require.NoError(t, err)

			factory := NewFactory()
			cfg := factory.CreateDefaultConfig().(*Config)

			sub, err := cm.Sub(tt.id.String())
			require.NoError(t, err)
			require.NoError(t, sub.Unmarshal(cfg))

			if tt.expected == nil {
				err = xconfmap.Validate(cfg)
				require.Error(t, err)
			} else {
				require.NoError(t, xconfmap.Validate(cfg))
				require.EqualExportedValues(t, tt.expected, cfg, "User-configurable fields should be parsed correctly")
			}
		})
	}
}

func TestMappingContextInConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		id       string
		expected K8sWorkloadMappingConfig
		data     K8sWorkloadMappingConfig
	}{
		{
			id: "default_datapoint_context",
			expected: K8sWorkloadMappingConfig{
				NameAttr:         "source_workload",
				NamespaceAttr:    "source_workload_namespace",
				WorkloadTypeAttr: "source_workload_type",
				context:          DataPointContext,
			},
			data: K8sWorkloadMappingConfig{
				NameAttr:         "source_workload",
				NamespaceAttr:    "source_workload_namespace",
				WorkloadTypeAttr: "source_workload_type",
			},
		},
		{
			id: "default_datapoint_context_no_namespace",
			expected: K8sWorkloadMappingConfig{
				NameAttr:         "source_workload",
				WorkloadTypeAttr: "source_workload_type",
				context:          DataPointContext,
			},
			data: K8sWorkloadMappingConfig{
				NameAttr:         "source_workload",
				WorkloadTypeAttr: "source_workload_type",
			},
		},
		{
			id: "default_datapoint_context_unrelated_prefix",
			expected: K8sWorkloadMappingConfig{
				NameAttr:         "something.source_workload",
				NamespaceAttr:    "something.source_workload_namespace",
				WorkloadTypeAttr: "something.source_workload_type",
				context:          DataPointContext,
			},
			data: K8sWorkloadMappingConfig{
				NameAttr:         "something.source_workload",
				NamespaceAttr:    "something.source_workload_namespace",
				WorkloadTypeAttr: "something.source_workload_type",
			},
		},
		{
			id: "datapoint_context",
			expected: K8sWorkloadMappingConfig{
				NameAttr:         "source_workload",
				NamespaceAttr:    "source_workload_namespace",
				WorkloadTypeAttr: "source_workload_type",
				context:          DataPointContext,
			},
			data: K8sWorkloadMappingConfig{
				NameAttr:         "datapoint.source_workload",
				NamespaceAttr:    "datapoint.source_workload_namespace",
				WorkloadTypeAttr: "datapoint.source_workload_type",
			},
		},
		{
			id: "datapoint_context_no_namespace",
			expected: K8sWorkloadMappingConfig{
				NameAttr:         "source_workload",
				WorkloadTypeAttr: "source_workload_type",
				context:          DataPointContext,
			},
			data: K8sWorkloadMappingConfig{
				NameAttr:         "datapoint.source_workload",
				WorkloadTypeAttr: "datapoint.source_workload_type",
			},
		},
		{
			id: "metric_context",
			expected: K8sWorkloadMappingConfig{
				NameAttr:         "source_workload",
				NamespaceAttr:    "source_workload_namespace",
				WorkloadTypeAttr: "source_workload_type",
				context:          MetricContext,
			},
			data: K8sWorkloadMappingConfig{
				NameAttr:         "metric.source_workload",
				NamespaceAttr:    "metric.source_workload_namespace",
				WorkloadTypeAttr: "metric.source_workload_type",
			},
		},
		{
			id: "metric_context_no_namespace",
			expected: K8sWorkloadMappingConfig{
				NameAttr:         "source_workload",
				WorkloadTypeAttr: "source_workload_type",
				context:          MetricContext,
			},
			data: K8sWorkloadMappingConfig{
				NameAttr:         "metric.source_workload",
				WorkloadTypeAttr: "metric.source_workload_type",
			},
		},
		{
			id: "scope_context",
			expected: K8sWorkloadMappingConfig{
				NameAttr:         "source_workload",
				NamespaceAttr:    "source_workload_namespace",
				WorkloadTypeAttr: "source_workload_type",
				context:          ScopeContext,
			},
			data: K8sWorkloadMappingConfig{
				NameAttr:         "scope.source_workload",
				NamespaceAttr:    "scope.source_workload_namespace",
				WorkloadTypeAttr: "scope.source_workload_type",
			},
		},
		{
			id: "scope_context_no_namespace",
			expected: K8sWorkloadMappingConfig{
				NameAttr:         "source_workload",
				WorkloadTypeAttr: "source_workload_type",
				context:          ScopeContext,
			},
			data: K8sWorkloadMappingConfig{
				NameAttr:         "scope.source_workload",
				WorkloadTypeAttr: "scope.source_workload_type",
			},
		},
		{
			id: "resource_context",
			expected: K8sWorkloadMappingConfig{
				NameAttr:         "source_workload",
				NamespaceAttr:    "source_workload_namespace",
				WorkloadTypeAttr: "source_workload_type",
				context:          ResourceContext,
			},
			data: K8sWorkloadMappingConfig{
				NameAttr:         "resource.source_workload",
				NamespaceAttr:    "resource.source_workload_namespace",
				WorkloadTypeAttr: "resource.source_workload_type",
			},
		},
		{
			id: "resource_context_no_namespace",
			expected: K8sWorkloadMappingConfig{
				NameAttr:         "source_workload",
				WorkloadTypeAttr: "source_workload_type",
				context:          ResourceContext,
			},
			data: K8sWorkloadMappingConfig{
				NameAttr:         "resource.source_workload",
				WorkloadTypeAttr: "resource.source_workload_type",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.id, func(t *testing.T) {

			// mock expected types for the test
			tt.expected.ExpectedTypes = []string{"deployments"}
			tt.data.ExpectedTypes = []string{"deployments"}
			mappedExpectedTypes := map[string]groupVersionResourceKind{
				"deployments": {
					kind: "Deployment",
				}}

			err := tt.data.validate(nil, mappedExpectedTypes)

			require.NoError(t, err)
			require.Equal(t, tt.expected, tt.data, "Initialized config should be equal to expected config")
		})
	}
}

func TestMappingMismatchedContextInConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		id   string
		data K8sWorkloadMappingConfig
	}{
		{
			id: "default_and_datapoint",
			data: K8sWorkloadMappingConfig{
				NameAttr:         "source_workload",
				NamespaceAttr:    "source_workload_namespace",
				WorkloadTypeAttr: "datapoint.source_workload_type",
			},
		},
		{
			id: "datapoint_and_scope",
			data: K8sWorkloadMappingConfig{
				NameAttr:         "scope.source_workload",
				WorkloadTypeAttr: "datapoint.source_workload_type",
			},
		},
		{
			id: "datapoint_and_resource",
			data: K8sWorkloadMappingConfig{
				NameAttr:         "resource.source_workload",
				NamespaceAttr:    "datapoint.source_workload_namespace",
				WorkloadTypeAttr: "resource.source_workload_type",
			},
		},
		{
			id: "metric_and_resource",
			data: K8sWorkloadMappingConfig{
				NameAttr:         "resource.source_workload",
				NamespaceAttr:    "metric.source_workload_namespace",
				WorkloadTypeAttr: "resource.source_workload_type",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.id, func(t *testing.T) {

			// mock expected types for the test
			tt.data.ExpectedTypes = []string{"deployments"}
			mappedExpectedTypes := map[string]groupVersionResourceKind{
				"deployments": {
					kind: "Deployment",
				}}

			err := tt.data.validate(nil, mappedExpectedTypes)

			require.EqualError(t, err, "inconsistent context in workload mapping")
		})
	}
}
