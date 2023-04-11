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
	"context"
	"fmt"

	admissionv1 "k8s.io/api/admission/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// log is for logging in this package.
var volumesnapshotdeltatokenlog = logf.Log.WithName("volumesnapshotdeltatoken-resource")

func (r *VolumeSnapshotDeltaToken) SetupWebhookWithManager(mgr ctrl.Manager) error {
	volumesnapshotdeltatokenlog.Info("Registering Webhook handler.")
	whServer := mgr.GetWebhookServer()
	// TODO: Declare as const
	whServer.Register("/validate-cbt-storage-k8s-io-v1alpha1-volumesnapshotdeltatoken", &webhook.Admission{Handler: &VolumeSnapshotDeltaToken{}})
	whServer.CertDir = "/tmp/k8s-webhook-server/serving-certs/"
	whServer.Port = 9443
	return nil
}

// +kubebuilder:webhook:path=/validate-cbt-storage-k8s-io-v1alpha1-volumesnapshotdeltatoken,mutating=false,failurePolicy=fail,sideEffects=None,groups=cbt.storage.k8s.io,resources=volumesnapshotdeltatokens,verbs=create;update,versions=v1alpha1,name=vvolumesnapshotdeltatoken.kb.io,admissionReviewVersions=v1
var _ admission.Handler = &VolumeSnapshotDeltaToken{}

func (a *VolumeSnapshotDeltaToken) Handle(ctx context.Context, req admission.Request) admission.Response {
	fmt.Printf("USERINFO::: %#v\n", req.UserInfo)
	return admission.Response{
		AdmissionResponse: admissionv1.AdmissionResponse{
			Allowed: true,
			Result:  &metav1.Status{},
		},
	}
}
