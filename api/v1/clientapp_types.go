/*
Copyright 2024.

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

package v1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ClientAppSpec defines the desired state of ClientApp
type ClientAppSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of ClientApp. Edit clientapp_types.go to remove/update
	//Foo string `json:"foo,omitempty"`
	Name      	string                      `json:"name,omitempty"`

	// +optional
	// +kubebuilder:default:="nginx:1.25"
	Image     	string                      `json:"image,omitempty"`

	// +optional
	// +kubebuilder:default:=1
	Replicas  	int32                       `json:"replicas,omitempty"`
	Env       	[]corev1.EnvVar             `json:"env,omitempty"`

	// +optional
	// +kubebuilder:default:=8080
	Port      	int32                       `json:"port,omitempty"`

	// +optional
	// +kubebuilder:default:="http"
	PortName  	string                      `json:"portname,omitempty"`
	Host      	string                      `json:"host,omitempty"`
	Resources 	corev1.ResourceRequirements `json:"resources,omitempty"`

	// +optional
	LastUpdated metav1.Time                 `json:"lastUpdated,omitempty"`
	LivenessProbe  *corev1.Probe            `json:"livenessProbe,omitempty"`
	ReadinessProbe *corev1.Probe            `json:"readinessProbe,omitempty"`
}

// ClientAppStatus defines the observed state of ClientApp
type ClientAppStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	URL        string             `json:"url,omitempty"`
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// ClientApp is the Schema for the clientapps API
type ClientApp struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ClientAppSpec   `json:"spec,omitempty"`
	Status ClientAppStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ClientAppList contains a list of ClientApp
type ClientAppList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ClientApp `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ClientApp{}, &ClientAppList{})
}
