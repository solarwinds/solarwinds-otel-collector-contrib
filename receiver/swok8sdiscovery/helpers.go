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
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// matchServiceForPod attempts to find a service in namespace selecting the pod.
// If multiple services match, prefer one exposing any of the provided ports.
func matchServiceForPod(pod corev1.Pod, ports []int32, services []corev1.Service) (string, []int32, []int32) {
	if len(services) == 0 {
		return "", nil, nil
	}
	podLabels := pod.GetLabels()
	type svcMatch struct {
		svc         corev1.Service
		svcPorts    []int32
		targetPorts []int32
		overlaps    bool
	}
	var candidates []svcMatch
	nameToPort := map[string]int32{}
	for _, c := range pod.Spec.Containers {
		for _, p := range c.Ports {
			if p.Name != "" {
				nameToPort[p.Name] = p.ContainerPort
			}
		}
	}
	portSet := map[int32]struct{}{}
	for _, p := range ports {
		portSet[p] = struct{}{}
	}

	for _, svc := range services {
		selector := svc.Spec.Selector
		if len(selector) == 0 {
			continue
		}
		if selectorMatches(selector, podLabels) {
			var svcPorts []int32
			var targetPorts []int32
			overlap := false
			for _, sp := range svc.Spec.Ports {
				svcPorts = append(svcPorts, sp.Port)
				var tp int32
				if sp.TargetPort.Type == intstr.Int {
					tp = int32(sp.TargetPort.IntValue())
				} else if sp.TargetPort.Type == intstr.String && sp.TargetPort.StrVal != "" {
					if v, ok := nameToPort[sp.TargetPort.StrVal]; ok {
						tp = v
					}
				}
				if tp == 0 {
					tp = sp.Port
				}
				targetPorts = append(targetPorts, tp)
				if _, ok := portSet[tp]; ok {
					overlap = true
				}
			}
			candidates = append(candidates, svcMatch{svc: svc, svcPorts: svcPorts, targetPorts: targetPorts, overlaps: overlap})
		}
	}
	if len(candidates) == 0 {
		return "", nil, nil
	}
	var overlapping []svcMatch
	for _, c := range candidates {
		if c.overlaps {
			overlapping = append(overlapping, c)
		}
	}
	if len(overlapping) == 0 {
		return "", nil, nil
	}
	var best svcMatch
	bestExtra := 0
	for i, c := range overlapping {
		cExtra := countExtra(c.targetPorts, portSet)
		if i == 0 || cExtra < bestExtra || (cExtra == bestExtra && c.svc.Name < best.svc.Name) {
			best = c
			bestExtra = cExtra
		}
	}
	return best.svc.Name, best.svcPorts, best.targetPorts
}

func selectorMatches(selector, labels map[string]string) bool {
	for k, v := range selector {
		if labels[k] != v {
			return false
		}
	}
	return true
}

// countExtra counts how many ports in slice are not in the discovered set
func countExtra(ports []int32, set map[int32]struct{}) int {
	extra := 0
	for _, p := range ports {
		if _, ok := set[p]; !ok {
			extra++
		}
	}
	return extra
}
