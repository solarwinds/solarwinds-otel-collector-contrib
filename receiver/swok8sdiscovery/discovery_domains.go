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

		l := r.setting.Logger.With(
			zap.String("service", svc.Name),
			zap.String("namespace", svc.Namespace),
			zap.String("external_name", external))

		l.Debug("Evaluating ExternalName service")

		matchingRules := make([]*DomainRule, 0)

		var matchedRule *DomainRule
		triedPatterns := []string{}

		// Find all rules whose patterns match the service's external name
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
			// If multiple rules match, apply tie-breaking logic:
			// prefer rules that match the service name or external name,
			// and then prefer rules with domain hints that match the service or external name.

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
			l.Debug("ExternalName service did not match any domain rule", zap.Int("patterns_tried", len(triedPatterns)))
			continue
		}

		matchedServices++
		l.Debug("Matched ExternalName service to database rule", zap.String("database_type", matchedRule.DatabaseType))

		database := databaseEvent{
			DatabaseType: matchedRule.DatabaseType,
			Name:         external,
			Address:      buildAddressWithPort(external, matchedRule.DefaultPort),
			Namespace:    svc.Namespace,
			WorkloadKind: kindService,
			WorkloadName: svc.Name,
		}

		r.setting.Logger.Debug("Publishing external database discovery event", zap.Any("database", database))

		//  service with external endpoint can be detected by other discoveries we mark discovery.id as `external`
		r.publishDatabaseEvent(ctx, "external", database)
	}

	r.setting.Logger.Info("Completed domain-based database discovery",
		zap.Int("domain_rules_count", len(r.config.Database.DomainRules)),
		zap.Int("external_name_services", externalNameServices),
		zap.Int("matched_services", matchedServices))
}
