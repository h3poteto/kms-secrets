package controllers

import (
	"context"
	"sync"
	"time"

	secretv1beta1 "github.com/h3poteto/kms-secrets/api/v1beta1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

var _ = Describe("shasumData", func() {
	It("should be matched", func() {
		data := map[string][]byte{
			"API_KEY":  []byte("hoge"),
			"PASSWORD": []byte("fuga"),
		}
		sum := shasumData(data)
		Expect(sum).To(Equal("b6b66b55b6b03c6ee6abc0027095d38a35937eb3e6ff2dc9f2aafa846c704e3b"))
	})
})

var _ = Describe("yamlParse", func() {
	It("yaml string should be parsed", func() {
		input := []byte("--- apikey")
		result, err := yamlParse(input)
		Expect(err).Should(BeNil())
		Expect(string(result)).To(Equal("apikey"))
	})
	It("string should not be changed", func() {
		input := []byte("apikey")
		result, err := yamlParse(input)
		Expect(err).Should(BeNil())
		Expect(string(result)).To(Equal("apikey"))
	})
})

var _ = Describe("KMSSecretReconciler", func() {
	// var c client.Client
	var stopMgr chan struct{}
	var mgrStopped *sync.WaitGroup

	const timeout = time.Second * 5

	It("should setup Manager&Reconciler", func() {
		mgr, err := manager.New(cfg, manager.Options{})
		Expect(err).NotTo(HaveOccurred())
		//c = mgr.GetClient()

		rc := &KMSSecretReconciler{
			Client: mgr.GetClient(),
			Log:    ctrl.Log.WithName("controllers").WithName("Captain"),
			Scheme: mgr.GetScheme(),
		}

		err = rc.SetupWithManager(mgr)
		Expect(err).NotTo(HaveOccurred())

		stopMgr, mgrStopped = StartTestManager(mgr)
	})

	It("should stop Manager", func() {
		close(stopMgr)
		mgrStopped.Wait()
	})
})

var _ = Describe("Resource is created", func() {
	// This is a dummy value.
	encryptedPassword := "AQECAHhRyjfcAU0BFeAevHu12PEb8p++xXO5Ct1XvR/ARXpyaQAAAGYwZAYJKoZIhvcNAQcGoFcwVQIBADBQBgkqhkiG9w0BBwEwHgYJYIZIAWUDBAEuMBEEDIG5VsUXqw3PTaFgcAIBEIAjEciyjvz5Rp+MAl0yH76o1EDP5N2qzmH8WkwgR3zRjyA8mcc="

	It("should decrypt encrypted secret", func() {
		_, err := manager.New(cfg, manager.Options{})
		Expect(err).NotTo(HaveOccurred())

		kmsSecret := &secretv1beta1.KMSSecret{
			Spec: secretv1beta1.KMSSecretSpec{
				Template: secretv1beta1.SecretTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-secret",
						Namespace: "default",
					},
				},
				EncryptedData: map[string][]byte{
					"password": []byte(encryptedPassword),
				},
				Region: "ap-northeast-1",
			},
		}
		kmsSecret.Name = "test-kms-secret"
		kmsSecret.Namespace = "default"
		ctx := context.Background()
		err = k8sClient.Create(ctx, kmsSecret)
		Expect(err).NotTo(HaveOccurred())

		kmsSecret = getKMSSecret(ctx, "default", "test-kms-secret")
		Expect(kmsSecret.Spec.EncryptedData).To(Equal(map[string][]byte{
			"password": []byte(encryptedPassword),
		}))
		Expect(kmsSecret.ObjectMeta.UID).ToNot(Equal(""))
	})
})

func getKMSSecret(ctx context.Context, namespace, name string) *secretv1beta1.KMSSecret {
	kmsSecret := &secretv1beta1.KMSSecret{}
	namespaced := types.NamespacedName{
		Namespace: namespace,
		Name:      name,
	}
	err := k8sClient.Get(ctx, namespaced, kmsSecret)
	Expect(err).NotTo(HaveOccurred())
	return kmsSecret
}
