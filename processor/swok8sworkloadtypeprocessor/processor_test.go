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
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/consumer/consumertest"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/processor/processortest"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestProcessorMetricsPipeline(t *testing.T) {

	testPod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test_pod",
			Namespace: "test_pod_namespace",
		},
	}
	testDeployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test_deployment",
			Namespace: "test_deployment_namespace",
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
			name:         "mapping matches existing pod",
			existingPods: []*corev1.Pod{testPod},
			workloadMappings: []*K8sWorkloadMappingConfig{
				{
					NameAttr:         "src_workload",
					NamespaceAttr:    "src_namespace",
					WorkloadTypeAttr: "src_type",
					ExpectedTypes:    []string{"pods"},
				},
			},
			receivedMetricAttrs: map[string]any{
				"src_workload":  testPod.Name,
				"src_namespace": testPod.Namespace,
			},
			expectedMetricAttrs: map[string]any{
				"src_workload":  testPod.Name,
				"src_namespace": testPod.Namespace,
				"src_type":      "Pod",
			},
		},
		{
			name:         "mapping does not match existing pod because of different name",
			existingPods: []*corev1.Pod{testPod},
			workloadMappings: []*K8sWorkloadMappingConfig{
				{
					NameAttr:         "src_workload",
					NamespaceAttr:    "src_namespace",
					WorkloadTypeAttr: "src_type",
					ExpectedTypes:    []string{"pods"},
				},
			},
			receivedMetricAttrs: map[string]any{
				"src_workload":  "other_pod",
				"src_namespace": testPod.Namespace,
			},
			expectedMetricAttrs: map[string]any{
				"src_workload":  "other_pod",
				"src_namespace": testPod.Namespace,
			},
		},
		{
			name:         "mapping does not match existing pod because of different namespace",
			existingPods: []*corev1.Pod{testPod},
			workloadMappings: []*K8sWorkloadMappingConfig{
				{
					NameAttr:         "src_workload",
					NamespaceAttr:    "src_namespace",
					WorkloadTypeAttr: "src_type",
					ExpectedTypes:    []string{"pods"},
				},
			},
			receivedMetricAttrs: map[string]any{
				"src_workload":  testPod.Name,
				"src_namespace": "other_pod_namespace",
			},
			expectedMetricAttrs: map[string]any{
				"src_workload":  testPod.Name,
				"src_namespace": "other_pod_namespace",
			},
		},
		{
			name:         "mapping does not match existing pod because of missing name attribute",
			existingPods: []*corev1.Pod{testPod},
			workloadMappings: []*K8sWorkloadMappingConfig{
				{
					NameAttr:         "src_workload",
					NamespaceAttr:    "src_namespace",
					WorkloadTypeAttr: "src_type",
					ExpectedTypes:    []string{"pods"},
				},
			},
			receivedMetricAttrs: map[string]any{
				"src_namespace": testPod.Namespace,
			},
			expectedMetricAttrs: map[string]any{
				"src_namespace": testPod.Namespace,
			},
		},
		{
			name:         "mapping does not match existing pod because of missing namespace attribute",
			existingPods: []*corev1.Pod{testPod},
			workloadMappings: []*K8sWorkloadMappingConfig{
				{
					NameAttr:         "src_workload",
					NamespaceAttr:    "src_namespace",
					WorkloadTypeAttr: "src_type",
					ExpectedTypes:    []string{"pods"},
				},
			},
			receivedMetricAttrs: map[string]any{
				"src_workload": testPod.Name,
			},
			expectedMetricAttrs: map[string]any{
				"src_workload": testPod.Name,
			},
		},
		{
			name:                "mapping matches existing deployment",
			existingPods:        []*corev1.Pod{testPod},
			existingDeployments: []*appsv1.Deployment{testDeployment},
			workloadMappings: []*K8sWorkloadMappingConfig{
				{
					NameAttr:         "src_workload",
					NamespaceAttr:    "src_namespace",
					WorkloadTypeAttr: "src_type",
					ExpectedTypes:    []string{"pods", "deployments"},
				},
			},
			receivedMetricAttrs: map[string]any{
				"src_workload":  testDeployment.Name,
				"src_namespace": testDeployment.Namespace,
			},
			expectedMetricAttrs: map[string]any{
				"src_workload":  testDeployment.Name,
				"src_namespace": testDeployment.Namespace,
				"src_type":      "Deployment",
			},
		},
		{
			name: "mapping matches existing pod even though there is a deployment with the same name and namespace",
			existingPods: []*corev1.Pod{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "shared_name",
						Namespace: "shared_namespace",
					},
				},
			},
			existingDeployments: []*appsv1.Deployment{testDeployment},
			workloadMappings: []*K8sWorkloadMappingConfig{
				{
					NameAttr:         "src_workload",
					NamespaceAttr:    "src_namespace",
					WorkloadTypeAttr: "src_type",
					ExpectedTypes:    []string{"pods", "deployments"},
				},
			},
			receivedMetricAttrs: map[string]any{
				"src_workload":  "shared_name",
				"src_namespace": "shared_namespace",
			},
			expectedMetricAttrs: map[string]any{
				"src_workload":  "shared_name",
				"src_namespace": "shared_namespace",
				"src_type":      "Pod",
			},
		},
		{
			name:         "mapping does not overwrite existing workload type",
			existingPods: []*corev1.Pod{testPod},
			workloadMappings: []*K8sWorkloadMappingConfig{
				{
					NameAttr:         "src_workload",
					NamespaceAttr:    "src_namespace",
					WorkloadTypeAttr: "src_type",
					ExpectedTypes:    []string{"pods"},
				},
			},
			receivedMetricAttrs: map[string]any{
				"src_workload":  testPod.Name,
				"src_namespace": testPod.Namespace,
				"src_type":      "already_set",
			},
			expectedMetricAttrs: map[string]any{
				"src_workload":  testPod.Name,
				"src_namespace": testPod.Namespace,
				"src_type":      "already_set",
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

func TestProcessorMetricsPipelineForDifferentMetricTypes(t *testing.T) {
	testPod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test_pod",
			Namespace: "test_pod_namespace",
		},
	}

	tests := []struct {
		name                   string
		receivedMetricProvider func(map[string]any) pmetric.Metrics
		actualAttrsProvider    func(pmetric.Metrics) map[string]any
	}{
		{
			name:                   "gauge",
			receivedMetricProvider: generateGaugeForTestProcessorMetricsPipeline,
			actualAttrsProvider: func(m pmetric.Metrics) map[string]any {
				return m.ResourceMetrics().At(0).ScopeMetrics().At(0).Metrics().At(0).Gauge().DataPoints().At(0).Attributes().AsRaw()
			},
		},
		{
			name:                   "sum",
			receivedMetricProvider: generateSumForTestProcessorMetricsPipeline,
			actualAttrsProvider: func(m pmetric.Metrics) map[string]any {
				return m.ResourceMetrics().At(0).ScopeMetrics().At(0).Metrics().At(0).Sum().DataPoints().At(0).Attributes().AsRaw()
			},
		},
		{
			name:                   "histogram",
			receivedMetricProvider: generateHistogramForTestProcessorMetricsPipeline,
			actualAttrsProvider: func(m pmetric.Metrics) map[string]any {
				return m.ResourceMetrics().At(0).ScopeMetrics().At(0).Metrics().At(0).Histogram().DataPoints().At(0).Attributes().AsRaw()
			},
		},
		{
			name:                   "exponential histogram",
			receivedMetricProvider: generateExponentialHistogramForTestProcessorMetricsPipeline,
			actualAttrsProvider: func(m pmetric.Metrics) map[string]any {
				return m.ResourceMetrics().At(0).ScopeMetrics().At(0).Metrics().At(0).ExponentialHistogram().DataPoints().At(0).Attributes().AsRaw()
			},
		},
		{
			name:                   "summary",
			receivedMetricProvider: generateSummaryForTestProcessorMetricsPipeline,
			actualAttrsProvider: func(m pmetric.Metrics) map[string]any {
				return m.ResourceMetrics().At(0).ScopeMetrics().At(0).Metrics().At(0).Summary().DataPoints().At(0).Attributes().AsRaw()
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, reset := MockKubeClient()
			defer reset()

			MockExistingObjectsInKubeClient(mock, []*corev1.Pod{testPod})

			output := runProcessorMetricsPipelineTest(t, []*K8sWorkloadMappingConfig{
				{
					NameAttr:         "src_workload",
					NamespaceAttr:    "src_namespace",
					WorkloadTypeAttr: "src_type",
					ExpectedTypes:    []string{"pods"},
				},
			}, tt.receivedMetricProvider(
				map[string]any{
					"src_workload":  testPod.Name,
					"src_namespace": testPod.Namespace,
				}))

			require.Equal(t,
				map[string]any{
					"src_workload":  testPod.Name,
					"src_namespace": testPod.Namespace,
					"src_type":      "Pod",
				},
				tt.actualAttrsProvider(output[0]),
				"Expected attributes should match the actual attributes on metric exiting the processor")
		})
	}
}

