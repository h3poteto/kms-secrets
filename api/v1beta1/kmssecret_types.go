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

package v1beta1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// SecretTemplateSpec defines the secret metadata
type SecretTemplateSpec struct {
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
}

// KMSSecretSpec defines the desired state of KMSSecret
type KMSSecretSpec struct {
	// +optional
	Template SecretTemplateSpec `json:"template"`

	// +kubebuilder:validation:Required
	EncryptedData map[string][]byte `json:"encryptedData"`
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Type:=string
	Region string `json:"region"`
}

// KMSSecretStatus defines the observed state of KMSSecret
type KMSSecretStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	SecretsSum string `json:"secretsSum,omitempty"`
}

// +kubebuilder:object:root=true

// KMSSecret is the Schema for the kmssecrets API
type KMSSecret struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   KMSSecretSpec   `json:"spec,omitempty"`
	Status KMSSecretStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// KMSSecretList contains a list of KMSSecret
type KMSSecretList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []KMSSecret `json:"items"`
}

func init() {
	SchemeBuilder.Register(&KMSSecret{}, &KMSSecretList{})
}
