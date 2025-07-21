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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/cache"
)

const (
	ServiceKind             = "Service"
	PodKind                 = "Pod"
	ReplicaSetKind          = "ReplicaSet"
	serviceTypeShort        = "svc"
	podTypeShort            = "pod"
	PodsWorkloadType        = "pods"
	ServicesWorkloadType    = "services"
	ReplicaSetsWorkloadType = "replicasets"
)

type LookupResult struct {
	Name      string
	Namespace string
	Kind      string
}

var EmptyLookupResult = LookupResult{}

// extractWorkloadKind extracts the kind of the workload from the given workload object.
// If preferPodOwner is true and the workload is a Pod, it looks up the owner kind of the Pod.
func extractWorkloadKind(workload any, logger *zap.Logger, informers map[string]cache.SharedIndexInformer, preferPodOwner bool) LookupResult {
	workloadObject, ok := workload.(runtime.Object)
	if !ok {
		logger.Error("Unexpected workload object type in cache", zap.String("workloadObjectType", fmt.Sprintf("%T", workload)))
		return EmptyLookupResult
	}
	kind := workloadObject.GetObjectKind().GroupVersionKind().Kind

	if preferPodOwner && kind == PodKind {
		ownerKind := lookupOwnerKindForWorkload(workload, logger, informers)
		if ownerKind != EmptyLookupResult {
			return ownerKind
		}
	}

	if kind != "" {
		workloadMeta, ok := workload.(metav1.Object)
		if !ok {
			logger.Error("Unexpected workload object type in cache", zap.String("workloadObjectType", fmt.Sprintf("%T", workload)))
			return EmptyLookupResult
		}
		return LookupResult{
			Name:      workloadMeta.GetName(),
			Namespace: workloadMeta.GetNamespace(),
			Kind:      kind,
		}
	} else {
		logger.Debug("Workload has no kind")
		return EmptyLookupResult
	}
}

// lookupOwnerKindForWorkload returns the kind of a workload's owner.
// If the owner is a ReplicaSet, it recursively looks up the owner of the ReplicaSet.
// If the workload has no owners, it returns an empty result.
func lookupOwnerKindForWorkload(workload any, logger *zap.Logger, informers map[string]cache.SharedIndexInformer) LookupResult {
	workloadMeta, ok := workload.(metav1.Object)
	if !ok {
		logger.Error("Unexpected workload object type in cache", zap.String("workloadObjectType", fmt.Sprintf("%T", workload)))
		return EmptyLookupResult
	}
	for _, owner := range workloadMeta.GetOwnerReferences() {
		if owner.Kind == ReplicaSetKind {
			ownerResult := lookupOwnerKindByNameAndNamespace(owner.Name, workloadMeta.GetNamespace(), ReplicaSetsWorkloadType, logger, informers)
			if ownerResult != EmptyLookupResult {
				return ownerResult
			}
		}
		if owner.Kind != "" {
			// Take the first non-empty kind from the owner references
			return LookupResult{
				Name:      owner.Name,
				Namespace: workloadMeta.GetNamespace(),
				Kind:      owner.Kind,
			}
		}
	}

	return EmptyLookupResult
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

func lookupWorkloadByNameAndNamespace(name string, namespace string, expectedTypes []string, logger *zap.Logger, informers map[string]cache.SharedIndexInformer) any {
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
			return workload
		}
	}
	return nil
}

// LookupWorkloadKindByNameAndNamespace looks up the workload kind by name and optional namespace.
// The list of ExpectedTypes in K8sWorkloadMappingConfig is used to determine which workload types to check.
// It returns the kind of the workload if found, or an empty result if not found.
func LookupWorkloadKindByNameAndNamespace(name string, namespace string, expectedTypes []string, logger *zap.Logger, informers map[string]cache.SharedIndexInformer, preferPodOwner bool) LookupResult {
	workload := lookupWorkloadByNameAndNamespace(name, namespace, expectedTypes, logger, informers)
	if workload == nil {
		return EmptyLookupResult
	} else {
		return extractWorkloadKind(workload, logger, informers, preferPodOwner)
	}
}

// lookupOwnerKindByNameAndNamespace looks up a workload by its name, namespace and expected type and returns the kind of its owner.
// If the workload is not found or has no owners, it returns an empty result.
func lookupOwnerKindByNameAndNamespace(name string, namespace string, workloadType string, logger *zap.Logger, informers map[string]cache.SharedIndexInformer) LookupResult {
	workload := lookupWorkloadByNameAndNamespace(name, namespace, []string{workloadType}, logger, informers)
	if workload == nil {
		return EmptyLookupResult
	} else {
		return lookupOwnerKindForWorkload(workload, logger, informers)
	}
}