func TestProcessorMetricsPipelineForDifferentContexts(t *testing.T) {

	testPod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test_pod",
			Namespace: "test_pod_namespace",
		},
	}
	testDeployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test_deployment",
			Namespace: "test_deployment_namespace",
		},
	}
	testStatefulSet := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test_statefulset",
			Namespace: "test_statefulset_namespace",
		},
	}
	testDaemonSet := &appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test_daemonset",
			Namespace: "test_daemonset_namespace",
		},
	}

	tests := []struct {
		name                   string
		existingPods           []*corev1.Pod
		existingDeployments    []*appsv1.Deployment
		existingStatefulSets   []*appsv1.StatefulSet
		existingDaemonSets     []*appsv1.DaemonSet
		workloadMappings       []*K8sWorkloadMappingConfig
		receivedResourceAttrs  map[string]any
		expectedResourceAttrs  map[string]any
		receivedScopeAttrs     map[string]any
		expectedScopeAttrs     map[string]any
		receivedMetricAttrs    map[string]any
		expectedMetricAttrs    map[string]any
		receivedDatapointAttrs map[string]any
		expectedDatapointAttrs map[string]any
	}{
		{
			name:                 "mapping matches all existing workload types defined in all scopes",
			existingPods:         []*corev1.Pod{testPod},
			existingDeployments:  []*appsv1.Deployment{testDeployment},
			existingStatefulSets: []*appsv1.StatefulSet{testStatefulSet},
			existingDaemonSets:   []*appsv1.DaemonSet{testDaemonSet},
			workloadMappings: []*K8sWorkloadMappingConfig{
				{
					NameAttr:         "resource.src_workload",
					NamespaceAttr:    "resource.src_namespace",
					WorkloadTypeAttr: "resource.src_type",
					ExpectedTypes:    []string{"pods"},
				},
				{
					NameAttr:         "metric.src_workload",
					NamespaceAttr:    "metric.src_namespace",
					WorkloadTypeAttr: "metric.src_type",
					ExpectedTypes:    []string{"daemonsets"},
				},
				{
					NameAttr:         "scope.src_workload",
					NamespaceAttr:    "scope.src_namespace",
					WorkloadTypeAttr: "scope.src_type",
					ExpectedTypes:    []string{"statefulsets"},
				},
				{
					NameAttr:         "datapoint.src_workload",
					NamespaceAttr:    "datapoint.src_namespace",
					WorkloadTypeAttr: "datapoint.src_type",
					ExpectedTypes:    []string{"deployments"},
				},
			},
			receivedResourceAttrs: map[string]any{
				"src_workload":  testPod.Name,
				"src_namespace": testPod.Namespace,
			},
			expectedResourceAttrs: map[string]any{
				"src_workload":  testPod.Name,
				"src_namespace": testPod.Namespace,
				"src_type":      "Pod",
			},
			receivedScopeAttrs: map[string]any{
				"src_workload":  testStatefulSet.Name,
				"src_namespace": testStatefulSet.Namespace,
			},
			expectedScopeAttrs: map[string]any{
				"src_workload":  testStatefulSet.Name,
				"src_namespace": testStatefulSet.Namespace,
				"src_type":      "StatefulSet",
			},
			receivedMetricAttrs: map[string]any{
				"src_workload":  testDaemonSet.Name,
				"src_namespace": testDaemonSet.Namespace,
			},
			expectedMetricAttrs: map[string]any{
				"src_workload":  testDaemonSet.Name,
				"src_namespace": testDaemonSet.Namespace,
				"src_type":      "DaemonSet",
			},
			receivedDatapointAttrs: map[string]any{
				"src_workload":  testDeployment.Name,
				"src_namespace": testDeployment.Namespace,
			},
			expectedDatapointAttrs: map[string]any{
				"src_workload":  testDeployment.Name,
				"src_namespace": testDeployment.Namespace,
				"src_type":      "Deployment",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, reset := MockKubeClient()
			defer reset()

			MockExistingObjectsInKubeClient(mock, tt.existingPods)
			MockExistingObjectsInKubeClient(mock, tt.existingDeployments)
			MockExistingObjectsInKubeClient(mock, tt.existingStatefulSets)
			MockExistingObjectsInKubeClient(mock, tt.existingDaemonSets)

			output := runProcessorMetricsPipelineTest(t, tt.workloadMappings, generateGaugeWithAllAttributesForTestProcessorMetricsPipeline(tt.receivedResourceAttrs, tt.receivedScopeAttrs, tt.receivedMetricAttrs, tt.receivedDatapointAttrs))

			resourceAttrs := output[0].ResourceMetrics().At(0).Resource().Attributes().AsRaw()
			require.Equal(t, tt.expectedResourceAttrs, resourceAttrs, "Expected attributes should match the actual resource attributes on metric exiting the processor")

			scopeAttrs := output[0].ResourceMetrics().At(0).ScopeMetrics().At(0).Scope().Attributes().AsRaw()
			require.Equal(t, tt.expectedScopeAttrs, scopeAttrs, "Expected attributes should match the actual scope attributes on metric exiting the processor")

			metricAttrs := output[0].ResourceMetrics().At(0).ScopeMetrics().At(0).Metrics().At(0).Metadata().AsRaw()
			require.Equal(t, tt.expectedMetricAttrs, metricAttrs, "Expected attributes should match the actual metric attributes on metric exiting the processor")

			datapointAttrs := output[0].ResourceMetrics().At(0).ScopeMetrics().At(0).Metrics().At(0).Gauge().DataPoints().At(0).Attributes().AsRaw()
			require.Equal(t, tt.expectedDatapointAttrs, datapointAttrs, "Expected attributes should match the actual datapoint attributes on metric exiting the processor")
		})
	}
}

