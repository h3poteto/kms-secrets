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
	v1 "k8s.io/api/core/v1"
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
// +kubebuilder:rbac:groups="",resources=secret,verbs=get;list;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=events,verbs=create;patch

func (r *KMSSecretReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("kmssecret", req.NamespacedName)

	// An error has occur when KMSSecret is deleted.
	log.Info("fetching KMSSecret resources")
	kind := secretv1beta1.KMSSecret{}
	if err := r.Client.Get(ctx, req.NamespacedName, &kind); err != nil {
		log.Error(err, "failed to get KMSSecret resources")

		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	log = log.WithValues("secret_name", kind.Name)

	decryptedData, err := decryptData(kind.Spec.EncryptedData, kind.Spec.Region)
	if err != nil {
		log.Error(err, "failed to decrypt data")

		return ctrl.Result{}, err
	}

	shasum := shasumData(decryptedData)

	log.Info("checking if an existing Secret exists for this resource")
	secret := v1.Secret{}
	err = r.Client.Get(ctx, client.ObjectKey{Namespace: kind.Namespace, Name: kind.Name}, &secret)

	// Create a new Secret if there is no secret associated with KMSSecret.
	if apierrors.IsNotFound(err) {
		log.Info("could not find existing Secret for KMSSecret, creating one...")

		secret := buildSecret(kind, decryptedData)
		if err := r.Client.Create(ctx, secret); err != nil {
			log.Error(err, "failed to create Secret resource")
			return ctrl.Result{}, err
		}

		r.Recorder.Eventf(&kind, v1.EventTypeNormal, "Created", "Created Secret %q", secret.Name)
		log.Info("created Secret resource for KMSSecret")

		kind.Status.SecretsSum = shasum
		if err := r.Client.Update(ctx, &kind); err != nil {
			log.Error(err, "failed to update KMSSecret status")
			return ctrl.Result{}, err
		}
		log.Info("updated KMSSecret resource status")

		return ctrl.Result{}, nil
	}
	if err != nil {
		log.Error(err, "failed to get Secret for KMSSecret resource")
		return ctrl.Result{}, err
	}

	// Check status and update secret if there are differences.
	if kind.Status.SecretsSum != shasum {
		log.Info("encryptedData is updated, so updating secret resource", "old_secrets_sum", kind.Status.SecretsSum)
		secret := buildSecret(kind, decryptedData)
		if err := r.Client.Update(ctx, secret); err != nil {
			log.Error(err, "failed to update Secret resource")
			return ctrl.Result{}, err
		}
		r.Recorder.Eventf(&kind, v1.EventTypeNormal, "Updated", "Updated Secret %q", secret.Name)
		log.Info("updated Secret resource for KMSSecret")

		kind.Status.SecretsSum = shasum
		if err := r.Client.Update(ctx, &kind); err != nil {
			log.Error(err, "failed to update KMSSecret status")
			return ctrl.Result{}, err
		}
	}

	log.Info("resource status synced")

	return ctrl.Result{}, nil
}

func (r *KMSSecretReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&secretv1beta1.KMSSecret{}).
		Owns(&v1.Secret{}).
		Complete(r)
}

func buildSecret(kind secretv1beta1.KMSSecret, decryptedData map[string][]byte) *v1.Secret {
	secret := v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:            kind.Name,
			Namespace:       kind.Namespace,
			OwnerReferences: []metav1.OwnerReference{*metav1.NewControllerRef(&kind, secretv1beta1.GroupVersion.WithKind("KMSSecret"))},
		},
		Data: decryptedData,
		Type: v1.SecretTypeOpaque,
	}
	return &secret
}

func decryptData(encryptedData map[string][]byte, region string) (map[string][]byte, error) {
	svc := kms.New(session.New(), aws.NewConfig().WithRegion(region))
	decryptedData := make(map[string][]byte)
	for key, value := range encryptedData {
		input := &kms.DecryptInput{
			CiphertextBlob: value,
		}
		decrypted, err := svc.Decrypt(input)
		if err != nil {
			return nil, err
		}
		decryptedData[key] = decrypted.Plaintext
	}
	return decryptedData, nil
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
