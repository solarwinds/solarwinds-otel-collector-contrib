package swok8sdiscovery

import (
	"context"
	"strings"

	corev1 "k8s.io/api/core/v1"
)

// discoverDatabasesByDomains matches ExternalName services to domain rules.
func (r *swok8sdiscoveryReceiver) discoverDatabasesByDomains(ctx context.Context, pods []corev1.Pod, services []corev1.Service) {
	if len(r.config.DomainRules) == 0 {
		return
	}
	for _, svc := range services {
		if svc.Spec.ExternalName == "" { // Only process ExternalName services for domain-based detection
			continue
		}
		external := svc.Spec.ExternalName

		matchingRules := make([]*DomainRule, 0)

		var matchedRule *DomainRule
		for i := range r.config.DomainRules {
			rule := &r.config.DomainRules[i]
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
		var svcTargetPorts []string
		var svcPorts []string
		for _, p := range svc.Spec.Ports {
			svcPorts = append(svcPorts, string(p.Port))
			svcTargetPorts = append(svcTargetPorts, p.TargetPort.StrVal)
		}

		r.publishDatabaseEvent(ctx, databaseEvent{
			DatabaseType:       matchedRule.DatabaseType,
			Namespace:          svc.Namespace,
			ServiceName:        svc.Name,
			ServicePorts:       svcPorts,
			ServiceTargetPorts: svcTargetPorts,
			Endpoint:           svc.Name,
			WorkloadKind:       wKind,
			WorkloadName:       wName,
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
