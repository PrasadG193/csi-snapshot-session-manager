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

// VolumeSnapshotDeltaTokenSpec defines the desired state of VolumeSnapshotDeltaToken
type VolumeSnapshotDeltaTokenSpec struct {
	// The name of the base CSI volume snapshot to use for comparison.
	// If not specified, return all changed blocks.
	// +optional
	BaseVolumeSnapshotName string `json:"baseVolumeSnapshotName,omitempty"`

	// The name of the target CSI volume snapshot to use for comparison.
	// Required.
	TargetVolumeSnapshotName string `json:"targetVolumeSnapshotName"`
}

// VolumeSnapshotDeltaTokenStatus defines the observed state of VolumeSnapshotDeltaToken
type VolumeSnapshotDeltaTokenStatus struct {
	// State of the VolumeSnapshotDeltaToken. One of the "Ready", "Pending", "Failed"
	SessionState SessionStateType `json:"sessionState"`
	// Captures any error encountered.
	Error string `json:"error,omitempty"`

	// CABundle client side CA used for server validation
	CABundle []byte `json:"cabundle,omitempty"`

	// Token cbt server token for validation
	Token []byte `json:"token,omitempty"`

	// URL to get CBT metadata from
	URL string `json:"url,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// VolumeSnapshotDeltaToken is the Schema for the volumesnapshotdeltatokens API
type VolumeSnapshotDeltaToken struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   VolumeSnapshotDeltaTokenSpec   `json:"spec,omitempty"`
	Status VolumeSnapshotDeltaTokenStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// VolumeSnapshotDeltaTokenList contains a list of VolumeSnapshotDeltaToken
type VolumeSnapshotDeltaTokenList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []VolumeSnapshotDeltaToken `json:"items"`
}

func init() {
	SchemeBuilder.Register(&VolumeSnapshotDeltaToken{}, &VolumeSnapshotDeltaTokenList{})
}