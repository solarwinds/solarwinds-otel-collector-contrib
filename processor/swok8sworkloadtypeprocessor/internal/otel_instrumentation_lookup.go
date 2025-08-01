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
	"net"
	"slices"

	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/cache"
)

// LookupWorkloadByOtelInstrumentationServiceName performs workload discovery using OpenTelemetry
// instrumentation conventions for service name resolution. This follows the standard priority order
// used by OTEL-based instrumentation tools for service discovery.
//
// Priority order (following OTEL instrumentation patterns):
// 1. Direct workload name lookup (deployments, statefulsets, etc.)
// 2. Pod annotation-based lookup (resource.opentelemetry.io/service.name)
// 3. Pod label-based lookup (app.kubernetes.io/name)
// 4. Pod name lookup
func LookupWorkloadByOtelInstrumentationServiceName(serviceName string, namespace string, expectedTypes []string, logger *zap.Logger, informers map[string]cache.SharedIndexInformer, preferPodOwner bool) LookupResult {
	// Try direct workload name lookups first (priority 1)
	for _, workloadType := range expectedTypes {
		result := LookupWorkloadKindByNameAndNamespace(serviceName, namespace, []string{workloadType}, logger, informers, preferPodOwner)
		if result != EmptyLookupResult {
			logger.Debug("Found workload by direct name lookup",
				zap.String("serviceName", serviceName),
				zap.String("workloadType", workloadType))
			return result
		}
	}

	// If direct lookup fails, try pod-based lookups (priorities 2-4)
	if slices.Contains(expectedTypes, PodsWorkloadType) {
		// First check if we have a pod informer
		podInformer, exists := informers[PodsWorkloadType]
		if !exists {
			return EmptyLookupResult
		}

		// Search through all pods to find matches based on Beyla's priority order
		pods := podInformer.GetStore().List()
		for _, podObj := range pods {
			pod, ok := podObj.(metav1.Object)
			if !ok {
				continue
			}

			// Skip pods not in the target namespace (if specified)
			if namespace != "" && pod.GetNamespace() != namespace {
				continue
			}

			// Priority 2: Check resource.opentelemetry.io/service.name annotation
			if annotations := pod.GetAnnotations(); annotations != nil {
				if annotationValue := annotations["resource.opentelemetry.io/service.name"]; annotationValue == serviceName {
					result := extractWorkloadKind(podObj, logger, informers, preferPodOwner)
					if result != EmptyLookupResult {
						logger.Debug("Found workload by OTEL service name annotation",
							zap.String("serviceName", serviceName),
							zap.String("podName", pod.GetName()))
						return result
					}
				}
			}

			// Priority 3: Check app.kubernetes.io/name label
			if labels := pod.GetLabels(); labels != nil {
				if labelValue := labels["app.kubernetes.io/name"]; labelValue == serviceName {
					result := extractWorkloadKind(podObj, logger, informers, preferPodOwner)
					if result != EmptyLookupResult {
						logger.Debug("Found workload by app.kubernetes.io/name label",
							zap.String("serviceName", serviceName),
							zap.String("podName", pod.GetName()))
						return result
					}
				}
			}

			// Priority 4: Check pod name
			if pod.GetName() == serviceName {
				result := extractWorkloadKind(podObj, logger, informers, preferPodOwner)
				if result != EmptyLookupResult {
					logger.Debug("Found workload by pod name",
						zap.String("serviceName", serviceName),
						zap.String("podName", pod.GetName()))
					return result
				}
			}
		}
	}

	logger.Debug("No workload found for OTEL instrumentation service name", zap.String("serviceName", serviceName))
	return EmptyLookupResult
}

// LookupWorkloadByOtelInstrumentationAddress extends the existing address lookup to handle OTEL
// instrumentation-style service identification. It first tries the existing hostname/IP lookup,
// and if that fails, it falls back to OTEL instrumentation service name mapping.
func LookupWorkloadByOtelInstrumentationAddress(address string, namespace string, expectedTypes []string, logger *zap.Logger, informers map[string]cache.SharedIndexInformer, preferPodOwner bool) LookupResult {
	// Extract host from address (removes protocol, port, etc.)
	host := ExtractHostFromAddress(address)

	// First try existing IP-based lookup
	if isValidIP(host) {
		result := LookupWorkloadKindByIp(host, expectedTypes, logger, informers, preferPodOwner)
		if result != EmptyLookupResult {
			return result
		}
	}

	// Then try existing hostname-based lookup
	result := LookupWorkloadKindByHostname(host, namespace, expectedTypes, logger, informers, preferPodOwner)
	if result != EmptyLookupResult {
		return result
	}

	// Finally, try OTEL instrumentation service name mapping
	// The host could be a service name extracted by OTEL instrumentation
	result = LookupWorkloadByOtelInstrumentationServiceName(host, namespace, expectedTypes, logger, informers, preferPodOwner)
	if result != EmptyLookupResult {
		return result
	}

	logger.Debug("No workload found for OTEL instrumentation address", zap.String("address", address))
	return EmptyLookupResult
	return EmptyLookupResult
}

// isValidIP checks if the given string is a valid IP address
func isValidIP(host string) bool {
	// Use net.ParseIP for accurate IP validation
	return net.ParseIP(host) != nil
}
