package e2e_test

import (
	"context"
	"fmt"
	"os"
	"time"

	secretv1beta1 "github.com/h3poteto/kms-secrets/api/v1beta1"
	"github.com/h3poteto/kms-secrets/e2e/fixtures"
	"github.com/h3poteto/kms-secrets/e2e/util"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("E2E", func() {
	var (
		cfg *rest.Config
	)
	BeforeSuite(func() {
		// Deploy operator controller
		configfile := os.Getenv("KUBECONFIG")
		if configfile == "" {
			configfile = "$HOME/.kube/config"
		}
		restConfig, err := clientcmd.BuildConfigFromFlags("", os.ExpandEnv(configfile))
		if err != nil {
			panic(err)
		}
		cfg = restConfig

		client, err := kubernetes.NewForConfig(restConfig)
		if err != nil {
			panic(err)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()

		if err := applyCRD(ctx, restConfig); err != nil {
			panic(err)
		}
		klog.Info("applying RBAC")
		if err := util.ApplyRBAC(ctx, restConfig); err != nil {
			panic(err)
		}

		// Apply manager
		klog.Info("applying manager")
		if err := applyManager(ctx, client, "default"); err != nil {
			panic(err)
		}
	})
	AfterSuite(func() {
		// Delete operator controller and custom resources
		configfile := os.Getenv("KUBECONFIG")
		if configfile == "" {
			configfile = "$HOME/.kube/config"
		}
		restConfig, err := clientcmd.BuildConfigFromFlags("", os.ExpandEnv(configfile))
		if err != nil {
			panic(err)
		}
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()

		client, err := kubernetes.NewForConfig(restConfig)
		if err != nil {
			panic(err)
		}

		if err := deleteManager(ctx, client, "default"); err != nil {
			klog.Error(err)
		}
		if err := util.DeleteRBAC(ctx, restConfig); err != nil {
			klog.Error(err)
		}
		if err := util.DeleteCRD(ctx, restConfig); err != nil {
			klog.Error(err)
		}
	})
	Describe("Secrets are created", func() {
		var (
			k8sClient  client.Client
			setupError error
			region     string = "ap-northeast-1"
			ns         string = "default"
			kmsSecret  *secretv1beta1.KMSSecret
		)
		JustBeforeEach(func() {
			ctx := context.Background()
			var err error
			err = secretv1beta1.AddToScheme(scheme.Scheme)
			if err != nil {
				klog.Error(err)
				setupError = err
				return
			}
			k8sClient, err = client.New(cfg, client.Options{Scheme: scheme.Scheme})
			if err != nil {
				klog.Error(err)
				setupError = err
				return
			}
			klog.Infof("creating: %s/%s", kmsSecret.Namespace, kmsSecret.Name)
			err = k8sClient.Create(ctx, kmsSecret)
			if err != nil {
				klog.Error(err)
				setupError = err
				return
			}
			setupError = wait.Poll(1*time.Second, 5*time.Minute, func() (bool, error) {
				res := secretv1beta1.KMSSecret{}
				err := k8sClient.Get(ctx, types.NamespacedName{
					Namespace: kmsSecret.Namespace,
					Name:      kmsSecret.Name,
				}, &res)
				if err != nil {
					if apierrors.IsNotFound(err) {
						return false, nil
					}
					klog.Error(err)
					return false, err
				}
				kmsSecret = &res
				return true, nil
			})
		})
		AfterEach(func() {
			ctx := context.Background()
			err := k8sClient.Delete(ctx, kmsSecret)
			if err != nil {
				panic(err)
			}
		})
		Context("Encrypted data using aws cli", func() {
			decrypt := func(key, value, expected string) {
				BeforeEach(func() {
					keyID := os.Getenv("KMS_KEY_ID")
					if keyID == "" {
						panic(fmt.Errorf("KMS_KEY_ID is required"))
					}
					data, err := util.EncryptString(value, keyID, os.Getenv("AWS_REGION"))
					if err != nil {
						panic(err)
					}
					kmsSecret = fixtures.NewKMSSecret(ns, "test-secret", region, map[string][]byte{
						key: data,
					})
				})
				It("Secret data should be decrepted", func() {
					ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
					defer cancel()
					Expect(setupError).To(BeNil())
					res := corev1.Secret{}
					err := wait.Poll(1*time.Second, 5*time.Minute, func() (bool, error) {
						err := k8sClient.Get(ctx, types.NamespacedName{Namespace: kmsSecret.Namespace, Name: kmsSecret.Name}, &res)
						if err != nil {
							if apierrors.IsNotFound(err) {
								return false, nil
							}
							klog.Error(err)
							return false, err
						}
						return true, nil
					})
					Expect(err).To(BeNil())
					val, ok := res.Data[key]
					Expect(ok).To(BeTrue())
					Expect(string(val)).To(Equal(expected))
				})
			}
			Context("Value is plain text", func() {
				decrypt("api_key", "hogehoge", "hogehoge")
			})
			Context("Value is yaml object", func() {
				decrypt("api_key", "hoge: fuga", "hoge: fuga")
			})
			Context("Value is yaml formatted text", func() {
				decrypt("api_key", "--- hogehoge", "hogehoge")
			})
		})
	})
})

func applyCRD(ctx context.Context, cfg *rest.Config) error {
	klog.Info("applying CRD")
	err := util.ApplyCRD(ctx, cfg)
	if err != nil {
		panic(err)
	}
	time.Sleep(10 * time.Second)
	return err
}

func applyManager(ctx context.Context, client *kubernetes.Clientset, ns string) error {
	image := os.Getenv("KMS_SECRETS_IMAGE")
	if image == "" {
		return fmt.Errorf("KMS_SECRETS_IMAGE is required")
	}
	region := os.Getenv("AWS_REGION")
	if region == "" {
		return fmt.Errorf("AWS_REGION is required")
	}
	accessKey := os.Getenv("AWS_ACCESS_KEY_ID")
	if accessKey == "" {
		return fmt.Errorf("AWS_ACCESS_KEY_ID is required")
	}
	secretKey := os.Getenv("AWS_SECRET_ACCESS_KEY")
	if secretKey == "" {
		return fmt.Errorf("AWS_SECRET_ACCESS_KEY is required")
	}
	sa, deployment := fixtures.NewManagerManifests(ns, "manager", image, region, accessKey, secretKey)
	if _, err := client.CoreV1().ServiceAccounts(ns).Create(ctx, sa, metav1.CreateOptions{}); err != nil {
		return err
	}
	if _, err := client.AppsV1().Deployments(ns).Create(ctx, deployment, metav1.CreateOptions{}); err != nil {
		return err
	}
	err := wait.Poll(10*time.Second, 5*time.Minute, func() (bool, error) {
		podList, err := client.CoreV1().Pods(ns).List(ctx, metav1.ListOptions{
			LabelSelector: fmt.Sprintf("%s=%s", fixtures.ManagerPodLabelKey, fixtures.ManagerPodLabelValue),
		})
		if err != nil {
			if kerrors.IsNotFound(err) {
				return false, nil
			}
			return false, err
		}

		return util.WaitPodRunning(podList)
	})
	if err != nil {
		return err
	}
	return nil
}

func deleteManager(ctx context.Context, client *kubernetes.Clientset, ns string) error {
	image := os.Getenv("KMS_SECRETS_IMAGE")
	if image == "" {
		return fmt.Errorf("KMS_SECRETS_IMAGE is required")
	}
	region := os.Getenv("AWS_REGION")
	if region == "" {
		return fmt.Errorf("AWS_REGION is required")
	}
	accessKey := os.Getenv("AWS_ACCESS_KEY_ID")
	if accessKey == "" {
		return fmt.Errorf("AWS_ACCESS_KEY_ID is required")
	}
	secretKey := os.Getenv("AWS_SECRET_ACCESS_KEY")
	if secretKey == "" {
		return fmt.Errorf("AWS_SECRET_ACCESS_KEY is required")
	}
	sa, deployment := fixtures.NewManagerManifests(ns, "manager", image, region, accessKey, secretKey)
	if err := client.AppsV1().Deployments(ns).Delete(ctx, deployment.Name, metav1.DeleteOptions{}); err != nil {
		return err
	}
	if err := client.CoreV1().ServiceAccounts(ns).Delete(ctx, sa.Name, metav1.DeleteOptions{}); err != nil {
		return err
	}
	err := wait.Poll(10*time.Second, 10*time.Minute, func() (bool, error) {
		podList, err := client.CoreV1().Pods(ns).List(ctx, metav1.ListOptions{
			LabelSelector: fmt.Sprintf("%s=%s", fixtures.ManagerPodLabelKey, fixtures.ManagerPodLabelValue),
		})
		if err != nil {
			if kerrors.IsNotFound(err) {
				return true, nil
			}
			return false, err
		}
		klog.V(4).Infof("Pods are: %#v", podList.Items)
		if len(podList.Items) == 0 {
			return true, nil
		}
		return false, nil
	})
	if err != nil {
		return err
	}
	return nil
}
