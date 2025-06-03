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
	"net"
	"net/url"
	"slices"
	"strings"

	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/cache"
)

const (
	ServiceKind          = "Service"
	PodKind              = "Pod"
	serviceKindShort     = "svc"
	podKindShort         = "pod"
	PodsWorkloadType     = "pods"
	ServicesWorkloadType = "services"
)

// extractWorkloadKind extracts the kind of the workload from the given workload object.
func extractWorkloadKind(workload any, logger *zap.Logger) string {
	workloadObject, ok := workload.(runtime.Object)
	if !ok {
		logger.Error("Unexpected workload object type in cache", zap.String("workloadObjectType", fmt.Sprintf("%T", workload)))
		return ""
	}
	kind := workloadObject.GetObjectKind().GroupVersionKind().Kind
	if kind != "" {
		return kind
	} else {
		logger.Debug("Workload has no kind")
		return ""
	}
}

// ExtractHostFromAddress extracts the hostname from a given address string.
func ExtractHostFromAddress(addr string) string {
	host := addr

	// If the address has a scheme, parse it as a URL
	if strings.Contains(addr, "://") {
		if u, err := url.Parse(addr); err == nil && u.Host != "" {
			host = u.Host
		}
	}

	// Try to remove port if present (handles both IPv4 and IPv6 with brackets)
	if h, _, err := net.SplitHostPort(host); err == nil {
		host = h
	}
	return host
}

// LookupWorkloadTypeByNameAndNamespace looks up the workload type by name and optional namespace.
// The list of ExpectedTypes in K8sWorkloadMappingConfig is used to determine which workload types to check.
// It returns the kind of the workload if found, or an empty string if not found.
func LookupWorkloadTypeByNameAndNamespace(name string, namespace string, expectedTypes []string, logger *zap.Logger, informers map[string]cache.SharedIndexInformer) string {
	workloadKey := name
	if namespace != "" {
		workloadKey = fmt.Sprintf("%s/%s", namespace, name)
	}

	for _, workloadType := range expectedTypes {
		logger := logger.WithLazy(zap.String("workloadType", workloadType), zap.String("workloadKey", workloadKey))
		workload, exists, err := informers[workloadType].GetStore().GetByKey(workloadKey)
		if err != nil {
			logger.Error("Error getting workload from cache", zap.Error(err))
			continue
		}
		if exists {
			return extractWorkloadKind(workload, logger)
		}
	}
	return ""
}

// LookupWorkloadTypeByIp looks up the workload type by IP address.
// The list of ExpectedTypes in K8sWorkloadMappingConfig is used to determine which workload types to check. Only Pods and Services are supported.
// It returns the kind of the workload if found, or an empty string if not found.
func LookupWorkloadTypeByIp(ip string, expectedTypes []string, logger *zap.Logger, informers map[string]cache.SharedIndexInformer) string {
	for _, workloadType := range expectedTypes {
		var indexerName string

		switch workloadType {
		case PodsWorkloadType:
			indexerName = PodIpIndex
		case ServicesWorkloadType:
			indexerName = ServiceIpIndex
		default:
			// Searching by IP is not supported for this workload type
			continue
		}

		logger := logger.WithLazy(zap.String("workloadType", workloadType), zap.String("ip", ip))

		workloads, err := informers[workloadType].GetIndexer().ByIndex(indexerName, ip)
		if err != nil {
			logger.Error("Error getting workload from cache", zap.Error(err))
			continue
		}
		switch len(workloads) {
		case 0:
			// No workload found for the given IP
		case 1:
			workload := workloads[0]
			return extractWorkloadKind(workload, logger)
		default:
			logger.Warn("Multiple workloads found for IP", zap.Int("count", len(workloads)))
		}
	}
	return ""
}

// extractNameAndNamespaceAndType extracts the possible name, namespace and type from the host string.
// It returns empty strings if the host is in an unknown format.
//
// The returned workloadTypeShort is either "pod", "svc" or empty. If not empty, the returned name and namespace are definite.
// Otherwise, the returned name and namespace are only possible values.
func extractNameAndNamespaceAndType(host string) (name string, namespace string, workloadTypeShort string) {
	parts := strings.Split(host, ".")
	slices.Reverse(parts)
	switch len(parts) {
	case 1:
		// "host" is a single word, so it could be a name
		return parts[0], "", ""
	case 2:
		// "host" is a two words, so it could be a name and a namespace
		return parts[1], parts[0], ""
	case 3:
		if parts[0] == serviceKindShort || parts[0] == podKindShort {
			// "host" is a three words, so it could be a name, a namespace and a type
			return parts[2], parts[1], parts[0]
		}
	case 5:
		if (parts[2] == serviceKindShort || parts[2] == podKindShort) && (parts[1] == "cluster" && parts[0] == "local") {
			// "host" is a five words, so it could be a name, a namespace, a type and a cluster domain
			return parts[4], parts[3], parts[2]
		}
	}

	// "host" is in an unknown format, so we don't know what it is
	return "", "", ""
}

// LookupWorkloadTypeByHostname looks up the workload type by hostname.
// It extracts the name and namespace from the hostname and uses them to look up the workload type.
// The namespaceFromAttr is used to validate the namespace extracted from the hostname.
// The expectedTypes are used to determine which workload types to check.
// It returns the kind of the workload if found, or an empty string if not found or if there is a mismatch in namespaces.
// If the hostname is in an unknown format, it returns an empty string.
// It has a special handling for well-known DNS formats for Pods and Services.
func LookupWorkloadTypeByHostname(hostname string, namespaceFromAttr string, expectedTypes []string, logger *zap.Logger, informers map[string]cache.SharedIndexInformer) string {
	nameFromHostname, namespaceFromHostname, workloadTypeShort := extractNameAndNamespaceAndType(hostname)
	if nameFromHostname == "" {
		// It's unclear what the address is, so we can't determine the workload type
		return ""
	}

	switch workloadTypeShort {
	case podKindShort:
		if namespaceFromAttr != "" && namespaceFromHostname != namespaceFromAttr {
			// The namespace in the address does not match the one in the attributes. This is suspicious.
			logger.Warn("Namespace mismatch", zap.String("namespaceInAddress", namespaceFromHostname), zap.String("namespaceInAttributes", namespaceFromAttr))
			return ""
		}
		return PodKind
	case serviceKindShort:
		if namespaceFromAttr != "" && namespaceFromHostname != namespaceFromAttr {
			// The namespace in the address does not match the one in the attributes. This is suspicious.
			logger.Warn("Namespace mismatch", zap.String("namespaceInAddress", namespaceFromHostname), zap.String("namespaceInAttributes", namespaceFromAttr))
			return ""
		}
		return ServiceKind
	default:
		if namespaceFromAttr != "" && namespaceFromHostname != "" && namespaceFromHostname != namespaceFromAttr {
			// The namespace in the address does not match the one in the attributes. It's unclear what the address is, so we can't determine the workload type.
			return ""
		}
		ns := namespaceFromHostname
		if ns == "" {
			ns = namespaceFromAttr
		}

		return LookupWorkloadTypeByNameAndNamespace(nameFromHostname, ns, expectedTypes, logger, informers)
	}
}
