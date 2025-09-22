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

	corev1 "k8s.io/api/core/v1"
)

// discoverDatabasesByDomains matches ExternalName services to domain rules.
func (r *swok8sdiscoveryReceiver) discoverDatabasesByDomains(ctx context.Context, pods []corev1.Pod, services []corev1.Service) {
	if r.config.Database == nil || len(r.config.Database.DomainRules) == 0 {
		return
	}
	for _, svc := range services {
		if svc.Spec.ExternalName == "" { // Only process ExternalName services for domain-based detection
			continue
		}
		external := svc.Spec.ExternalName

		matchingRules := make([]*DomainRule, 0)

		var matchedRule *DomainRule
		for i := range r.config.Database.DomainRules {
			rule := &r.config.Database.DomainRules[i]
			for _, rx := range rule.PatternsCompiled {
				if rx.MatchString(external) {
					matchingRules = append(matchingRules, rule)
					break
				}
			}
		}

		if len(matchingRules) == 1 {
			matchedRule = matchingRules[0]
		} else {

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
			continue
		}

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

		//  service with external endpoint can be detected by other discoveries we mark discovery.id as `external`
		r.publishDatabaseEvent(ctx, "external", databaseEvent{
			DatabaseType: matchedRule.DatabaseType,
			Namespace:    svc.Namespace,
			Ports:        svcPorts,
			Endpoint:     external,
			WorkloadKind: wKind,
			WorkloadName: wName,
		})
	}
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
