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

package main

import (
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog"
	"sigs.k8s.io/apiserver-runtime/pkg/builder"

	// +kubebuilder:scaffold:resource-imports
	"github.com/PrasadG193/cbt-svc/pkg/apis/cbt/v1alpha1"
	cbtv1alpha1 "github.com/PrasadG193/cbt-svc/pkg/apis/cbt/v1alpha1"
	cbtstorage "github.com/PrasadG193/cbt-svc/pkg/storage"
)

func main() {
	scheme := runtime.NewScheme()
	scheme.AddKnownTypes(v1alpha1.SchemeGroupVersion, &v1alpha1.VolumeSnapshotDeltaToken{})
	apiserver := builder.APIServer.
		WithResourceAndHandler(&cbtv1alpha1.VolumeSnapshotDeltaToken{}, cbtstorage.CBTHandlerProvider).
		//WithAdditionalSchemeInstallers(v1alpha1.RegisterDefaults).
		//WithAdditionalSchemeInstallers(v1alpha1.RegisterConversions).
		WithoutEtcd().
		WithLocalDebugExtension()

	if err := apiserver.Execute(); err != nil {
		klog.Fatal(err)
	}
}
