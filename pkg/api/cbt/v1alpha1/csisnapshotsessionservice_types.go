/*
Copyright 2023.

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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CSISnapshotSessionServiceSpec defines the desired state of CSISnapshotSessionService
type CSISnapshotSessionServiceSpec struct {
	// CABundle client side CA used for server validation
	CACert []byte `json:"caCert,omitempty"`

	Address string `json:"address,omitempty"`
}

// CSISnapshotSessionServiceStatus defines the observed state of CSISnapshotSessionService
type CSISnapshotSessionServiceStatus struct {
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// CSISnapshotSessionService is the Schema for the csisnapshotsessionservices API
type CSISnapshotSessionService struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec CSISnapshotSessionServiceSpec `json:"spec,omitempty"`
}

//+kubebuilder:object:root=true

// CSISnapshotSessionServiceList contains a list of CSISnapshotSessionService
type CSISnapshotSessionServiceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CSISnapshotSessionService `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CSISnapshotSessionService{}, &CSISnapshotSessionServiceList{})
}
