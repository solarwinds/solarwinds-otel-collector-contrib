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
	"slices"

	"github.com/solarwinds/solarwinds-otel-collector-contrib/internal/k8sconfig"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/discovery"
	fakeDiscovery "k8s.io/client-go/discovery/fake"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
)

func MockKubeClient() (mock *FakeKubeClient, reset func()) {
	origClient := initKubeClient

	newClient := &FakeKubeClient{
		Clientset: *fake.NewSimpleClientset(),
	}

	initKubeClient = func(k8sconfig.APIConfig) (kubernetes.Interface, error) {
		return newClient, nil
	}
	return newClient, func() { initKubeClient = origClient }
}

type FakeKubeClient struct {
	fake.Clientset
	fakeDiscovery.FakeDiscovery
	MockedServerPreferredResources []*metav1.APIResourceList
}

func (c *FakeKubeClient) Discovery() discovery.DiscoveryInterface {
	return c
}

func (c *FakeKubeClient) ServerPreferredResources() ([]*metav1.APIResourceList, error) {
	return c.MockedServerPreferredResources, nil
}

// MockServerPreferredResources extends the mocked ServerPreferredResources with an additional kind, if it's not already there.
func (c *FakeKubeClient) MockServerPreferredResources(groupVersion string, name string, kind string) {
	var resourceList *metav1.APIResourceList
	for _, resList := range c.MockedServerPreferredResources {
		if resList.GroupVersion == groupVersion {
			resourceList = resList
			break
		}
	}

	if resourceList == nil {
		resourceList = &metav1.APIResourceList{GroupVersion: groupVersion}
		c.MockedServerPreferredResources = append(c.MockedServerPreferredResources, resourceList)
	}

	if !slices.ContainsFunc(resourceList.APIResources, func(r metav1.APIResource) bool {
		return r.Name == name
	}) {
		resourceList.APIResources = append(resourceList.APIResources, metav1.APIResource{
			Name: name,
			Kind: kind,
		})
	}
}

// MockExistingObjectsInKubeClient creates the provided items in the fake Kubernetes client.
// It will also mock the ServerPreferredResources for the respective kind if it is not already mocked.
func MockExistingObjectsInKubeClient[T corev1.Pod | appsv1.Deployment | appsv1.StatefulSet | appsv1.DaemonSet | corev1.Service | appsv1.ReplicaSet](c *FakeKubeClient, items []*T) {
	for _, item := range items {
		switch v := any(item).(type) {
		case *corev1.Pod:
			c.MockServerPreferredResources("v1", "pods", "Pod")
			c.CoreV1().Pods(v.Namespace).Create(context.Background(), v, metav1.CreateOptions{})
		case *appsv1.Deployment:
			c.MockServerPreferredResources("apps/v1", "deployments", "Deployment")
			c.AppsV1().Deployments(v.Namespace).Create(context.Background(), v, metav1.CreateOptions{})
		case *appsv1.StatefulSet:
			c.MockServerPreferredResources("apps/v1", "statefulsets", "StatefulSet")
			c.AppsV1().StatefulSets(v.Namespace).Create(context.Background(), v, metav1.CreateOptions{})
		case *appsv1.DaemonSet:
			c.MockServerPreferredResources("apps/v1", "daemonsets", "DaemonSet")
			c.AppsV1().DaemonSets(v.Namespace).Create(context.Background(), v, metav1.CreateOptions{})
		case *corev1.Service:
			c.MockServerPreferredResources("v1", "services", "Service")
			c.CoreV1().Services(v.Namespace).Create(context.Background(), v, metav1.CreateOptions{})
		case *appsv1.ReplicaSet:
			c.MockServerPreferredResources("apps/v1", "replicasets", "ReplicaSet")
			c.AppsV1().ReplicaSets(v.Namespace).Create(context.Background(), v, metav1.CreateOptions{})
		}
	}
}
