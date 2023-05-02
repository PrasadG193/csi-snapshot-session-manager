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
	"encoding/json"
	"fmt"
	"net/http"

	admissionv1 "k8s.io/api/admission/v1"
	authnv1 "k8s.io/api/authentication/v1"
	authzv1 "k8s.io/api/authorization/v1"
	v1 "k8s.io/api/authorization/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// log is for logging in this package.
var volumesnapshotdeltatokenlog = logf.Log.WithName("volumesnapshotdeltatoken-resource")

func (r *CSISnapshotSessionAccess) SetupWebhookWithManager(mgr ctrl.Manager) error {
	volumesnapshotdeltatokenlog.Info("Registering Webhook handler.")
	validator := &CSISnapshotSessionAccessValidator{}
	whServer := mgr.GetWebhookServer()
	// TODO: Declare as const
	whServer.Register("/validate-cbt-storage-k8s-io-v1alpha1-volumesnapshotdeltatoken", &webhook.Admission{Handler: validator})
	whServer.CertDir = "/tmp/k8s-webhook-server/serving-certs/"
	whServer.Port = 9443

	decoder, err := admission.NewDecoder(mgr.GetScheme())
	if err != nil {
		return err
	}
	validator.decoder = decoder
	cfg, err := ctrl.GetConfig()
	if err != nil {
		return err
	}
	kubeClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return err
	}
	validator.cli = kubeClient
	return nil
}

var _ admission.Handler = &CSISnapshotSessionAccessValidator{}

// +kubebuilder:webhook:path=/validate-cbt-storage-k8s-io-v1alpha1-volumesnapshotdeltatoken,mutating=true,failurePolicy=fail,sideEffects=None,groups=cbt.storage.k8s.io,resources=csisnapshotsessionaccesses,verbs=create;update,versions=v1alpha1,name=vvolumesnapshotdeltatoken.kb.io,admissionReviewVersions=v1
// +kubebuilder:object:generate=false
type CSISnapshotSessionAccessValidator struct {
	decoder *admission.Decoder
	cli     kubernetes.Interface
}

func (v *CSISnapshotSessionAccessValidator) authorizeUser(ctx context.Context, req admission.Request) (bool, error) {
	extra := make(map[string]v1.ExtraValue, len(req.UserInfo.Extra))
	for u, e := range req.UserInfo.Extra {
		extra[u] = v1.ExtraValue(e)
	}
	allowedVS, err := v.canAccessVolumeSnapshots(ctx, req.Namespace, req.UserInfo, extra)
	if err != nil {
		return false, err
	}
	allowedPVC, err := v.canAccessPVC(ctx, req.Namespace, req.UserInfo, extra)
	if err != nil {
		return false, err
	}
	return allowedVS || allowedPVC, nil
}

func (v *CSISnapshotSessionAccessValidator) canAccessVolumeSnapshots(ctx context.Context, namespace string, userInfo authnv1.UserInfo, extraValues map[string]authzv1.ExtraValue) (bool, error) {
	return v.subjectAccessReview(ctx, namespace, userInfo, extraValues, "get", metav1.GroupVersionResource{Group: "snapshot.storage.k8s.io", Version: "v1", Resource: "volumesnapshots"})
}

func (v *CSISnapshotSessionAccessValidator) canAccessPVC(ctx context.Context, namespace string, userInfo authnv1.UserInfo, extraValues map[string]authzv1.ExtraValue) (bool, error) {
	return v.subjectAccessReview(ctx, namespace, userInfo, extraValues, "get", metav1.GroupVersionResource{Group: "", Version: "v1", Resource: "persistentvolumeclaims"})
}

func (v *CSISnapshotSessionAccessValidator) subjectAccessReview(ctx context.Context, namespace string, userInfo authnv1.UserInfo, extraValues map[string]authzv1.ExtraValue, verb string, gvr metav1.GroupVersionResource) (bool, error) {
	sar := &v1.SubjectAccessReview{
		Spec: v1.SubjectAccessReviewSpec{
			ResourceAttributes: &v1.ResourceAttributes{
				Verb:      verb,
				Namespace: namespace,
				Group:     gvr.Group,
				Version:   gvr.Version,
				Resource:  gvr.Resource,
			},
			User:   userInfo.Username,
			Groups: userInfo.Groups,
			Extra:  extraValues,
			UID:    userInfo.UID,
		},
	}
	sarResp, err := v.cli.AuthorizationV1().SubjectAccessReviews().Create(ctx, sar, metav1.CreateOptions{})
	if err != nil {
		//return admission.Errored(http.StatusBadRequest, err)
		return false, err
	}
	//sarJson, err := sarResp.Status.Marshal()
	//fmt.Printf("debug: SAR Response:: %s %v\n", string(sarJson), err)
	if !sarResp.Status.Allowed || sarResp.Status.Denied {
		return false, nil
		//return admission.Response{
		//	AdmissionResponse: admissionv1.AdmissionResponse{
		//		Allowed: false,
		//		Result:  &metav1.Status{},
		//	},
		//}
	}
	return true, nil
}

func (v *CSISnapshotSessionAccessValidator) Handle(ctx context.Context, req admission.Request) admission.Response {
	fmt.Printf("USERINFO::: %#v\n", req.UserInfo)
	// Do the validation
	vsd := &CSISnapshotSessionAccess{}
	err := v.decoder.Decode(req, vsd)
	if err != nil {
		volumesnapshotdeltatokenlog.Error(err, "Failed to decode request object")
		return admission.Errored(http.StatusBadRequest, err)
	}

	// Verify if the user has permission to GET volumesnapshot resources
	//userInfo, err := userToInfo(req.UserInfo)
	//if err != nil {
	//	volumesnapshotdeltatokenlog.Error(err, "Failed to decode request object")
	//	return admission.Errored(http.StatusBadRequest, err)
	//}
	//att :=
	//fmt.Println(att)
	authz, err := v.authorizeUser(ctx, req)
	if err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}
	if !authz {
		return admission.Response{
			AdmissionResponse: admissionv1.AdmissionResponse{
				Allowed: false,
				Result:  &metav1.Status{},
			},
		}
	}
	vsd.Status.SessionState = SessionStateTypePending
	marshaledObject, err := json.Marshal(runtime.Object(vsd))
	if err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}
	fmt.Println("debug: webhook req object Raw:: ", string(req.Object.Raw))
	fmt.Println("debug: webhook result:: ", string(marshaledObject))
	patched := admission.PatchResponseFromRaw(req.Object.Raw, marshaledObject)
	fmt.Printf("debug: webhook patched response:: %#v\n", patched)
	return patched
}

//func getResourceAttributes(namespace string) *v1.ResourceAttributes {
//	apiVerb := "get"
//	vsGroup := "snapshot.storage.k8s.io"
//	vsVersion := "v1"
//	vsResource := "volumesnapshots"
//	return &v1.ResourceAttributes{
//		Verb:      apiVerb,
//		Namespace: namespace,
//		Group:     vsGroup,
//		Version:   vsVersion,
//		Resource:  vsResource,
//	}
//}
