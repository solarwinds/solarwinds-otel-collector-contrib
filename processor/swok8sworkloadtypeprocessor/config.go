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

package swok8sworkloadtypeprocessor // import "github.com/solarwinds/solarwinds-otel-collector-contrib/processor/swok8sworkloadtypeprocessor"

import (
	"fmt"
	"maps"
	"slices"
	"strings"
	"time"

	"github.com/solarwinds/solarwinds-otel-collector-contrib/internal/k8sconfig"
	"github.com/solarwinds/solarwinds-otel-collector-contrib/processor/swok8sworkloadtypeprocessor/internal"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/kubernetes"
)

type Config struct {
	k8sconfig.APIConfig `mapstructure:",squash"`
	WorkloadMappings    []*K8sWorkloadMappingConfig `mapstructure:"workload_mappings"`
	WatchSyncPeriod     time.Duration               `mapstructure:"watch_sync_period"`
	mappedExpectedTypes map[string]groupVersionResourceKind
}

type groupVersionResourceKind struct {
	gvr  *schema.GroupVersionResource
	kind string
}

type K8sWorkloadMappingConfig struct {
	NameAttr              string   `mapstructure:"name_attr"`
	NamespaceAttr         string   `mapstructure:"namespace_attr"`
	AddressAttr           string   `mapstructure:"address_attr"`
	WorkloadNameAttr      string   `mapstructure:"workload_name_attr"`
	WorkloadNamespaceAttr string   `mapstructure:"workload_namespace_attr"`
	WorkloadTypeAttr      string   `mapstructure:"workload_type_attr"`
	ExpectedTypes         []string `mapstructure:"expected_types"`
	PreferOwnerForPods    bool     `mapstructure:"prefer_owner_for_pods"`
	context               statementContext
}

type statementContext string

const (
	ResourceContext  statementContext = "resource"
	ScopeContext     statementContext = "scope"
	MetricContext    statementContext = "metric"
	DataPointContext statementContext = "datapoint"
)

func (c *Config) Validate() error {
	if len(c.WorkloadMappings) == 0 {
		return fmt.Errorf("workload_mappings cannot be empty")
	}

	validObjects, err := c.getValidObjects()
	if err != nil {
		return err
	}

	for _, mapping := range c.WorkloadMappings {
		if mapping == nil {
			return fmt.Errorf("workload_mapping cannot be nil")
		}
		if err := mapping.validate(validObjects, c.mappedExpectedTypes); err != nil {
			return err
		}
	}

	return nil
}

func (m *K8sWorkloadMappingConfig) validate(validObjects map[string][]groupVersionResourceKind, mappedExpectedTypes map[string]groupVersionResourceKind) error {
	if m.WorkloadTypeAttr == "" {
		return fmt.Errorf("workload_type_attr cannot be empty")
	}

	extractContext := func(attrName *string) statementContext {
		if before, after, found := strings.Cut(*attrName, "."); found {
			ctx := statementContext(before)
			if ctx == DataPointContext || ctx == MetricContext || ctx == ScopeContext || ctx == ResourceContext {
				*attrName = after
				return ctx
			}
		}
		return ""
	}

	nameCtx := extractContext(&m.NameAttr)
	workloadTypeCtx := extractContext(&m.WorkloadTypeAttr)
	namespaceCtx := extractContext(&m.NamespaceAttr)
	addressCtx := extractContext(&m.AddressAttr)
	workloadNameCtx := extractContext(&m.WorkloadNameAttr)
	workloadNamespaceCtx := extractContext(&m.WorkloadNamespaceAttr)

	switch {
	case len(m.NameAttr) == 0 && len(m.AddressAttr) == 0:
		return fmt.Errorf("name_attr or address_attr must be set")
	case len(m.NameAttr) > 0 && len(m.AddressAttr) > 0:
		return fmt.Errorf("name_attr and address_attr cannot be set at the same time")
	case len(m.NameAttr) > 0:
		m.context = nameCtx
	case len(m.AddressAttr) > 0:
		// Address mapping supports pods, services, and other workload types for Beyla-style resolution
		// Legacy restriction was only for pods and services, but we now support enhanced lookup
		m.context = addressCtx
	}

	if (m.context != workloadTypeCtx) || (m.NamespaceAttr != "" && namespaceCtx != m.context) || (m.WorkloadNameAttr != "" && workloadNameCtx != m.context) || (m.WorkloadNamespaceAttr != "" && workloadNamespaceCtx != m.context) {
		return fmt.Errorf("inconsistent context in workload mapping")
	}

	if m.context == "" {
		m.context = DataPointContext
	}

	if len(m.ExpectedTypes) == 0 {
		return fmt.Errorf("expected_types cannot be empty")
	}

	allExpectedTypes := m.ExpectedTypes
	if _, ok := validObjects[internal.ReplicaSetsWorkloadType]; ok && m.PreferOwnerForPods && slices.Contains(m.ExpectedTypes, internal.PodsWorkloadType) {
		// if we prefer owner for pods, we add replicasets to the expected types to be able to get their owner
		allExpectedTypes = append(allExpectedTypes, internal.ReplicaSetsWorkloadType)
	}

	for _, expectedType := range allExpectedTypes {
		if mappedExpectedTypes[expectedType] != (groupVersionResourceKind{}) {
			continue
		}

		gvrs, ok := validObjects[expectedType]
		if !ok {
			return fmt.Errorf("invalid expected_type '%s', valid types for this cluster are: %v", expectedType, slices.Collect(maps.Keys(validObjects)))
		}

		// take only the first one, as that seems to be the common behavior of the other processors
		mappedExpectedTypes[expectedType] = gvrs[0]
	}

	return nil
}

func (c *Config) getK8sClient() (kubernetes.Interface, error) {
	return initKubeClient(c.APIConfig)
}

func (c *Config) getDiscoveryClient() (discovery.ServerResourcesInterface, error) {
	client, err := initKubeClient(c.APIConfig)
	if err != nil {
		return nil, err
	}

	return client.Discovery(), nil
}

func (c *Config) getValidObjects() (map[string][]groupVersionResourceKind, error) {
	dc, err := c.getDiscoveryClient()
	if err != nil {
		return nil, err
	}

	res, err := dc.ServerPreferredResources()
	if err != nil {
		// Check if Partial result is returned from discovery client, that means some API servers have issues,
		// but we can still continue, as we check for the needed groups later in Validate function.
		if res != nil && !discovery.IsGroupDiscoveryFailedError(err) {
			return nil, err
		}
	}

	validObjects := make(map[string][]groupVersionResourceKind)

	for _, group := range res {
		split := strings.Split(group.GroupVersion, "/")
		if len(split) == 1 && group.GroupVersion == "v1" {
			split = []string{"", "v1"}
		}
		for _, resource := range group.APIResources {
			if resource.Kind == "" {
				// if a resource does not have "Kind", this processor is basically a no-op for it, so we can just skip it
				continue
			}
			validObjects[resource.Name] = append(validObjects[resource.Name],
				groupVersionResourceKind{
					gvr: &schema.GroupVersionResource{
						Group:    split[0],
						Version:  split[1],
						Resource: resource.Name,
					},
					kind: resource.Kind,
				})
		}
	}
	return validObjects, nil
}
