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

package swok8sdiscovery

import (
	"context"
	"strings"

	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
)

// discoverDatabases matches containers to database image rules, resolves ports and associated services.
func (r *swok8sdiscoveryReceiver) discoverDatabasesByImages(ctx context.Context, pods []corev1.Pod, services []corev1.Service) {

	if r.config.Database == nil || len(r.config.Database.ImageRules) == 0 {
		r.setting.Logger.Info("Image-based database discovery skipped",
			zap.Bool("database_config_present", r.config.Database != nil),
			zap.Int("image_rules_count", 0))
		return
	}

	r.setting.Logger.Info("Starting image-based database discovery",
		zap.Int("image_rules_count", len(r.config.Database.ImageRules)),
		zap.Int("pods_to_scan", len(pods)))

	if r.setting.Logger.Core().Enabled(zap.DebugLevel) {
		for i, rule := range r.config.Database.ImageRules {
			r.setting.Logger.Debug("Image rule configured",
				zap.Int("rule_index", i),
				zap.String("database_type", rule.DatabaseType),
				zap.Strings("patterns", rule.Patterns),
				zap.Int32("default_port", rule.DefaultPort))
		}
	}

	// Reduce scope of services to those in each namespace for faster lookup.
	svcByNamespace := map[string][]corev1.Service{}
	for _, svc := range services {
		svcByNamespace[svc.Namespace] = append(svcByNamespace[svc.Namespace], svc)
	}

	// Track emitted workload+db+port combinations to avoid duplicate events while allowing distinct ports.
	// Key format: namespace|workloadKind|workloadName|dbType|ports(sorted,comma-separated)
	emittedWorkloads := make(map[string]struct{})
	matchedContainers := 0
	evaluatedContainers := 0

	for _, pod := range pods {
		r.setting.Logger.Debug("Evaluating pod for database discovery",
			zap.String("pod", pod.Name),
			zap.String("namespace", pod.Namespace),
			zap.Int("containers_count", len(pod.Spec.Containers)))

		for _, container := range pod.Spec.Containers {

			evaluatedContainers++

			r.setting.Logger.Debug("Evaluating container image",
				zap.String("pod", pod.Name),
				zap.String("namespace", pod.Namespace),
				zap.String("container", container.Name),
				zap.String("image", container.Image))

			matchedRule := (*ImageRule)(nil)
			triedPatterns := []string{}

			for i := range r.config.Database.ImageRules {
				rule := r.config.Database.ImageRules[i]
				for _, rx := range rule.PatternsCompiled {
					triedPatterns = append(triedPatterns, rx.String())
					if rx.MatchString(container.Image) {
						matchedRule = rule
						break
					}
				}
				if matchedRule != nil {
					break
				}
			}

			if matchedRule == nil {
				r.setting.Logger.Debug("Container image did not match any database rule",
					zap.String("pod", pod.Name),
					zap.String("namespace", pod.Namespace),
					zap.String("container", container.Name),
					zap.String("image", container.Image),
					zap.Int("patterns_tried", len(triedPatterns)))
				continue
			}

			matchedContainers++
			r.setting.Logger.Debug("Matched container image to database rule",
				zap.String("pod", pod.Name),
				zap.String("namespace", pod.Namespace),
				zap.String("container", container.Name),
				zap.String("image", container.Image),
				zap.String("database_type", matchedRule.DatabaseType))

			// Resolve ports
			ports := resolveContainerPorts(container, matchedRule.DefaultPort)

			// Try to match service exposing one of these ports by label selector (pod labels subset service selector)
			svcName, svcPorts := matchServiceForPod(pod, container, ports, svcByNamespace[pod.Namespace])

			// Resolve workload (top-level)
			wKind, wName, _ := r.resolveWorkloadForPod(ctx, &pod)

			// duplicate workload+db+port combo
			if wKind != "" && wName != "" {
				// Build stable port key (ports slice already reflects selected port(s))
				portKey := strings.Join(portsAsStrings(ports), ",")
				key := pod.Namespace + "|" + wKind + "|" + wName + "|" + matchedRule.DatabaseType + "|" + portKey
				if _, exists := emittedWorkloads[key]; exists {
					r.setting.Logger.Debug("Skipping duplicate database discovery event",
						zap.String("namespace", pod.Namespace),
						zap.String("workload_kind", wKind),
						zap.String("workload_name", wName),
						zap.String("database_type", matchedRule.DatabaseType))
					continue
				}
				emittedWorkloads[key] = struct{}{}
			}

			database := databaseEvent{
				DatabaseType: matchedRule.DatabaseType,
				Namespace:    pod.Namespace,
				Endpoint:     pod.Name,
				Ports:        ports,
				WorkloadKind: wKind,
				WorkloadName: wName,
			}

			if svcName != "" && len(svcPorts) != 0 {
				database.Endpoint = svcName
				database.Ports = svcPorts
			}

			r.setting.Logger.Debug("Publishing database discovery event",
				zap.String("database_type", database.DatabaseType),
				zap.String("namespace", database.Namespace),
				zap.String("endpoint", database.Endpoint),
				zap.String("workload_kind", database.WorkloadKind),
				zap.String("workload_name", database.WorkloadName),
				zap.Int32s("ports", database.Ports))

			r.publishDatabaseEvent(ctx, r.clusterUid, database)
		}
	}

	r.setting.Logger.Info("Completed image-based database discovery",
		zap.Int("evaluated_containers", evaluatedContainers),
		zap.Int("matched_containers", matchedContainers),
		zap.Int("non_matched_containers", evaluatedContainers-matchedContainers),
		zap.Int("events_published", len(emittedWorkloads)))
}

// resolveContainerPorts picks either the default port (if present) or all container ports.
func resolveContainerPorts(c corev1.Container, defaultPort int32) []int32 {
	var res []int32
	hasDefault := false
	for _, p := range c.Ports {
		res = append(res, p.ContainerPort)
		if p.ContainerPort == defaultPort {
			hasDefault = true
		}
	}
	if defaultPort != 0 && hasDefault {
		return []int32{defaultPort}
	}
	return res
}
