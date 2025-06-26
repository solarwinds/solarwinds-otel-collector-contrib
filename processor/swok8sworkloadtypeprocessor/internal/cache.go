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

package internal // import "github.com/solarwinds/solarwinds-otel-collector-contrib/processor/swok8sworkloadtypeprocessor/internal"

import (
	"fmt"

	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/cache"
)

func PodTransformFunc(logger *zap.Logger, keepOwnerReferences bool) cache.TransformFunc {
	return func(obj any) (any, error) {
		pod, ok := obj.(*corev1.Pod)
		if !ok {
			logger.Error("Received an unexpected workload object type for podTransform", zap.String("workloadObjectType", fmt.Sprintf("%T", obj)))
			return obj, nil
		}
		logger.Debug("Received pod", zap.String("name", pod.Name), zap.String("namespace", pod.Namespace))

		var ownerRefs []metav1.OwnerReference
		if keepOwnerReferences {
			ownerRefs = copyOwnerReferences(pod)
		}

		return &corev1.Pod{
			TypeMeta: metav1.TypeMeta{
				Kind: PodKind,
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:            pod.Name,
				Namespace:       pod.Namespace,
				OwnerReferences: ownerRefs,
			},
			Spec: corev1.PodSpec{
				HostNetwork: pod.Spec.HostNetwork,
			},
			Status: corev1.PodStatus{
				PodIPs: pod.Status.PodIPs,
			},
		}, nil
	}
}

func ServiceTransformFunc(logger *zap.Logger) cache.TransformFunc {
	return func(obj any) (any, error) {
		service, ok := obj.(*corev1.Service)
		if !ok {
			logger.Error("Received an unexpected workload object type for serviceTransform", zap.String("workloadObjectType", fmt.Sprintf("%T", obj)))
			return obj, nil
		}
		logger.Debug("Received service", zap.String("name", service.Name), zap.String("namespace", service.Namespace))
		return &corev1.Service{
			TypeMeta: metav1.TypeMeta{
				Kind: ServiceKind,
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      service.Name,
				Namespace: service.Namespace,
			},
			Spec: corev1.ServiceSpec{
				ClusterIPs:  service.Spec.ClusterIPs,
				ExternalIPs: service.Spec.ExternalIPs,
			},
		}, nil
	}
}

func GenericTransformFunc(logger *zap.Logger, kind string, keepOwnerReferences bool) cache.TransformFunc {
	return func(obj any) (any, error) {
		workload, ok := obj.(metav1.Object)
		if !ok {
			logger.Error("Received an unexpected workload object type", zap.String("workloadObjectType", fmt.Sprintf("%T", obj)))
			return obj, nil
		}
		logger.Debug("Received workload", zap.String("workloadName", workload.GetName()), zap.String("workloadNamespace", workload.GetNamespace()))

		var ownerRefs []metav1.OwnerReference
		if keepOwnerReferences {
			ownerRefs = copyOwnerReferences(workload)
		}

		return &metav1.PartialObjectMetadata{
			TypeMeta: metav1.TypeMeta{
				Kind: kind,
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:            workload.GetName(),
				Namespace:       workload.GetNamespace(),
				OwnerReferences: ownerRefs,
			},
		}, nil
	}
}

func copyOwnerReferences(workload metav1.Object) []metav1.OwnerReference {
	origOwnerRefs := workload.GetOwnerReferences()
	copiedRefs := make([]metav1.OwnerReference, len(origOwnerRefs))
	for i, ref := range origOwnerRefs {
		copiedRefs[i] = metav1.OwnerReference{
			Kind: ref.Kind,
			Name: ref.Name,
		}
	}
	return copiedRefs
}

const (
	PodIpIndex     = "podIpIndex"
	ServiceIpIndex = "serviceIpIndex"
)

func PodIpIndexer(logger *zap.Logger) cache.Indexers {
	return cache.Indexers{
		PodIpIndex: func(obj any) ([]string, error) {
			pod, ok := obj.(*corev1.Pod)
			if !ok {
				logger.Error("Received an unexpected workload object type for podIpIndex", zap.String("workloadObjectType", fmt.Sprintf("%T", obj)))
				return nil, nil
			}

			var ips []string
			if !pod.Spec.HostNetwork {
				for _, ip := range pod.Status.PodIPs {
					ips = append(ips, ip.IP)
				}
			}
			return ips, nil
		},
	}
}

func ServiceIpIndexer(logger *zap.Logger) cache.Indexers {
	return cache.Indexers{
		ServiceIpIndex: func(obj any) ([]string, error) {
			service, ok := obj.(*corev1.Service)
			if !ok {
				logger.Error("Received an unexpected workload object type for serviceIpIndex", zap.String("workloadObjectType", fmt.Sprintf("%T", obj)))
				return nil, nil
			}

			var ips []string
			for _, ip := range service.Spec.ClusterIPs {
				if ip != "" && ip != "None" {
					ips = append(ips, ip)
				}
			}
			for _, ip := range service.Spec.ExternalIPs {
				if ip != "" && ip != "None" {
					ips = append(ips, ip)
				}
			}
			return ips, nil
		},
	}
}
