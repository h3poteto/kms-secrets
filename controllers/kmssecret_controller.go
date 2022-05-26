/*


Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"crypto/sha256"
	"fmt"
	"sort"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/go-logr/logr"
	"github.com/h3poteto/controller-klog/pkg/ctrklog"
	"gopkg.in/yaml.v3"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	secretv1beta1 "github.com/h3poteto/kms-secrets/api/v1beta1"
)

// KMSSecretReconciler reconciles a KMSSecret object
type KMSSecretReconciler struct {
	client.Client
	Log      logr.Logger
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

// +kubebuilder:rbac:groups=secret.h3poteto.dev,resources=kmssecrets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=secret.h3poteto.dev,resources=kmssecrets/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=events,verbs=create;patch

func (r *KMSSecretReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	ctx = ctrklog.SetController(ctx, "kmssecret")
	// An error has occur when KMSSecret is deleted.
	ctrklog.Info(ctx, "fetching KMSSecret resources")

	kind := secretv1beta1.KMSSecret{}
	if err := r.Client.Get(ctx, req.NamespacedName, &kind); err != nil {
		ctrklog.Errorf(ctx, "failed to get KMSSecret: %v", err)
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	ctx = ctrklog.SetObject(ctx, kind.Name)

	decryptedData, err := decryptData(ctx, kind.Spec.EncryptedData, kind.Spec.Region)
	if err != nil {
		ctrklog.Errorf(ctx, "failed to decrypt data: %v", err)

		return ctrl.Result{}, err
	}

	shasum := shasumData(decryptedData)

	ctrklog.Info(ctx, "checking if an existing Secret for this resource")
	secret := corev1.Secret{}
	err = r.Client.Get(ctx, client.ObjectKey{Namespace: kind.Namespace, Name: kind.Name}, &secret)

	// Create a new Secret if there is no secret associated with KMSSecret.
	if apierrors.IsNotFound(err) {
		ctrklog.Info(ctx, "could not find existing Secret for KMSSecret, creating one...")

		secret := buildSecret(kind, decryptedData)
		if err := r.Client.Create(ctx, secret); err != nil {
			ctrklog.Errorf(ctx, "failed to create Secret %s/%s: %v", secret.Namespace, secret.Name, err)
			return ctrl.Result{}, err
		}

		r.Recorder.Eventf(&kind, corev1.EventTypeNormal, "Created", "Created Secret %s/%s", secret.Namespace, secret.Name)
		ctrklog.Infof(ctx, "created Secret %s/%s", secret.Namespace, secret.Name)

		kind.Status.SecretsSum = shasum
		if err := r.Client.Update(ctx, &kind); err != nil {
			ctrklog.Errorf(ctx, "failed to update KMSSecret %s/%s: %v", kind.Namespace, kind.Name, err)
			return ctrl.Result{}, err
		}
		ctrklog.Infof(ctx, "updated KMSSecret resource status %s/%s", kind.Namespace, kind.Name)

		return ctrl.Result{}, nil
	}
	if err != nil {
		ctrklog.Errorf(ctx, "failed to get Secret for KMSSecret %s/%s: %v", kind.Namespace, kind.Name, err)
		return ctrl.Result{}, err
	}

	// Check status and update secret if there are differences.
	if kind.Status.SecretsSum != shasum {
		ctrklog.Infof(ctx, "encryptedData is updated, so updating secret resource", "old_secrets_sum", kind.Status.SecretsSum)
		secret := buildSecret(kind, decryptedData)
		if err := r.Client.Update(ctx, secret); err != nil {
			ctrklog.Errorf(ctx, "failed to update Secret %s/%s: %v", secret.Namespace, secret.Name, err)
			return ctrl.Result{}, err
		}
		r.Recorder.Eventf(&kind, corev1.EventTypeNormal, "Updated", "Updated Secret %s/%s", secret.Namespace, secret.Name)
		ctrklog.Info(ctx, "updated Secret %s/%s", secret.Namespace, secret.Name)

		kind.Status.SecretsSum = shasum
		if err := r.Client.Update(ctx, &kind); err != nil {
			ctrklog.Errorf(ctx, "failed to update KMSSecret %s/%s: %v", kind.Namespace, kind.Name, err)
			return ctrl.Result{}, err
		}
	}

	ctrklog.Info(ctx, "resource status synced")

	return ctrl.Result{}, nil
}

func (r *KMSSecretReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&secretv1beta1.KMSSecret{}).
		Owns(&corev1.Secret{}).
		Complete(r)
}

func buildSecret(kind secretv1beta1.KMSSecret, decryptedData map[string][]byte) *corev1.Secret {
	secret := corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:            kind.Name,
			Namespace:       kind.Namespace,
			OwnerReferences: []metav1.OwnerReference{*metav1.NewControllerRef(&kind, secretv1beta1.GroupVersion.WithKind("KMSSecret"))},
			Labels:          kind.Spec.Template.GetLabels(),
			Annotations:     kind.Spec.Template.GetAnnotations(),
		},
		Data: decryptedData,
		Type: corev1.SecretTypeOpaque,
	}
	return &secret
}

// decryptData decrypt data using AWS KMS.
func decryptData(ctx context.Context, encryptedData map[string][]byte, region string) (map[string][]byte, error) {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	svc := kms.New(sess, aws.NewConfig().WithRegion(region))
	decryptedData := make(map[string][]byte)
	for key, value := range encryptedData {
		input := &kms.DecryptInput{
			CiphertextBlob: value,
		}
		decrypted, err := svc.Decrypt(input)
		if err != nil {
			ctrklog.Errorf(ctx, "failed to decrypt: %v", err)
			return nil, err
		}
		plain := decrypted.Plaintext
		value, err = yamlParse(plain)
		if err != nil {
			ctrklog.Warningf(ctx, "failed to yaml parse for %s, so insert plain text", key)
			decryptedData[key] = plain
			continue
		}
		decryptedData[key] = value

	}
	return decryptedData, nil
}

func yamlParse(input []byte) ([]byte, error) {
	var res string
	if err := yaml.Unmarshal(input, &res); err != nil {
		return nil, fmt.Errorf("failed to unmarshal: %w", err)
	}
	return []byte(res), nil
}

func shasumData(data map[string][]byte) string {
	keys := make([]string, 0, len(data))
	for k := range data {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	values := make([]string, 0, len(data))
	for _, k := range keys {
		values = append(values, string(data[k]))
	}
	raw := strings.Join(keys, ",") + ":" + strings.Join(values, ",")
	sum := sha256.Sum256([]byte(raw))
	return fmt.Sprintf("%x", sum)

}
