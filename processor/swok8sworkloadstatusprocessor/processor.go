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

package swok8sworkloadstatusprocessor // import "github.com/solarwinds/solarwinds-otel-collector-contrib/processor/swok8sworkloadstatusprocessor"

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"

	"github.com/solarwinds/solarwinds-otel-collector-contrib/processor/swok8sworkloadstatusprocessor/internal"
)

type swok8sworkloadstatusprocessor struct {
	logger      *zap.Logger
	podInformer cache.SharedIndexInformer
	cancel      context.CancelFunc
	config      *Config
	factory     informers.SharedInformerFactory
}

func (p *swok8sworkloadstatusprocessor) Start(_ context.Context, _ component.Host) error {

	p.logger.Info("Starting swok8sworkloadstatus processor")
	cancelCtx, cancelFunc := context.WithCancel(context.Background())
	p.cancel = cancelFunc

	client, err := p.config.getK8sClient()
	if err != nil {
		return err
	}

	p.factory = informers.NewSharedInformerFactory(client, p.config.WatchSyncPeriod)

	p.podInformer = p.factory.Core().V1().Pods().Informer()
	p.podInformer.SetTransform(internal.PodTransformFunc(p.logger))
	p.podInformer.AddIndexers(internal.WorkloadOwnerIndexer(p.logger))

	p.logger.Info("Starting pod informer")

	p.factory.Start(cancelCtx.Done())
	initialSyncResult := p.factory.WaitForCacheSync(cancelCtx.Done())
	for v, ok := range initialSyncResult {
		if !ok {
			return fmt.Errorf("caches failed to sync: %v", v)
		}
	}
	p.logger.Info("All informers have synced successfully")

	return nil
}

func (p *swok8sworkloadstatusprocessor) Shutdown(_ context.Context) error {
	if p.cancel != nil {
		p.cancel()
	}
	if p.factory != nil {
		p.factory.Shutdown()
	}
	return nil
}

// processLogs iterates logs, extracts pod and owner info from k8s manifests contained in log bodies
// or from already enriched attributes.
func (p *swok8sworkloadstatusprocessor) processLogs(_ context.Context, ld plog.Logs) (plog.Logs, error) {
	for _, rl := range ld.ResourceLogs().All() {
		for _, sl := range rl.ScopeLogs().All() {
			for _, lr := range sl.LogRecords().All() {
				attrs := lr.Attributes()
				kindAttr, ok := attrs.Get("k8s.object.kind")
				if !ok {
					continue
				}

				// We only operate on string bodies; skip otherwise early.
				if lr.Body().Type() != pcommon.ValueTypeStr {
					continue
				}

				bodyStr := lr.Body().Str()
				kind := kindAttr.Str()

				// Process by workload kind
				switch kind {
				case internal.DeploymentKind:
					if status, ok := p.computeDeploymentStatus(bodyStr); ok {
						attrs.PutStr("sw.k8s.deployment.status", status)
					}
				case internal.DaemonSetKind, internal.StatefulSetKind:
					if wlName, wlNamespace, ok := extractNameNamespace(bodyStr); ok {
						if status, ok := p.computePodAggregateStatus(wlName, wlNamespace, kind); ok {
							attrs.PutStr("sw.k8s."+strings.ToLower(kind)+".status", status)
						}
					}
				}
			}
		}
	}
	return ld, nil
}

// computeDeploymentStatus parses a Deployment manifest JSON string and returns the derived status.
// Returned bool indicates if parsing / computation succeeded.
func (p *swok8sworkloadstatusprocessor) computeDeploymentStatus(body string) (string, bool) {
	type deploymentManifest struct {
		Status struct {
			Replicas   int `json:"replicas"`
			Conditions []struct {
				Type   string `json:"type"`
				Status string `json:"status"`
			} `json:"conditions"`
		} `json:"status"`
	}
	var manifest deploymentManifest
	if err := json.Unmarshal([]byte(body), &manifest); err != nil {
		return "", false
	}
	if manifest.Status.Replicas == 0 {
		return "", true
	}
	available := false
	progressing := false
	for _, cond := range manifest.Status.Conditions {
		if cond.Type == "Available" && cond.Status == "True" {
			available = true
		}
		if cond.Type == "Progressing" && cond.Status == "True" {
			progressing = true
		}
	}
	switch {
	case available && progressing:
		return "AVAILABLE", true
	case available && !progressing:
		return "OUTDATED", true
	case !available:
		return "UNAVAILABLE", true
	default:
		return "OTHER", true
	}
}

// extractNameNamespace extracts metadata.name and metadata.namespace from a manifest with metadata only.
func extractNameNamespace(body string) (name, namespace string, ok bool) {
	var manifest struct {
		Metadata struct {
			Name      string `json:"name"`
			Namespace string `json:"namespace"`
		} `json:"metadata"`
	}
	if err := json.Unmarshal([]byte(body), &manifest); err != nil {
		return "", "", false
	}
	return manifest.Metadata.Name, manifest.Metadata.Namespace, true
}

// computePodAggregateStatus derives DaemonSet/StatefulSet overall status from their pods.
// Formula:
//
//	"" if totalPods == 0
//	"RUNNING" if (running + succeeded) == totalPods
//	"FAILED" if (failed + unknown) == totalPods
//	"PENDING" otherwise
func (p *swok8sworkloadstatusprocessor) computePodAggregateStatus(workloadName, namespace, kind string) (string, bool) {
	pods := p.getPodsForWorkload(workloadName, namespace, kind)
	total := len(pods)
	if total == 0 {
		return "", true
	}
	var running, succeeded, failed, unknown int
	for _, pod := range pods {
		switch pod.Status.Phase {
		case corev1.PodRunning:
			running++
		case corev1.PodSucceeded:
			succeeded++
		case corev1.PodFailed:
			failed++
		case corev1.PodUnknown:
			unknown++
		}
	}
	switch {
	case running+succeeded == total:
		return "RUNNING", true
	case failed+unknown == total:
		return "FAILED", true
	default:
		return "PENDING", true
	}
}

// getPodsForWorkload finds pods for a given workload (DaemonSet/StatefulSet) by name/namespace/kind using informers.
func (p *swok8sworkloadstatusprocessor) getPodsForWorkload(workloadName, namespace, kind string) []*corev1.Pod {
	var pods []*corev1.Pod
	if p.podInformer == nil {
		return pods
	}

	indexKey := fmt.Sprintf("%s/%s/%s", namespace, kind, workloadName)
	objs, err := p.podInformer.GetIndexer().ByIndex(internal.WorkloadOwnerIndex, indexKey)
	if err != nil {
		p.logger.Error("Failed to lookup pods by workload owner index",
			zap.Error(err),
			zap.String("workloadName", workloadName),
			zap.String("namespace", namespace),
			zap.String("kind", kind))
		return pods
	}

	for _, obj := range objs {
		if pod, ok := obj.(*corev1.Pod); ok {
			pods = append(pods, pod)
		}
	}
	return pods
}
