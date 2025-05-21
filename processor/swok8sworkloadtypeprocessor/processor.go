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
	"slices"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/processor"
	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
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
		attributes := dp.Attributes()
		for _, workloadMapping := range cp.config.WorkloadMappings {
			if cp.getAttribute(attributes, workloadMapping.WorkloadTypeAttr) != "" {
				// Skip if the workload type attribute is already set
				continue
			}
			name := cp.getAttribute(attributes, workloadMapping.NameAttr)
			if name == "" {
				continue
			}
			namespace := cp.getAttribute(attributes, workloadMapping.NamespaceAttr)
			var workloadKey string
			if namespace != "" {
				workloadKey = fmt.Sprintf("%s/%s", namespace, name)
			} else {
				workloadKey = name
			}

			for _, workloadType := range workloadMapping.ExpectedTypes {
				workload, exists, err := cp.informers[workloadType].GetStore().GetByKey(workloadKey)
				if err != nil {
					cp.logger.Error("Error getting workload from cache", zap.String("workloadType", workloadType), zap.String("workloadKey", workloadKey), zap.Error(err))
					continue
				}
				if exists {
					workloadObject, ok := workload.(runtime.Object)
					if !ok {
						cp.logger.Error("Unexpected workload object type in cache", zap.String("workloadType", workloadType), zap.String("workloadKey", workloadKey), zap.String("workloadObjectType", fmt.Sprintf("%T", workload)))
						break
					}
					kind := workloadObject.GetObjectKind().GroupVersionKind().Kind
					if kind != "" {
						attributes.PutStr(workloadMapping.WorkloadTypeAttr, kind)
					} else {
						cp.logger.Debug("Workload has no kind", zap.String("workloadType", workloadType), zap.String("workloadKey", workloadKey))
					}
					break
				}
			}
		}
	}
}

func (cp *swok8sworkloadtypeProcessor) processMetrics(ctx context.Context, md pmetric.Metrics) (pmetric.Metrics, error) {
	for _, rm := range md.ResourceMetrics().All() {
		for _, sm := range rm.ScopeMetrics().All() {
			for _, m := range sm.Metrics().All() {
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

	cp.informers = make(map[string]cache.SharedIndexInformer, len(cp.config.mappedExpectedTypes))
	for workloadType, mappedWorkloadType := range cp.config.mappedExpectedTypes {
		informer, err := cp.factory.ForResource(*mappedWorkloadType.gvr)
		if err != nil {
			return fmt.Errorf("error creating informer for workload type '%s': %w", workloadType, err)
		}
		cp.informers[workloadType] = informer.Informer()
		cp.informers[workloadType].SetTransform(func(obj any) (any, error) {
			workload, ok := obj.(metav1.Object)
			if !ok {
				cp.logger.Error("Received an unexpected workload object type", zap.String("workloadObjectType", fmt.Sprintf("%T", obj)))
				return obj, nil
			}
			cp.logger.Debug("Received workload", zap.String("workloadName", workload.GetName()), zap.String("workloadNamespace", workload.GetNamespace()))
			return &metav1.PartialObjectMetadata{
				TypeMeta: metav1.TypeMeta{
					Kind: mappedWorkloadType.kind,
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      workload.GetName(),
					Namespace: workload.GetNamespace(),
				},
			}, nil
		})
	}

	cp.logger.Info("Starting informers", zap.Any("informers", slices.Collect(maps.Keys(cp.informers))))

	cp.factory.Start(cancelCtx.Done())
	initialSyncResult := cp.factory.WaitForCacheSync(cancelCtx.Done())
	for v, ok := range initialSyncResult {
		if !ok {
			return fmt.Errorf("caches failed to sync: %v", v)
		}
	}

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
