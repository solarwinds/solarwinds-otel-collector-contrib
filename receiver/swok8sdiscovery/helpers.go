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
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// matchServiceForPod attempts to find a service in namespace selecting the pod.
// If multiple services match, prefer one exposing any of the provided ports.
func matchServiceForPod(pod corev1.Pod, container corev1.Container, ports []int32, services []corev1.Service) (string, []int32) {
	if len(services) == 0 {
		return "", nil
	}
	podLabels := pod.GetLabels()
	type svcMatch struct {
		svc         corev1.Service
		svcPorts    []int32
		targetPorts []int32
		portMatches int
	}
	var candidates []svcMatch
	nameToPort := map[string]int32{}
	for _, p := range container.Ports {
		if p.Name != "" {
			nameToPort[p.Name] = p.ContainerPort
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
			portMatches := 0
			for _, sp := range svc.Spec.Ports {
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

				if _, ok := portSet[tp]; ok {
					targetPorts = append(targetPorts, tp)
					svcPorts = append(svcPorts, sp.Port)
					portMatches++
				}
			}
			if portMatches == len(ports) {
				// we have exact match, let's return it immediately
				return svc.Name, svcPorts
			} else if portMatches != 0 {
				candidates = append(candidates, svcMatch{svc: svc, svcPorts: svcPorts, targetPorts: targetPorts, portMatches: portMatches})
			}
		}
	}
	if len(candidates) == 0 {
		return "", nil
	}

	var best svcMatch
	bestMatches := 0
	// in case we have multiple candidates, pick one with best port match and
	// prefer one with alphabetically order to be deterministic with each run
	for i, c := range candidates {
		cMatches := c.portMatches
		if i == 0 || cMatches < bestMatches || (cMatches == bestMatches && c.svc.Name < best.svc.Name) {
			best = c
			bestMatches = cMatches
		}
	}

	return best.svc.Name, best.svcPorts
}

func selectorMatches(selector, labels map[string]string) bool {
	for k, v := range selector {
		if labels[k] != v {
			return false
		}
	}
	return true
}
