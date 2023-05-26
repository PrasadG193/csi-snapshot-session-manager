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

// CSISnapshotSessionDataSpec defines the desired state of CSISnapshotSessionData
type CSISnapshotSessionDataSpec struct {
	Expiry       *metav1.Time `json:"expiryTime,nomitempty"`
	SessionToken string       `json:"sessionToken,omitempty"`
	Snapshots    []Snapshot   `json:"snapshots,omitempty"`
	Volumes      []Volume     `json:"volumes,omitempty"`
}

// Volume
type Volume struct {
	ID   string `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}

// Snapshot
type Snapshot struct {
	ID     string `json:"id,omitempty"`
	Name   string `json:"name,omitempty"`
	Volume string `json:"volume,omitempty"`
}

//+kubebuilder:object:root=true

// CSISnapshotSessionData is the Schema for the csisnapshotsessions API
type CSISnapshotSessionData struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec CSISnapshotSessionDataSpec `json:"spec,omitempty"`
}

//+kubebuilder:object:root=true

// CSISnapshotSessionDataList contains a list of CSISnapshotSessionData
type CSISnapshotSessionDataList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CSISnapshotSessionData `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CSISnapshotSessionData{}, &CSISnapshotSessionDataList{})
}
