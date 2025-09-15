package swok8sdiscovery

import (
	"context"
	"strconv"
	"strings"

	corev1 "k8s.io/api/core/v1"
)

// discoverDatabases matches containers to database image rules, resolves ports and associated services.
func (r *swok8sdiscoveryReceiver) discoverDatabasesByImages(ctx context.Context, pods []corev1.Pod, services []corev1.Service) {
	if len(r.config.ImageRules) == 0 {
		return
	}

	// Reduce scope of services to those in each namespace for faster lookup.
	svcByNamespace := map[string][]corev1.Service{}
	for _, svc := range services {
		svcByNamespace[svc.Namespace] = append(svcByNamespace[svc.Namespace], svc)
	}

	// Track emitted workload+db+port combinations to avoid duplicate events while allowing distinct ports.
	// Key format: namespace|workloadKind|workloadName|dbType|ports(sorted,comma-separated)
	emittedWorkloads := make(map[string]struct{})
	for _, pod := range pods {
		for _, container := range pod.Spec.Containers {
			matchedRule := (*ImageRule)(nil)
			for i := range r.config.ImageRules {
				rule := &r.config.ImageRules[i]
				for _, rx := range rule.PatternsCompiled {
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
				continue
			}

			// Resolve ports
			ports := resolveContainerPorts(container, matchedRule.DefaultPort)

			// Try to match service exposing one of these ports by label selector (pod labels subset service selector)
			svcName, svcPorts, svcTargetPorts := matchServiceForPod(pod, ports, svcByNamespace[pod.Namespace])

			// Resolve workload (top-level)
			wKind, wName, _ := r.resolveWorkloadForPod(ctx, &pod)

			// duplicate workload+db+port combo
			if wKind != "" && wName != "" {
				// Build stable port key (ports slice already reflects selected port(s))
				portKey := strings.Join(portsAsStrings(ports), ",")
				key := pod.Namespace + "|" + wKind + "|" + wName + "|" + matchedRule.DatabaseType + "|" + portKey
				if _, exists := emittedWorkloads[key]; exists {
					continue
				}
				emittedWorkloads[key] = struct{}{}
			}

			r.publishDatabaseEvent(ctx, databaseEvent{
				DatabaseType:       matchedRule.DatabaseType,
				Namespace:          pod.Namespace,
				ServiceName:        svcName,
				Endpoint:           firstNonEmpty(svcName, pod.Name),
				Ports:              portsAsStrings(ports),
				ServicePorts:       portsAsStrings(svcPorts),
				ServiceTargetPorts: portsAsStrings(svcTargetPorts),
				WorkloadKind:       wKind,
				WorkloadName:       wName,
			})
		}
	}
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

// firstNonEmpty returns the first non-empty string from the arguments.
func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}

func portsAsStrings(ports []int32) []string {
	if len(ports) == 0 {
		return nil
	}
	res := make([]string, len(ports))
	for i, p := range ports {
		res[i] = strconv.FormatInt(int64(p), 10)
	}
	return res
}
