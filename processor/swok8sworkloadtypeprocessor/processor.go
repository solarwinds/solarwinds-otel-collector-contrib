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
	"context"
	"fmt"
	"iter"
	"maps"
	"net"
	"slices"

	"github.com/solarwinds/solarwinds-otel-collector-contrib/processor/swok8sworkloadtypeprocessor/internal"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/processor"
	"go.uber.org/zap"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"
)

type swok8sworkloadtypeProcessor struct {
	logger    *zap.Logger
	cancel    context.CancelFunc
	config    *Config
	settings  processor.Settings
	factory   informers.SharedInformerFactory
	informers map[string]cache.SharedIndexInformer
}

func (cp *swok8sworkloadtypeProcessor) getAttribute(attributes pcommon.Map, attrName string) string {
	if attrVal, ok := attributes.Get(attrName); ok && attrVal.Type() == pcommon.ValueTypeStr {
		return attrVal.Str()
	}
	return ""
}

type dataPointSlice[DP dataPoint] interface{ All() iter.Seq2[int, DP] }
type dataPoint interface{ Attributes() pcommon.Map }

func processDatapoints[DPS dataPointSlice[DP], DP dataPoint](cp *swok8sworkloadtypeProcessor, datapoints DPS, resourceAttributes pcommon.Map) {
	cp.logger.Debug("processDatapoints called")
	for _, dp := range datapoints.All() {
		cp.logger.Debug("Processing datapoint")
		processAttributesWithResourceFallback(cp, dp.Attributes(), DataPointContext, resourceAttributes)
	}
}

func processAttributes(cp *swok8sworkloadtypeProcessor, attributes pcommon.Map, statementContext statementContext) {
	processAttributesWithResourceFallback(cp, attributes, statementContext, pcommon.NewMap())
}

func processAttributesWithResourceFallback(cp *swok8sworkloadtypeProcessor, attributes pcommon.Map, statementContext statementContext, resourceAttributes pcommon.Map) {
	cp.logger.Debug("processAttributes called",
		zap.String("context", string(statementContext)),
		zap.Int("workloadMappingsCount", len(cp.config.WorkloadMappings)))

	for i, workloadMapping := range cp.config.WorkloadMappings {
		cp.logger.Debug("Checking workload mapping",
			zap.Int("index", i),
			zap.String("context", string(statementContext)),
			zap.String("mappingContext", string(workloadMapping.context)))

		if workloadMapping.context != statementContext {
			cp.logger.Debug("Skipping workload mapping due to context mismatch",
				zap.String("expectedContext", string(statementContext)),
				zap.String("mappingContext", string(workloadMapping.context)))
			continue
		}

		cp.logger.Debug("Processing workload mapping",
			zap.String("context", string(statementContext)),
			zap.String("nameAttr", workloadMapping.NameAttr),
			zap.String("addressAttr", workloadMapping.AddressAttr),
			zap.String("namespaceAttr", workloadMapping.NamespaceAttr),
			zap.String("workloadTypeAttr", workloadMapping.WorkloadTypeAttr),
			zap.Strings("expectedTypes", workloadMapping.ExpectedTypes),
			zap.Bool("preferOwnerForPods", workloadMapping.PreferOwnerForPods))

		if cp.getAttribute(attributes, workloadMapping.WorkloadTypeAttr) != "" {
			// Skip if the workload type attribute is already set
			cp.logger.Debug("Skipping workload mapping - workload type already set",
				zap.String("workloadTypeAttr", workloadMapping.WorkloadTypeAttr),
				zap.String("existingValue", cp.getAttribute(attributes, workloadMapping.WorkloadTypeAttr)))
			continue
		}

		// Log input attributes for debugging
		allAttrs := make(map[string]interface{})
		attributes.Range(func(k string, v pcommon.Value) bool {
			allAttrs[k] = v.AsString()
			return true
		})
		cp.logger.Debug("Input attributes for workload mapping", zap.Any("attributes", allAttrs))

		var res internal.LookupResult
		if workloadMapping.NameAttr != "" {
			res = cp.lookupWorkloadTypeByNameAttrWithResourceFallback(workloadMapping, attributes, resourceAttributes)
		} else if workloadMapping.AddressAttr != "" {
			res = cp.lookupWorkloadTypeByAddressAttrWithResourceFallback(workloadMapping, attributes, resourceAttributes)
		} else {
			cp.logger.Error("Unexpected workload mapping configuration")
			continue
		}

		cp.logger.Debug("Lookup result",
			zap.String("resultKind", res.Kind),
			zap.String("resultName", res.Name),
			zap.String("resultNamespace", res.Namespace))

		if res.Kind != "" {
			attributes.PutStr(workloadMapping.WorkloadTypeAttr, res.Kind)
			cp.logger.Debug("Set workload type attribute",
				zap.String("attribute", workloadMapping.WorkloadTypeAttr),
				zap.String("value", res.Kind))
		}
		if res.Name != "" && workloadMapping.WorkloadNameAttr != "" {
			attributes.PutStr(workloadMapping.WorkloadNameAttr, res.Name)
			cp.logger.Debug("Set workload name attribute",
				zap.String("attribute", workloadMapping.WorkloadNameAttr),
				zap.String("value", res.Name))
		}
		if res.Namespace != "" && workloadMapping.WorkloadNamespaceAttr != "" {
			attributes.PutStr(workloadMapping.WorkloadNamespaceAttr, res.Namespace)
			cp.logger.Debug("Set workload namespace attribute",
				zap.String("attribute", workloadMapping.WorkloadNamespaceAttr),
				zap.String("value", res.Namespace))
		}
	}
}

