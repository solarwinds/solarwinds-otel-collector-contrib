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

// Source: https://github.com/open-telemetry/opentelemetry-collector-contrib
// Changes customizing the original source code

package swok8sdiscovery

import (
	"context"
	"strings"

	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
)

// discoverDatabasesByDomains matches ExternalName services to domain rules.
func (r *swok8sdiscoveryReceiver) discoverDatabasesByDomains(ctx context.Context, pods []corev1.Pod, services []corev1.Service) {
	if r.config.Database == nil || len(r.config.Database.DomainRules) == 0 {
		r.setting.Logger.Info("Domain-based database discovery skipped",
			zap.Bool("database_config_present", r.config.Database != nil),
			zap.Int("domain_rules_count", 0))
		return
	}

	r.setting.Logger.Info("Starting domain-based database discovery",
		zap.Int("domain_rules_count", len(r.config.Database.DomainRules)),
		zap.Int("total_services", len(services)))

	if r.setting.Logger.Core().Enabled(zap.DebugLevel) {
		for i, rule := range r.config.Database.DomainRules {
			r.setting.Logger.Debug("Domain rule configured",
				zap.Int("rule_index", i),
				zap.String("database_type", rule.DatabaseType),
				zap.Strings("patterns", rule.Patterns),
				zap.Strings("domain_hints", rule.DomainHints))
		}
	}

	externalNameServices := 0
	matchedServices := 0

	for _, svc := range services {
		if svc.Spec.ExternalName == "" { // Only process ExternalName services for domain-based detection
			continue
		}
		externalNameServices++
		external := svc.Spec.ExternalName

		r.setting.Logger.Debug("Evaluating ExternalName service",
			zap.String("service", svc.Name),
			zap.String("namespace", svc.Namespace),
			zap.String("external_name", external))

		matchingRules := make([]*DomainRule, 0)

		var matchedRule *DomainRule
		triedPatterns := []string{}

		for i := range r.config.Database.DomainRules {
			rule := r.config.Database.DomainRules[i]
			for _, rx := range rule.PatternsCompiled {
				triedPatterns = append(triedPatterns, rx.String())
				if rx.MatchString(external) {
					matchingRules = append(matchingRules, rule)
					break
				}
			}
		}

		if len(matchingRules) == 1 {
			matchedRule = matchingRules[0]
		} else if len(matchingRules) > 1 {

			lower_external := strings.ToLower(external)
			lower_name := strings.ToLower(svc.Name)

			for _, rule := range matchingRules {
				if strings.Contains(lower_external, rule.DatabaseType) ||
					strings.Contains(lower_name, rule.DatabaseType) {
					matchedRule = rule
					break
				}

				if len(rule.DomainHints) > 0 {
					for _, hint := range rule.DomainHints {
						if strings.Contains(lower_external, hint) ||
							strings.Contains(lower_name, hint) {
							matchedRule = rule
							break
						}
					}

					if matchedRule != nil {
						break
					}
				}
			}
		}

		if matchedRule == nil {
			r.setting.Logger.Debug("ExternalName service did not match any domain rule",
				zap.String("service", svc.Name),
				zap.String("namespace", svc.Namespace),
				zap.String("external_name", external),
				zap.Int("patterns_tried", len(triedPatterns)))
			continue
		}

		matchedServices++
		r.setting.Logger.Debug("Matched ExternalName service to database rule",
			zap.String("service", svc.Name),
			zap.String("namespace", svc.Namespace),
			zap.String("external_name", external),
			zap.String("database_type", matchedRule.DatabaseType))

		var wKind, wName string
		if len(svc.Spec.Selector) > 0 {
			matchedPods := findServicePods(&svc, pods)
			if len(matchedPods) > 0 {
				wKind, wName, _ = r.resolveWorkloadForPod(ctx, &matchedPods[0])
			}
		}

		// get list of Service svcTargetPorts and Target svcTargetPorts
		var svcPorts []int32
		for _, p := range svc.Spec.Ports {
			svcPorts = append(svcPorts, p.Port)
		}

		// Validate port count: only create entity when exactly one port is present
		if len(svcPorts) == 0 {
			r.setting.Logger.Debug("Skipping ExternalName service with no ports",
				zap.String("service", svc.Name),
				zap.String("namespace", svc.Namespace),
				zap.String("external_name", external),
				zap.String("database_type", matchedRule.DatabaseType))
			continue
		}

		if len(svcPorts) != 1 {
			// TODO: Consider implementing direct connection attempts to identify the actual database port
			// when multiple ports are detected. We could attempt database-specific handshakes on each
			// port to automatically determine which one is the database port, eliminating ambiguity.
			r.setting.Logger.Debug("Skipping ExternalName service with multiple ports",
				zap.String("service", svc.Name),
				zap.String("namespace", svc.Namespace),
				zap.String("external_name", external),
				zap.String("database_type", matchedRule.DatabaseType),
				zap.Int32s("ports", svcPorts),
				zap.String("reason", "Entity creation requires exactly one port."))
			continue
		}

		//  service with external endpoint can be detected by other discoveries we mark discovery.id as `external`
		r.setting.Logger.Debug("Publishing external database discovery event",
			zap.String("database_type", matchedRule.DatabaseType),
			zap.String("namespace", svc.Namespace),
			zap.String("endpoint", external),
			zap.String("workload_kind", wKind),
			zap.String("workload_name", wName),
			zap.Int32s("ports", svcPorts))

		r.publishDatabaseEvent(ctx, "external", databaseEvent{
			DatabaseType: matchedRule.DatabaseType,
			Namespace:    svc.Namespace,
			Ports:        svcPorts,
			Endpoint:     external,
			WorkloadKind: wKind,
			WorkloadName: wName,
		})
	}

	r.setting.Logger.Info("Completed domain-based database discovery",
		zap.Int("domain_rules_count", len(r.config.Database.DomainRules)),
		zap.Int("external_name_services", externalNameServices),
		zap.Int("matched_services", matchedServices))
}

// findServicePods returns pods whose labels satisfy the service selector (best effort; requires caller to pass pod list).
func findServicePods(svc *corev1.Service, pods []corev1.Pod) []corev1.Pod {
	if svc == nil || len(svc.Spec.Selector) == 0 {
		return nil
	}
	var out []corev1.Pod
	for _, p := range pods {
		if p.Namespace != svc.Namespace {
			continue
		}
		if selectorMatches(svc.Spec.Selector, p.GetLabels()) {
			out = append(out, p)
		}
	}
	return out
}
