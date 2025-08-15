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
	"testing"

	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestBeylaAddressMapping(t *testing.T) {
	testPod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test_service",
			Namespace: "default",
			Labels: map[string]string{
				"app.kubernetes.io/name":    "my-app",
				"app.kubernetes.io/part-of": "my-namespace",
			},
		},
		Status: corev1.PodStatus{
			PodIP: "10.0.0.1",
		},
	}

	testDeployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-app",
			Namespace: "default", // Put in same namespace as pod
		},
	}

	tests := []struct {
		name                string
		existingPods        []*corev1.Pod
		existingDeployments []*appsv1.Deployment
		workloadMappings    []*K8sWorkloadMappingConfig
		receivedMetricAttrs map[string]any
		expectedMetricAttrs map[string]any
	}{
		{
			name:                "map_server_address_from_k8s_pod_name",
			existingPods:        []*corev1.Pod{testPod},
			existingDeployments: []*appsv1.Deployment{testDeployment},
			workloadMappings: []*K8sWorkloadMappingConfig{
				{
					AddressAttr:           "server.address",
					WorkloadTypeAttr:      "server.workload.type",
					WorkloadNameAttr:      "server.workload.name",
					WorkloadNamespaceAttr: "server.workload.namespace",
					ExpectedTypes:         []string{"pods"},
					PreferOwnerForPods:    true,
				},
			},
			receivedMetricAttrs: map[string]any{
				"client.address": "external-client",
				"server.address": "test_service",
			},
			expectedMetricAttrs: map[string]any{
				"client.address":            "external-client",
				"server.address":            "test_service",
				"server.workload.type":      "Pod",
				"server.workload.name":      "test_service",
				"server.workload.namespace": "default",
			},
		},
		{
			name:                "map_service_name_to_deployment_using_beyla_strategy",
			existingPods:        []*corev1.Pod{testPod},
			existingDeployments: []*appsv1.Deployment{testDeployment},
			workloadMappings: []*K8sWorkloadMappingConfig{
				{
					NameAttr:              "server.address",
					NamespaceAttr:         "server.namespace",
					WorkloadTypeAttr:      "server.workload.type",
					WorkloadNameAttr:      "server.workload.name",
					WorkloadNamespaceAttr: "server.workload.namespace",
					ExpectedTypes:         []string{"deployments"},
					PreferOwnerForPods:    true,
				},
			},
			receivedMetricAttrs: map[string]any{
				"client.address":   "external-client",
				"server.address":   "my-app",  // This should map to deployment by name
				"server.namespace": "default", // Updated to match deployment namespace
			},
			expectedMetricAttrs: map[string]any{
				"client.address":            "external-client",
				"server.address":            "my-app",
				"server.namespace":          "default", // Updated to match deployment namespace
				"server.workload.type":      "Deployment",
				"server.workload.name":      "my-app",
				"server.workload.namespace": "default", // Updated to match deployment namespace
			},
		},
		{
			name:                "map_beyla_service_name_by_address_to_deployment",
			existingPods:        []*corev1.Pod{testPod},
			existingDeployments: []*appsv1.Deployment{testDeployment},
			workloadMappings: []*K8sWorkloadMappingConfig{
				{
					AddressAttr:           "server.address",
					NamespaceAttr:         "server.namespace", // Add namespace attribute
					WorkloadTypeAttr:      "server.workload.type",
					WorkloadNameAttr:      "server.workload.name",
					WorkloadNamespaceAttr: "server.workload.namespace",
					ExpectedTypes:         []string{"deployments"},
					PreferOwnerForPods:    true,
				},
			},
			receivedMetricAttrs: map[string]any{
				"client.address":   "external-client",
				"server.address":   "my-app",  // This should map to deployment using Beyla address lookup
				"server.namespace": "default", // Add namespace to help with lookup
			},
			expectedMetricAttrs: map[string]any{
				"client.address":            "external-client",
				"server.address":            "my-app",
				"server.namespace":          "default", // Update expected result
				"server.workload.type":      "Deployment",
				"server.workload.name":      "my-app",
				"server.workload.namespace": "default",
			},
		},
		{
			name:                "map_beyla_service_name_with_deployment_priority",
			existingPods:        []*corev1.Pod{testPod},
			existingDeployments: []*appsv1.Deployment{testDeployment},
			workloadMappings: []*K8sWorkloadMappingConfig{
				{
					AddressAttr:           "server.address",
					NamespaceAttr:         "server.namespace", // Add namespace attribute
					WorkloadTypeAttr:      "server.workload.type",
					WorkloadNameAttr:      "server.workload.name",
					WorkloadNamespaceAttr: "server.workload.namespace",
					ExpectedTypes:         []string{"pods", "deployments"},
					PreferOwnerForPods:    true,
				},
			},
			receivedMetricAttrs: map[string]any{
				"client.address":   "external-client",
				"server.address":   "my-app",  // This finds deployment directly (priority 1) before checking pod labels
				"server.namespace": "default", // Add namespace to help with lookup
			},
			expectedMetricAttrs: map[string]any{
				"client.address":            "external-client",
				"server.address":            "my-app",
				"server.namespace":          "default",    // Update expected result
				"server.workload.type":      "Deployment", // Deployment found by direct name match (highest priority)
				"server.workload.name":      "my-app",
				"server.workload.namespace": "default",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, reset := MockKubeClient()
			defer reset()

			MockExistingObjectsInKubeClient(mock, tt.existingPods)
			MockExistingObjectsInKubeClient(mock, tt.existingDeployments)

			output := runProcessorMetricsPipelineTest(t, tt.workloadMappings, generateGaugeForTestProcessorMetricsPipeline(tt.receivedMetricAttrs))

			attrs := output[0].ResourceMetrics().At(0).ScopeMetrics().At(0).Metrics().At(0).Gauge().DataPoints().At(0).Attributes().AsRaw()
			require.Equal(t, tt.expectedMetricAttrs, attrs, "Expected attributes should match the actual attributes on metric exiting the processor")
		})
	}
}
