package swok8sdiscovery

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Canonical Kubernetes workload kind strings (local constants to avoid typos)
const (
	kindDeployment  = "Deployment"
	kindReplicaSet  = "ReplicaSet"
	kindStatefulSet = "StatefulSet"
	kindDaemonSet   = "DaemonSet"
	kindJob         = "Job"
	kindCronJob     = "CronJob"
)

// resolveWorkloadForPod attempts to identify the top-level workload controlling the pod.
// It follows these transitions:
// Pod -> ReplicaSet -> Deployment (for Deployments)
// Pod -> Job -> CronJob (one extra hop) if needed.
// For StatefulSet/DaemonSet it returns directly.
func (r *swok8sdiscoveryReceiver) resolveWorkloadForPod(ctx context.Context, pod *corev1.Pod) (string, string, bool) {
	if pod == nil {
		return "", "", false
	}
	owners := pod.GetOwnerReferences()
	if len(owners) == 0 {
		return "", "", false
	}
	var ctrl *metav1.OwnerReference
	for i := range owners {
		if owners[i].Controller != nil && *owners[i].Controller {
			ctrl = &owners[i]
			break
		}
	}
	if ctrl == nil {
		ctrl = &owners[0]
	}
	kind := ctrl.Kind
	name := ctrl.Name
	ns := pod.Namespace

	switch kind {
	case kindReplicaSet:
		rs, err := r.kclient.AppsV1().ReplicaSets(ns).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return kind, name, true
		}
		for i := range rs.OwnerReferences {
			or := rs.OwnerReferences[i]
			if or.Controller != nil && *or.Controller && or.Kind == kindDeployment {
				return kindDeployment, or.Name, true
			}
		}
		return kind, name, true
	case kindJob:
		job, err := r.kclient.BatchV1().Jobs(ns).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return kind, name, true
		}
		for i := range job.OwnerReferences {
			or := job.OwnerReferences[i]
			if or.Controller != nil && *or.Controller && or.Kind == kindCronJob {
				return kindCronJob, or.Name, true
			}
		}
		return kind, name, true
	case kindStatefulSet, kindDaemonSet, kindCronJob, kindDeployment:
		return kind, name, true
	default:
		return kind, name, true
	}
}
