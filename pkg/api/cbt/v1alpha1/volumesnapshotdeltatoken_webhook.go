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

func (r *VolumeSnapshotDeltaToken) SetupWebhookWithManager(mgr ctrl.Manager) error {
	volumesnapshotdeltatokenlog.Info("Registering Webhook handler.")
	validator := &VolumeSnapshotDeltaTokenValidator{}
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

var _ admission.Handler = &VolumeSnapshotDeltaTokenValidator{}

// +kubebuilder:webhook:path=/validate-cbt-storage-k8s-io-v1alpha1-volumesnapshotdeltatoken,mutating=true,failurePolicy=fail,sideEffects=None,groups=cbt.storage.k8s.io,resources=volumesnapshotdeltatokens,verbs=create;update,versions=v1alpha1,name=vvolumesnapshotdeltatoken.kb.io,admissionReviewVersions=v1
// +kubebuilder:object:generate=false
type VolumeSnapshotDeltaTokenValidator struct {
	decoder *admission.Decoder
	cli     kubernetes.Interface
}

func (v *VolumeSnapshotDeltaTokenValidator) Handle(ctx context.Context, req admission.Request) admission.Response {
	fmt.Printf("USERINFO::: %#v\n", req.UserInfo)
	// Do the validation
	vsd := &VolumeSnapshotDeltaToken{}
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
	extra := make(map[string]v1.ExtraValue, len(req.UserInfo.Extra))
	for u, e := range req.UserInfo.Extra {
		extra[u] = v1.ExtraValue(e)
	}

	sar := &v1.SubjectAccessReview{
		Spec: v1.SubjectAccessReviewSpec{
			ResourceAttributes: GetResourceAttributes(req.Namespace),
			User:               req.UserInfo.Username,
			Groups:             req.UserInfo.Groups,
			Extra:              extra,
			UID:                req.UserInfo.UID,
		},
	}
	sarResp, err := v.cli.AuthorizationV1().SubjectAccessReviews().Create(ctx, sar, metav1.CreateOptions{})
	if err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}
	//sarJson, err := sarResp.Status.Marshal()
	//fmt.Printf("debug: SAR Response:: %s %v\n", string(sarJson), err)
	if !sarResp.Status.Allowed || sarResp.Status.Denied {
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

//// TODO: Find if there is better way to conver auth.UserInfo to user.Info
//func userToInfo(userInfo v1.UserInfo) (user.Info, error) {
//	//b, err := userInfo.Marshal()
//	//if err != nil {
//	//	return nil, err
//	//}
//	//ui := user.DefaultInfo{}
//	//err = json.Unmarshal(b, &ui)
//	//return &ui, err
//	defInfo := &user.DefaultInfo{
//		Name:   userInfo.Username,
//		UID:    userInfo.UID,
//		Groups: userInfo.Groups,
//		Extra:  make(map[string]string),
//	}
//	for k, v := range userInfo.Extra {
//		defInfo.Extra[k] = string(v)
//	}
//	return defInfo, nil
//}

func GetResourceAttributes(namespace string) *v1.ResourceAttributes {
	apiVerb := "get"
	vsGroup := "snapshot.storage.k8s.io"
	vsVersion := "v1"
	vsResource := "volumesnapshots"
	return &v1.ResourceAttributes{
		Verb:      apiVerb,
		Namespace: namespace,
		Group:     vsGroup,
		Version:   vsVersion,
		Resource:  vsResource,
	}
}
