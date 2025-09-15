package swok8sdiscovery

import (
	// ...existing imports...
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	k8s "k8s.io/client-go/kubernetes"
	k8sfake "k8s.io/client-go/kubernetes/fake"
)

// mockClient encapsulates predefined k8s objects for tests.
type mockClient struct {
	pods     []*corev1.Pod
	services []*corev1.Service
}

func newMockClient() *mockClient {
	return &mockClient{
		pods:     []*corev1.Pod{},
		services: []*corev1.Service{},
	}
}

func (m *mockClient) getMockClient() (k8s.Interface, error) {
	objs := make([]runtime.Object, 0, len(m.pods)+len(m.services))
	for _, p := range m.pods {
		objs = append(objs, p)
	}
	for _, s := range m.services {
		objs = append(objs, s)
	}
	clientset := k8sfake.NewSimpleClientset(objs...)
	return clientset, nil
}
