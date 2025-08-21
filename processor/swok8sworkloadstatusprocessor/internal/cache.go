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

package internal // import "github.com/solarwinds/solarwinds-otel-collector-contrib/processor/swok8sworkloadstatusprocessor/internal"

import (
	"fmt"

	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/cache"
)

func PodTransformFunc(logger *zap.Logger) cache.TransformFunc {
	return func(obj any) (any, error) {
		pod, ok := obj.(*corev1.Pod)
		if !ok {
			logger.Error("Received an unexpected workload object type for podTransform", zap.String("workloadObjectType", fmt.Sprintf("%T", obj)))
			return obj, nil
		}
		logger.Debug("Received pod", zap.String("name", pod.Name), zap.String("namespace", pod.Namespace))

		ownerRefs := copyOwnerReferences(pod)

		return &corev1.Pod{
			TypeMeta: metav1.TypeMeta{
				Kind: PodKind,
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:            pod.Name,
				Namespace:       pod.Namespace,
				OwnerReferences: ownerRefs,
			},
			Status: corev1.PodStatus{
				Phase: pod.Status.Phase,
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
