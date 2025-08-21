package swok8sworkloadstatusprocessor

import (
	"github.com/solarwinds/solarwinds-otel-collector-contrib/internal/k8sconfig"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
