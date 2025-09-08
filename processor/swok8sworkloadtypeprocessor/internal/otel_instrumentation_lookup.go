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

package internal // import "github.com/solarwinds/solarwinds-otel-collector-contrib/processor/swok8sworkloadtypeprocessor/internal"

import (
	"fmt"
	"slices"

	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/cache"
)

const (
	OtelServiceNameAnnotation      = "resource.opentelemetry.io/service.name"
	OtelServiceNamespaceAnnotation = "resource.opentelemetry.io/service.namespace"
	KubernetesAppNameLabel         = "app.kubernetes.io/name"
	KubernetesAppPartofLabel       = "app.kubernetes.io/part-of"
)

// LookupWorkloadByOtelInstrumentationServiceName performs workload discovery using OpenTelemetry
// instrumentation conventions for service name resolution.
//
// Priority order (following OTEL instrumentation patterns):
// 1. Service name and namespace set via Pod env variables
// 2. Service name and namespace set set via Pod annotations
// 3. Service name and namespace set set via Pod labels
// 4. Service name set to a Pod's parent workload's name
// 5. Service name set to a Pod's name
// 6. Service name set to a Container's name
// 7. Service name set to a process's name
// 8. Service name set to a domain name or IP address
//
// Options 1., 6. and 7. are not currently implemented.
// Options 4., 5. and 8. are already handled by [LookupWorkloadKindByHostname] and [LookupWorkloadKindByIp].
// Options 2. and 3. are handled by this function.
func LookupWorkloadByOtelInstrumentationServiceName(serviceName string, namespaceFromAttr string, expectedTypes []string, logger *zap.Logger, informers map[string]cache.SharedIndexInformer, preferPodOwner bool) LookupResult {
	nameFromHostname, namespaceFromHostname, _ := ExtractNameAndNamespaceAndType(serviceName)
	if nameFromHostname == "" {
		// It's unclear what the address is, so we can't determine the workload kind
		return EmptyLookupResult
	}

	podInformer, exists := informers[PodsWorkloadType]
	if !exists || !slices.Contains(expectedTypes, PodsWorkloadType) {
		return EmptyLookupResult
	}

	lookupByIndex := func(indexName string, getNamespaceFromLabelsOrAnnotations func(pod metav1.Object) string) LookupResult {
		pods, err := podInformer.GetIndexer().ByIndex(indexName, nameFromHostname)
		if err != nil {
			logger.Error("Error getting pods from cache", zap.Error(err))
		}

		foundPods := slices.Collect(func(yield func(metav1.Object) bool) {
			for _, podObj := range pods {
				pod, ok := podObj.(metav1.Object)
				if !ok {
					logger.Error("Unexpected workload object type in cache", zap.String("workloadObjectType", fmt.Sprintf("%T", podObj)))
					continue
				}

				if namespaceFromHostname == "" || (getNamespaceFromLabelsOrAnnotations(pod) == namespaceFromHostname || namespaceFromAttr == namespaceFromHostname) {
					if !yield(pod) {
						return
					}
				}
			}
		})

		switch len(foundPods) {
		case 0:
			// No Pod found for the given Service name
		case 1:
			return extractWorkloadKind(foundPods[0], logger, informers, preferPodOwner)
		default:
			// If multiple Pods are found, we need to check they have the same owner
			if !preferPodOwner {
				// Multiple pods found for the same Service name
			} else {
				owners := make(map[LookupResult]bool)
				for _, pod := range foundPods {
					owner := lookupOwnerForWorkload(pod, logger, informers)
					if owner != EmptyLookupResult {
						owners[owner] = true
					}
				}

				switch len(owners) {
				case 0:
					// Multiple pods with no parents found for the same Service name
				case 1:
					// If all pods have the same owner, return it
					for owner := range owners {
						return owner
					}
				default:
					// Multiple workloads found for the same Service name
				}
			}
		}

		return EmptyLookupResult
	}

	// Service name and namespace set set via Pod labels
	resultByLabel := lookupByIndex(ServiceNameFromPodLabelIndex, func(pod metav1.Object) string {
		return pod.GetLabels()[KubernetesAppPartofLabel]
	})

	if resultByLabel != EmptyLookupResult {
		return resultByLabel
	}

	// Service name and namespace set set via Pod annotations
	resultByAnnotation := lookupByIndex(ServiceNameFromPodAnnotationIndex, func(pod metav1.Object) string {
		return pod.GetAnnotations()[OtelServiceNamespaceAnnotation]
	})

	if resultByAnnotation != EmptyLookupResult {
		return resultByAnnotation
	}

	logger.Debug("No workload found for OTEL instrumentation service name", zap.String("serviceName", serviceName))
	return EmptyLookupResult
}
