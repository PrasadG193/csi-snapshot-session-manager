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

type SessionStateType string

const (
	SessionStateTypeReady   = "Ready"
	SessionStateTypePending = "Pending"
	SessionStateTypeFailed  = "Failed"
)

// CSISnapshotSessionAccessSpec defines the desired state of CSISnapshotSessionAccess
type CSISnapshotSessionAccessSpec struct {
	// The list of snapshots to generate session for
	Snapshots []string `json:"snapshots,omitempty"`
}

// CSISnapshotSessionAccessStatus defines the observed state of CSISnapshotSessionAccess
type CSISnapshotSessionAccessStatus struct {
	// State of the CSISnapshotSessionAccess. One of the "Ready", "Pending", "Failed"
	SessionState SessionStateType `json:"sessionState"`

	// Captures any error encountered.
	Error string `json:"error,omitempty"`

	// CABundle client side CA used for server validation
	CACert []byte `json:"caCert,omitempty"`

	// SessionToken cbt server token for validation
	SessionToken []byte `json:"sessionToken,omitempty"`

	// SessionURL to get CBT metadata from
	SessionURL string `json:"sessionURL,omitempty"`

	// ExpiryTime
	ExpiryTime *metav1.Time `json:"expiryTime,omitempty"`
}

//+kubebuilder:object:root=true

// CSISnapshotSessionAccess is the Schema for the csisnapshotsessionaccesses API
type CSISnapshotSessionAccess struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CSISnapshotSessionAccessSpec   `json:"spec,omitempty"`
	Status CSISnapshotSessionAccessStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// CSISnapshotSessionAccessList contains a list of CSISnapshotSessionAccess
type CSISnapshotSessionAccessList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CSISnapshotSessionAccess `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CSISnapshotSessionAccess{}, &CSISnapshotSessionAccessList{})
}
