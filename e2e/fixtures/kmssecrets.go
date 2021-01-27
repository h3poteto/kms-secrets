package fixtures

import (
	secretv1beta1 "github.com/h3poteto/kms-secrets/api/v1beta1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	SecretLabelKey   = "secret.h3poteto.dev"
	SecretLabelValue = "secret"
)

var SecretLabels = map[string]string{
	SecretLabelKey: SecretLabelValue,
}

func NewKMSSecret(ns, name, region string, data map[string][]byte) *secretv1beta1.KMSSecret {
	return kmssecret(ns, name, region, data)
}

func kmssecret(ns, name, region string, data map[string][]byte) *secretv1beta1.KMSSecret {
	return &secretv1beta1.KMSSecret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "KMSSecret",
			APIVersion: "secret.h3poteto.dev/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ns,
		},
		Spec: secretv1beta1.KMSSecretSpec{
			Template: secretv1beta1.SecretTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: SecretLabels,
				},
			},
			EncryptedData: data,
			Region:        region,
		},
	}
}