func generateGaugeForTestProcessorMetricsPipeline(attrs map[string]any) pmetric.Metrics {
	metrics, m := generateEmptyMetricForTestProcessorMetricsPipeline()
	dp := m.SetEmptyGauge().DataPoints().AppendEmpty()
	dp.Attributes().FromRaw(attrs)
	dp.SetIntValue(123)
	dp.SetTimestamp(pcommon.NewTimestampFromTime(time.Now()))
	return metrics
}

func generateSumForTestProcessorMetricsPipeline(attrs map[string]any) pmetric.Metrics {
	metrics, m := generateEmptyMetricForTestProcessorMetricsPipeline()
	dp := m.SetEmptySum().DataPoints().AppendEmpty()
	dp.Attributes().FromRaw(attrs)
	dp.SetIntValue(123)
	dp.SetTimestamp(pcommon.NewTimestampFromTime(time.Now()))
	return metrics
}

func generateHistogramForTestProcessorMetricsPipeline(attrs map[string]any) pmetric.Metrics {
	metrics, m := generateEmptyMetricForTestProcessorMetricsPipeline()
	dp := m.SetEmptyHistogram().DataPoints().AppendEmpty()
	dp.Attributes().FromRaw(attrs)
	dp.SetCount(123)
	dp.SetTimestamp(pcommon.NewTimestampFromTime(time.Now()))
	return metrics
}

