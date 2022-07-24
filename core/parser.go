package core

import (
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

func checkStableAndGetLatestUpdatedTime(deploy *appsv1.Deployment) (time.Time, bool) {
	var available bool
	var availableLastUpdateTime time.Time
	var processing bool
	var processingLastUpdateTime time.Time
	for _, cond := range deploy.Status.Conditions {
		if cond.Type == appsv1.DeploymentAvailable && cond.Status == corev1.ConditionTrue {
			available = true
			availableLastUpdateTime = cond.LastUpdateTime.Time
		}
		if cond.Type == appsv1.DeploymentProgressing && cond.Status == corev1.ConditionTrue && cond.Reason == _NewRSAvailableReason {
			processing = true
			processingLastUpdateTime = cond.LastUpdateTime.Time
		}
	}

	if available && processing {
		if availableLastUpdateTime.After(processingLastUpdateTime) {
			return availableLastUpdateTime, true
		} else {
			return processingLastUpdateTime, true
		}
	}
	return time.Time{}, false
}

func listReferencedConfigurations(deploy *appsv1.Deployment) (map[string]empty, map[string]empty) {
	configMapNames := make(map[string]empty, 0)
	secretNames := make(map[string]empty, 0)

	for _, container := range deploy.Spec.Template.Spec.Containers {
		for _, env := range container.Env {
			if env.ValueFrom == nil {
				continue
			}

			if env.ValueFrom.ConfigMapKeyRef != nil {
				configMapNames[env.ValueFrom.ConfigMapKeyRef.Name] = empty{}
			}
			if env.ValueFrom.SecretKeyRef != nil {
				secretNames[env.ValueFrom.SecretKeyRef.Name] = empty{}
			}

			for _, envFrom := range container.EnvFrom {
				if envFrom.ConfigMapRef != nil {
					configMapNames[envFrom.ConfigMapRef.Name] = empty{}
				}
				if envFrom.SecretRef != nil {
					secretNames[envFrom.SecretRef.Name] = empty{}
				}
			}
		}
	}
	return configMapNames, secretNames
}
