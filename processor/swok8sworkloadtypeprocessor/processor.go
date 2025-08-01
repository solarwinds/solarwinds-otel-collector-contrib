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

func processDatapoints[DPS dataPointSlice[DP], DP dataPoint](cp *swok8sworkloadtypeProcessor, datapoints DPS) {
	for _, dp := range datapoints.All() {
		processAttributes(cp, dp.Attributes(), DataPointContext)
	}
}

func processAttributes(cp *swok8sworkloadtypeProcessor, attributes pcommon.Map, statementContext statementContext) {
	for _, workloadMapping := range cp.config.WorkloadMappings {
		if workloadMapping.context != statementContext {
			continue
		}

		if cp.getAttribute(attributes, workloadMapping.WorkloadTypeAttr) != "" {
			// Skip if the workload type attribute is already set
			continue
		}

		var res internal.LookupResult
		if workloadMapping.NameAttr != "" {
			res = cp.lookupWorkloadTypeByNameAttr(workloadMapping, attributes)
		} else if workloadMapping.AddressAttr != "" {
			res = cp.lookupWorkloadTypeByAddressAttr(workloadMapping, attributes)
		} else {
			cp.logger.Error("Unexpected workload mapping configuration")
			continue
		}
		if res.Kind != "" {
			attributes.PutStr(workloadMapping.WorkloadTypeAttr, res.Kind)
		}
		if res.Name != "" && workloadMapping.WorkloadNameAttr != "" {
			attributes.PutStr(workloadMapping.WorkloadNameAttr, res.Name)
		}
		if res.Namespace != "" && workloadMapping.WorkloadNamespaceAttr != "" {
			attributes.PutStr(workloadMapping.WorkloadNamespaceAttr, res.Namespace)
		}
	}
}

func (cp *swok8sworkloadtypeProcessor) lookupWorkloadTypeByNameAttr(workloadMapping *K8sWorkloadMappingConfig, attributes pcommon.Map) internal.LookupResult {
	name := cp.getAttribute(attributes, workloadMapping.NameAttr)
	if name == "" {
		return internal.EmptyLookupResult
	}
	namespace := cp.getAttribute(attributes, workloadMapping.NamespaceAttr)

	return internal.LookupWorkloadKindByNameAndNamespace(name, namespace, workloadMapping.ExpectedTypes, cp.logger, cp.informers, workloadMapping.PreferOwnerForPods)
}

func (cp *swok8sworkloadtypeProcessor) lookupWorkloadTypeByAddressAttr(workloadMapping *K8sWorkloadMappingConfig, attributes pcommon.Map) internal.LookupResult {
	addr := cp.getAttribute(attributes, workloadMapping.AddressAttr)
	if addr == "" {
		return internal.EmptyLookupResult
	}
	host := internal.ExtractHostFromAddress(addr)

	if net.ParseIP(host) != nil {
		return internal.LookupWorkloadKindByIp(host, workloadMapping.ExpectedTypes, cp.logger, cp.informers, workloadMapping.PreferOwnerForPods)
	} else {
		namespace := cp.getAttribute(attributes, workloadMapping.NamespaceAttr)

		// If namespace attribute is configured but not provided in the data,
		// and the hostname doesn't contain namespace information, require namespace
		if workloadMapping.NamespaceAttr != "" && namespace == "" {
			// Check if the hostname contains structured namespace info
			_, namespaceFromHostname, _ := internal.ExtractNameAndNamespaceAndType(host)
			if namespaceFromHostname == "" {
				// No namespace in hostname and no namespace in attributes, but namespace is required
				return internal.EmptyLookupResult
			}
		}

		// Use enhanced OTEL instrumentation address lookup that can handle service names
		return internal.LookupWorkloadByOtelInstrumentationAddress(addr, namespace, workloadMapping.ExpectedTypes, cp.logger, cp.informers, workloadMapping.PreferOwnerForPods)
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
					processDatapoints(cp, m.Gauge().DataPoints())
				case pmetric.MetricTypeSum:
					processDatapoints(cp, m.Sum().DataPoints())
				case pmetric.MetricTypeHistogram:
					processDatapoints(cp, m.Histogram().DataPoints())
				case pmetric.MetricTypeExponentialHistogram:
					processDatapoints(cp, m.ExponentialHistogram().DataPoints())
				case pmetric.MetricTypeSummary:
					processDatapoints(cp, m.Summary().DataPoints())
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