func generateExponentialHistogramForTestProcessorMetricsPipeline(attrs map[string]any) pmetric.Metrics {
	metrics, m := generateEmptyMetricForTestProcessorMetricsPipeline()
	dp := m.SetEmptyExponentialHistogram().DataPoints().AppendEmpty()
	dp.Attributes().FromRaw(attrs)
	dp.SetCount(123)
	dp.SetTimestamp(pcommon.NewTimestampFromTime(time.Now()))
	return metrics
}

func generateSummaryForTestProcessorMetricsPipeline(attrs map[string]any) pmetric.Metrics {
	metrics, m := generateEmptyMetricForTestProcessorMetricsPipeline()
	dp := m.SetEmptySummary().DataPoints().AppendEmpty()
	dp.Attributes().FromRaw(attrs)
	dp.SetCount(123)
	dp.SetTimestamp(pcommon.NewTimestampFromTime(time.Now()))
	return metrics
}

func generateEmptyMetricForTestProcessorMetricsPipeline() (pmetric.Metrics, pmetric.Metric) {
	metrics := pmetric.NewMetrics()
	rm := metrics.ResourceMetrics().AppendEmpty()
	m := rm.ScopeMetrics().AppendEmpty().Metrics().AppendEmpty()
	m.SetName("test_metric")
	return metrics, m
}

func generateGaugeWithAllAttributesForTestProcessorMetricsPipeline(resourceAttrs map[string]any, scopeAttrs map[string]any, metricAttrs map[string]any, datapointAttrs map[string]any) pmetric.Metrics {
	metrics := generateGaugeForTestProcessorMetricsPipeline(datapointAttrs)
	metrics.ResourceMetrics().At(0).Resource().Attributes().FromRaw(resourceAttrs)
	metrics.ResourceMetrics().At(0).ScopeMetrics().At(0).Scope().Attributes().FromRaw(scopeAttrs)
	metrics.ResourceMetrics().At(0).ScopeMetrics().At(0).Metrics().At(0).Metadata().FromRaw(metricAttrs)
	return metrics
}

func runProcessorMetricsPipelineTest(t *testing.T, workloadMappings []*K8sWorkloadMappingConfig, input pmetric.Metrics) []pmetric.Metrics {
	factory := NewFactory()

	cfg := factory.CreateDefaultConfig().(*Config)
	cfg.WorkloadMappings = workloadMappings
	err := cfg.Validate()
	require.NoError(t, err)

	sink := new(consumertest.MetricsSink)

	c, err := factory.CreateMetrics(context.Background(), processortest.NewNopSettings(factory.Type()), cfg, sink)
	require.NoError(t, err)

	err = c.Start(context.Background(), componenttest.NewNopHost())
	require.NoError(t, err)

	require.NotPanics(t, func() {
		metric := input
		err = c.ConsumeMetrics(context.Background(), metric)
	})

	require.NoError(t, err)
	err = c.Shutdown(context.Background())
	require.NoError(t, err)

	return sink.AllMetrics()
}