func (cp *swok8sworkloadtypeProcessor) lookupWorkloadTypeByNameAttr(workloadMapping *K8sWorkloadMappingConfig, attributes pcommon.Map) internal.LookupResult {
	return cp.lookupWorkloadTypeByNameAttrWithResourceFallback(workloadMapping, attributes, pcommon.NewMap())
}

func (cp *swok8sworkloadtypeProcessor) lookupWorkloadTypeByNameAttrWithResourceFallback(workloadMapping *K8sWorkloadMappingConfig, attributes pcommon.Map, resourceAttributes pcommon.Map) internal.LookupResult {
	name := cp.getAttribute(attributes, workloadMapping.NameAttr)
	if name == "" {
		cp.logger.Debug("Name attribute not found or empty",
			zap.String("nameAttr", workloadMapping.NameAttr))
		return internal.EmptyLookupResult
	}

	// Try to get namespace from attributes first, then fallback to resource attributes
	namespace := cp.getAttribute(attributes, workloadMapping.NamespaceAttr)
	if namespace == "" && workloadMapping.NamespaceAttr != "" {
		namespace = cp.getAttribute(resourceAttributes, workloadMapping.NamespaceAttr)
		if namespace != "" {
			cp.logger.Debug("Using namespace from resource attributes as fallback",
				zap.String("namespaceAttr", workloadMapping.NamespaceAttr),
				zap.String("namespace", namespace))
		}
	}

	cp.logger.Debug("Looking up workload by name and namespace",
		zap.String("name", name),
		zap.String("namespace", namespace),
		zap.String("nameAttr", workloadMapping.NameAttr),
		zap.String("namespaceAttr", workloadMapping.NamespaceAttr),
		zap.Strings("expectedTypes", workloadMapping.ExpectedTypes),
		zap.Bool("preferOwnerForPods", workloadMapping.PreferOwnerForPods))

	return internal.LookupWorkloadKindByNameAndNamespace(name, namespace, workloadMapping.ExpectedTypes, cp.logger, cp.informers, workloadMapping.PreferOwnerForPods)
}

func (cp *swok8sworkloadtypeProcessor) lookupWorkloadTypeByAddressAttr(workloadMapping *K8sWorkloadMappingConfig, attributes pcommon.Map) internal.LookupResult {
	return cp.lookupWorkloadTypeByAddressAttrWithResourceFallback(workloadMapping, attributes, pcommon.NewMap())
}

func (cp *swok8sworkloadtypeProcessor) lookupWorkloadTypeByAddressAttrWithResourceFallback(workloadMapping *K8sWorkloadMappingConfig, attributes pcommon.Map, resourceAttributes pcommon.Map) internal.LookupResult {
	addr := cp.getAttribute(attributes, workloadMapping.AddressAttr)
	if addr == "" {
		cp.logger.Debug("Address attribute not found or empty",
			zap.String("addressAttr", workloadMapping.AddressAttr))
		return internal.EmptyLookupResult
	}
	host := internal.ExtractHostFromAddress(addr)

	cp.logger.Debug("Looking up workload by address",
		zap.String("address", addr),
		zap.String("extractedHost", host),
		zap.String("addressAttr", workloadMapping.AddressAttr),
		zap.Strings("expectedTypes", workloadMapping.ExpectedTypes),
		zap.Bool("preferOwnerForPods", workloadMapping.PreferOwnerForPods))

	if net.ParseIP(host) != nil {
		cp.logger.Debug("Host is an IP address, looking up by IP", zap.String("ip", host))
		return internal.LookupWorkloadKindByIp(host, workloadMapping.ExpectedTypes, cp.logger, cp.informers, workloadMapping.PreferOwnerForPods)
	} else {
		// Try to get namespace from attributes first, then fallback to resource attributes
		namespace := cp.getAttribute(attributes, workloadMapping.NamespaceAttr)
		if namespace == "" && workloadMapping.NamespaceAttr != "" {
			namespace = cp.getAttribute(resourceAttributes, workloadMapping.NamespaceAttr)
			if namespace != "" {
				cp.logger.Debug("Using namespace from resource attributes as fallback for address lookup",
					zap.String("namespaceAttr", workloadMapping.NamespaceAttr),
					zap.String("namespace", namespace))
			}
		}

		cp.logger.Debug("Host is a hostname, looking up by hostname",
			zap.String("hostname", host),
			zap.String("namespace", namespace),
			zap.String("namespaceAttr", workloadMapping.NamespaceAttr))
		return internal.LookupWorkloadKindByHostname(host, namespace, workloadMapping.ExpectedTypes, cp.logger, cp.informers, workloadMapping.PreferOwnerForPods)
	}
}

