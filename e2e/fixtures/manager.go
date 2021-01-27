package fixtures

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilpointer "k8s.io/utils/pointer"
)

const (
	// ManagerName is container name of manager.
	ManagerName = "manager"
	// ManagerPodLabelKey is label key for ManagerPodLabels.
	ManagerPodLabelKey = "operator.h3poteto.dev"
	// ManagerPodLabelValue is label value for ManagerPodLabels.
	ManagerPodLabelValue = "control-plane"
)

// ManagerPodLabels is label for generated manager Pods. This label is used in e2e tests when find generated Pods.
var ManagerPodLabels = map[string]string{
	ManagerPodLabelKey: ManagerPodLabelValue,
}

// NewManagerManifests generates ServiceAccount and Deployment manifests.
func NewManagerManifests(ns, sa, image, region, accessKey, secretKey string) (*corev1.ServiceAccount, *appsv1.Deployment) {
	return serviceAccount(ns, sa), deployment(ns, sa, image, region, accessKey, secretKey)
}

func serviceAccount(ns, name string) *corev1.ServiceAccount {
	return &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ns,
		},
	}
}

func deployment(ns, sa, image, region, accessKey, secretKey string) *appsv1.Deployment {
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "kms-secrets-manager",
			Namespace: ns,
			Labels:    ManagerPodLabels,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: utilpointer.Int32Ptr(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: ManagerPodLabels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: ManagerPodLabels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  ManagerName,
							Image: image,
							Command: []string{
								"/manager",
							},
							Args: []string{
								"--enable-leader-election",
							},
							EnvFrom: nil,
							Env: []corev1.EnvVar{
								{
									Name:  "AWS_REGION",
									Value: region,
								},
								{
									Name:  "AWS_ACCESS_KEY_ID",
									Value: accessKey,
								},
								{
									Name:  "AWS_SECRET_ACCESS_KEY",
									Value: secretKey,
								},
							},
						},
					},
					ServiceAccountName: sa,
				},
			},
		},
	}
}
