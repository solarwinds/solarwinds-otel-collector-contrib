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

package internal

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/cache"
)

// TestLookupWorkloadKindByNameAndNamespace tests the basic lookup functionality
func TestLookupWorkloadKindByNameAndNamespace(t *testing.T) {
	logger := zap.NewNop()

	// Create a test service
	testService := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-service",
			Namespace: "test-namespace",
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
	}

	serviceInformer := cache.NewSharedIndexInformer(
		nil, // ListerWatcher is nil for testing
		testService,
		0, // resyncPeriod
		cache.Indexers{},
	)
	err := serviceInformer.GetStore().Add(testService)
	assert.NoError(t, err)

	informers := map[string]cache.SharedIndexInformer{
		ServicesWorkloadType: serviceInformer,
	}

	// Test successful lookup with both name and namespace
	result := LookupWorkloadKindByNameAndNamespace(
		"test-service",
		"test-namespace",
		[]string{ServicesWorkloadType},
		logger,
		informers,
		false,
	)

	assert.Equal(t, "test-service", result.Name)
	assert.Equal(t, "test-namespace", result.Namespace)
	assert.Equal(t, ServiceKind, result.Kind)
}

// TestLookupWorkloadKindByNameAndNamespaceNotFound tests that lookup returns nil when not found
func TestLookupWorkloadKindByNameAndNamespaceNotFound(t *testing.T) {
	logger := zap.NewNop()

	serviceInformer := cache.NewSharedIndexInformer(
		nil, // ListerWatcher is nil for testing
		&corev1.Service{},
		0, // resyncPeriod
		cache.Indexers{},
	)

	informers := map[string]cache.SharedIndexInformer{
		ServicesWorkloadType: serviceInformer,
	}

	// Test lookup with non-existent service
	result := LookupWorkloadKindByNameAndNamespace(
		"non-existent-service",
		"test-namespace",
		[]string{ServicesWorkloadType},
		logger,
		informers,
		false,
	)

	assert.Equal(t, "", result.Name, "Should not find non-existent service")
	assert.Equal(t, "", result.Namespace, "Should not find non-existent service")
	assert.Equal(t, "", result.Kind, "Should not find non-existent service")
}

// TestLookupWorkloadByNameAndNamespace tests the internal lookup function
func TestLookupWorkloadByNameAndNamespace(t *testing.T) {
	logger := zap.NewNop()

	// Create a test service
	testService := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-service",
			Namespace: "test-namespace",
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
	}

	serviceInformer := cache.NewSharedIndexInformer(
		nil, // ListerWatcher is nil for testing
		testService,
		0, // resyncPeriod
		cache.Indexers{},
	)
	err := serviceInformer.GetStore().Add(testService)
	assert.NoError(t, err)

	informers := map[string]cache.SharedIndexInformer{
		ServicesWorkloadType: serviceInformer,
	}

	// Test successful lookup
	result := lookupWorkloadByNameAndNamespace(
		"test-service",
		"test-namespace",
		[]string{ServicesWorkloadType},
		logger,
		informers,
	)

	assert.NotNil(t, result, "Should find the service")

	// Cast result to proper type to check fields
	service, ok := result.(*corev1.Service)
	assert.True(t, ok, "Result should be a Service")
	assert.Equal(t, "test-service", service.Name)
	assert.Equal(t, "test-namespace", service.Namespace)

	// Test lookup with empty namespace (should fail without cross-namespace fallback)
	result2 := lookupWorkloadByNameAndNamespace(
		"test-service",
		"",
		[]string{ServicesWorkloadType},
		logger,
		informers,
	)

	assert.Nil(t, result2, "Should not find service with empty namespace")
}