// LookupWorkloadKindByIp looks up the workload kind by IP address.
// The list of ExpectedTypes in K8sWorkloadMappingConfig is used to determine which workload types to check. Only Pods and Services are supported.
// It returns the kind of the workload if found, or an empty result if not found.
func LookupWorkloadKindByIp(ip string, expectedTypes []string, logger *zap.Logger, informers map[string]cache.SharedIndexInformer, preferPodOwner bool) LookupResult {
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
			return extractWorkloadKind(workload, logger, informers, preferPodOwner)
		default:
			logger.Warn("Multiple workloads found for IP", zap.Int("count", len(workloads)))
		}
	}
	return EmptyLookupResult
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
		if parts[0] == serviceTypeShort || parts[0] == podTypeShort {
			// "host" is a three words, so it could be a name, a namespace and a type
			return parts[2], parts[1], parts[0]
		}
	case 5:
		if (parts[2] == serviceTypeShort || parts[2] == podTypeShort) && (parts[1] == "cluster" && parts[0] == "local") {
			// "host" is a five words, so it could be a name, a namespace, a type and a cluster domain
			return parts[4], parts[3], parts[2]
		}
	case 6:
		if (parts[3] == serviceTypeShort || parts[3] == podTypeShort) && (parts[2] == "cluster" && parts[1] == "local" && parts[0] == "") {
			// "host" is a six words, so it could be a name, a namespace, a type and a cluster domain ending with a dot
			return parts[5], parts[4], parts[3]
		}
	}

	// "host" is in an unknown format, so we don't know what it is
	return "", "", ""
}

// LookupWorkloadKindByHostname looks up the workload kind by hostname.
// It extracts the name and namespace from the hostname and uses them to look up the workload type.
// The namespaceFromAttr is used to validate the namespace extracted from the hostname.
// The expectedTypes are used to determine which workload types to check.
// It returns the kind of the workload if found, or an empty result if not found or if there is a mismatch in namespaces.
// If the hostname is in an unknown format, it returns an empty result.
// It has a special handling for well-known DNS formats for Pods and Services.
func LookupWorkloadKindByHostname(hostname string, namespaceFromAttr string, expectedTypes []string, logger *zap.Logger, informers map[string]cache.SharedIndexInformer, preferPodOwner bool) LookupResult {
	nameFromHostname, namespaceFromHostname, workloadTypeShort := extractNameAndNamespaceAndType(hostname)
	if nameFromHostname == "" {
		// It's unclear what the address is, so we can't determine the workload kind
		return EmptyLookupResult
	}

	switch workloadTypeShort {
	case podTypeShort:
		if namespaceFromAttr != "" && namespaceFromHostname != namespaceFromAttr {
			// The namespace in the address does not match the one in the attributes. This is suspicious.
			logger.Warn("Namespace mismatch", zap.String("namespaceInAddress", namespaceFromHostname), zap.String("namespaceInAttributes", namespaceFromAttr))
			return EmptyLookupResult
		}

		if preferPodOwner {
			ownerKind := lookupOwnerKindByNameAndNamespace(nameFromHostname, namespaceFromHostname, PodsWorkloadType, logger, informers)
			if ownerKind != EmptyLookupResult {
				return ownerKind
			}
		}

		return LookupResult{
			Name:      nameFromHostname,
			Namespace: namespaceFromHostname,
			Kind:      PodKind,
		}
	case serviceTypeShort:
		if namespaceFromAttr != "" && namespaceFromHostname != namespaceFromAttr {
			// The namespace in the address does not match the one in the attributes. This is suspicious.
			logger.Warn("Namespace mismatch", zap.String("namespaceInAddress", namespaceFromHostname), zap.String("namespaceInAttributes", namespaceFromAttr))
			return EmptyLookupResult
		}
		return LookupResult{
			Name:      nameFromHostname,
			Namespace: namespaceFromHostname,
			Kind:      ServiceKind,
		}
	default:
		if namespaceFromAttr != "" && namespaceFromHostname != "" && namespaceFromHostname != namespaceFromAttr {
			// The namespace in the address does not match the one in the attributes. It's unclear what the address is, so we can't determine the workload kind.
			return EmptyLookupResult
		}
		ns := namespaceFromHostname
		if ns == "" {
			ns = namespaceFromAttr
		}

		return LookupWorkloadKindByNameAndNamespace(nameFromHostname, ns, expectedTypes, logger, informers, preferPodOwner)
	}
}
