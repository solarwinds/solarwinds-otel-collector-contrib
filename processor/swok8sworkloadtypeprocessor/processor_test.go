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
	"maps"
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
		workloadMappings    []K8sWorkloadMappingConfig
		receivedMetricAttrs map[string]string
		expectedMetricAttrs map[string]any
	}{
		{
			name:         "mapping matches existing pod",
			existingPods: []*corev1.Pod{testPod},
			workloadMappings: []K8sWorkloadMappingConfig{
				{
					NameAttr:         "src_workload",
					NamespaceAttr:    "src_namespace",
					WorkloadTypeAttr: "src_type",
					ExpectedTypes:    []string{"pods"},
				},
			},
			receivedMetricAttrs: map[string]string{
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
			workloadMappings: []K8sWorkloadMappingConfig{
				{
					NameAttr:         "src_workload",
					NamespaceAttr:    "src_namespace",
					WorkloadTypeAttr: "src_type",
					ExpectedTypes:    []string{"pods"},
				},
			},
			receivedMetricAttrs: map[string]string{
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
			workloadMappings: []K8sWorkloadMappingConfig{
				{
					NameAttr:         "src_workload",
					NamespaceAttr:    "src_namespace",
					WorkloadTypeAttr: "src_type",
					ExpectedTypes:    []string{"pods"},
				},
			},
			receivedMetricAttrs: map[string]string{
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
			workloadMappings: []K8sWorkloadMappingConfig{
				{
					NameAttr:         "src_workload",
					NamespaceAttr:    "src_namespace",
					WorkloadTypeAttr: "src_type",
					ExpectedTypes:    []string{"pods"},
				},
			},
			receivedMetricAttrs: map[string]string{
				"src_namespace": testPod.Namespace,
			},
			expectedMetricAttrs: map[string]any{
				"src_namespace": testPod.Namespace,
			},
		},
		{
			name:         "mapping does not match existing pod because of missing namespace attribute",
			existingPods: []*corev1.Pod{testPod},
			workloadMappings: []K8sWorkloadMappingConfig{
				{
					NameAttr:         "src_workload",
					NamespaceAttr:    "src_namespace",
					WorkloadTypeAttr: "src_type",
					ExpectedTypes:    []string{"pods"},
				},
			},
			receivedMetricAttrs: map[string]string{
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
			workloadMappings: []K8sWorkloadMappingConfig{
				{
					NameAttr:         "src_workload",
					NamespaceAttr:    "src_namespace",
					WorkloadTypeAttr: "src_type",
					ExpectedTypes:    []string{"pods", "deployments"},
				},
			},
			receivedMetricAttrs: map[string]string{
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
			workloadMappings: []K8sWorkloadMappingConfig{
				{
					NameAttr:         "src_workload",
					NamespaceAttr:    "src_namespace",
					WorkloadTypeAttr: "src_type",
					ExpectedTypes:    []string{"pods", "deployments"},
				},
			},
			receivedMetricAttrs: map[string]string{
				"src_workload":  "shared_name",
				"src_namespace": "shared_namespace",
			},
			expectedMetricAttrs: map[string]any{
				"src_workload":  "shared_name",
				"src_namespace": "shared_namespace",
				"src_type":      "Pod",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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

			for _, pod := range tt.existingPods {
				mock.CoreV1().Pods(pod.Namespace).Create(context.Background(), pod, metav1.CreateOptions{})
			}
			for _, deployment := range tt.existingDeployments {
				mock.AppsV1().Deployments(deployment.Namespace).Create(context.Background(), deployment, metav1.CreateOptions{})
			}

			factory := NewFactory()

			cfg := factory.CreateDefaultConfig().(*Config)
			cfg.WorkloadMappings = tt.workloadMappings
			err := cfg.Validate()
			require.NoError(t, err)

			sink := new(consumertest.MetricsSink)

			c, err := factory.CreateMetrics(context.Background(), processortest.NewNopSettings(factory.Type()), cfg, sink)
			require.NoError(t, err)

			err = c.Start(context.Background(), componenttest.NewNopHost())
			require.NoError(t, err)

			require.NotPanics(t, func() {
				metric := generateGaugeForTestProcessorMetricsPipeline(tt.receivedMetricAttrs)
				err = c.ConsumeMetrics(context.Background(), metric)
			})

			require.NoError(t, err)
			err = c.Shutdown(context.Background())
			require.NoError(t, err)

			sentMetrics := sink.AllMetrics()

			attrs := sentMetrics[0].ResourceMetrics().At(0).ScopeMetrics().At(0).Metrics().At(0).Gauge().DataPoints().At(0).Attributes().AsRaw()
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
		existingDeployments    []*appsv1.Deployment
		receivedMetricProvider func(map[string]string) pmetric.Metrics
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
			mock.MockedServerPreferredResources = []*metav1.APIResourceList{
				{
					GroupVersion: "v1",
					APIResources: []metav1.APIResource{
						{
							Name: "pods",
							Kind: "Pod",
						},
					},
				},
			}

			mock.CoreV1().Pods(testPod.Namespace).Create(context.Background(), testPod, metav1.CreateOptions{})

			factory := NewFactory()

			cfg := factory.CreateDefaultConfig().(*Config)
			cfg.WorkloadMappings = []K8sWorkloadMappingConfig{
				{
					NameAttr:         "src_workload",
					NamespaceAttr:    "src_namespace",
					WorkloadTypeAttr: "src_type",
					ExpectedTypes:    []string{"pods"},
				},
			}
			err := cfg.Validate()
			require.NoError(t, err)

			sink := new(consumertest.MetricsSink)

			c, err := factory.CreateMetrics(context.Background(), processortest.NewNopSettings(factory.Type()), cfg, sink)
			require.NoError(t, err)

			err = c.Start(context.Background(), componenttest.NewNopHost())
			require.NoError(t, err)

			require.NotPanics(t, func() {
				metric := tt.receivedMetricProvider(
					map[string]string{
						"src_workload":  testPod.Name,
						"src_namespace": testPod.Namespace,
					})
				err = c.ConsumeMetrics(context.Background(), metric)
			})

			require.NoError(t, err)
			err = c.Shutdown(context.Background())
			require.NoError(t, err)

			sentMetrics := sink.AllMetrics()

			require.Equal(t,
				map[string]any{
					"src_workload":  testPod.Name,
					"src_namespace": testPod.Namespace,
					"src_type":      "Pod",
				},
				tt.actualAttrsProvider(sentMetrics[0]),
				"Expected attributes should match the actual attributes on metric exiting the processor")
		})
	}
}

func generateGaugeForTestProcessorMetricsPipeline(attrs map[string]string) pmetric.Metrics {
	metrics, m := generateEmptyMetricForTestProcessorMetricsPipeline()
	dp := m.SetEmptyGauge().DataPoints().AppendEmpty()
	for k, v := range maps.All(attrs) {
		dp.Attributes().PutStr(k, v)
	}
	dp.SetIntValue(123)
	dp.SetTimestamp(pcommon.NewTimestampFromTime(time.Now()))
	return metrics
}

func generateSumForTestProcessorMetricsPipeline(attrs map[string]string) pmetric.Metrics {
	metrics, m := generateEmptyMetricForTestProcessorMetricsPipeline()
	dp := m.SetEmptySum().DataPoints().AppendEmpty()
	for k, v := range maps.All(attrs) {
		dp.Attributes().PutStr(k, v)
	}
	dp.SetIntValue(123)
	dp.SetTimestamp(pcommon.NewTimestampFromTime(time.Now()))
	return metrics
}

func generateHistogramForTestProcessorMetricsPipeline(attrs map[string]string) pmetric.Metrics {
	metrics, m := generateEmptyMetricForTestProcessorMetricsPipeline()
	dp := m.SetEmptyHistogram().DataPoints().AppendEmpty()
	for k, v := range maps.All(attrs) {
		dp.Attributes().PutStr(k, v)
	}
	dp.SetCount(123)
	dp.SetTimestamp(pcommon.NewTimestampFromTime(time.Now()))
	return metrics
}

func generateExponentialHistogramForTestProcessorMetricsPipeline(attrs map[string]string) pmetric.Metrics {
	metrics, m := generateEmptyMetricForTestProcessorMetricsPipeline()
	dp := m.SetEmptyExponentialHistogram().DataPoints().AppendEmpty()
	for k, v := range maps.All(attrs) {
		dp.Attributes().PutStr(k, v)
	}
	dp.SetCount(123)
	dp.SetTimestamp(pcommon.NewTimestampFromTime(time.Now()))
	return metrics
}

func generateSummaryForTestProcessorMetricsPipeline(attrs map[string]string) pmetric.Metrics {
	metrics, m := generateEmptyMetricForTestProcessorMetricsPipeline()
	dp := m.SetEmptySummary().DataPoints().AppendEmpty()
	for k, v := range maps.All(attrs) {
		dp.Attributes().PutStr(k, v)
	}
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