func (cp *swok8sworkloadtypeProcessor) processMetrics(ctx context.Context, md pmetric.Metrics) (pmetric.Metrics, error) {
	for _, rm := range md.ResourceMetrics().All() {
		processAttributes(cp, rm.Resource().Attributes(), ResourceContext)
		for _, sm := range rm.ScopeMetrics().All() {
			processAttributes(cp, sm.Scope().Attributes(), ScopeContext)
			for _, m := range sm.Metrics().All() {
				processAttributes(cp, m.Metadata(), MetricContext)

				switch m.Type() {
				case pmetric.MetricTypeGauge:
					processDatapoints(cp, m.Gauge().DataPoints(), rm.Resource().Attributes())
				case pmetric.MetricTypeSum:
					processDatapoints(cp, m.Sum().DataPoints(), rm.Resource().Attributes())
				case pmetric.MetricTypeHistogram:
					processDatapoints(cp, m.Histogram().DataPoints(), rm.Resource().Attributes())
				case pmetric.MetricTypeExponentialHistogram:
					processDatapoints(cp, m.ExponentialHistogram().DataPoints(), rm.Resource().Attributes())
				case pmetric.MetricTypeSummary:
					processDatapoints(cp, m.Summary().DataPoints(), rm.Resource().Attributes())
				default:
					cp.logger.Debug("Unsupported metric type", zap.Any("metricType", m.Type()))
					continue
				}
			}
		}
	}

	return md, nil
}

func (cp *swok8sworkloadtypeProcessor) Start(ctx context.Context, _ component.Host) error {
	cp.logger.Info("Starting swok8sworkloadtype processor")
	cancelCtx, cancelFunc := context.WithCancel(context.Background())
	cp.cancel = cancelFunc

	client, err := cp.config.getK8sClient()
	if err != nil {
		return err
	}

	cp.factory = informers.NewSharedInformerFactory(client, cp.config.WatchSyncPeriod)

	copyOwners := slices.ContainsFunc(cp.config.WorkloadMappings, func(mapping *K8sWorkloadMappingConfig) bool {
		return mapping.PreferOwnerForPods && slices.Contains(mapping.ExpectedTypes, internal.PodsWorkloadType)
	})

	cp.informers = make(map[string]cache.SharedIndexInformer, len(cp.config.mappedExpectedTypes))
	for workloadType, mappedWorkloadType := range cp.config.mappedExpectedTypes {
		informer, err := cp.factory.ForResource(*mappedWorkloadType.gvr)
		if err != nil {
			return fmt.Errorf("error creating informer for workload type '%s': %w", workloadType, err)
		}
		cp.informers[workloadType] = informer.Informer()

		if mappedWorkloadType.kind == internal.PodKind {
			cp.informers[workloadType].SetTransform(internal.PodTransformFunc(cp.logger, copyOwners))
			cp.informers[workloadType].AddIndexers(internal.PodIpIndexer(cp.logger))
		} else if mappedWorkloadType.kind == internal.ServiceKind {
			cp.informers[workloadType].SetTransform(internal.ServiceTransformFunc(cp.logger))
			cp.informers[workloadType].AddIndexers(internal.ServiceIpIndexer(cp.logger))
		} else if mappedWorkloadType.kind == internal.ReplicaSetKind {
			cp.informers[workloadType].SetTransform(internal.GenericTransformFunc(cp.logger, mappedWorkloadType.kind, copyOwners))
		} else {
			cp.informers[workloadType].SetTransform(internal.GenericTransformFunc(cp.logger, mappedWorkloadType.kind, false))
		}
	}

	cp.logger.Info("Starting informers", zap.Any("informers", slices.Collect(maps.Keys(cp.informers))))

	cp.factory.Start(cancelCtx.Done())
	initialSyncResult := cp.factory.WaitForCacheSync(cancelCtx.Done())
	for v, ok := range initialSyncResult {
		if !ok {
			return fmt.Errorf("caches failed to sync: %v", v)
		}
	}
	cp.logger.Info("All informers have synced successfully")

	return nil
}

func (cp *swok8sworkloadtypeProcessor) Shutdown(_ context.Context) error {
	if cp.cancel != nil {
		cp.cancel()
	}
	if cp.factory != nil {
		cp.factory.Shutdown()
	}
	return nil
}
