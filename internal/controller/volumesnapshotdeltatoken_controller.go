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

package controller

import (
	"context"
	"fmt"
	"os"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	cbtv1alpha1 "github.com/PrasadG193/cbt-datapath/pkg/api/cbt/v1alpha1"
	"github.com/go-logr/logr"
)

const CSIEndpointEnvName = "CSI_ENDPOINT"

// VolumeSnapshotDeltaTokenReconciler reconciles a VolumeSnapshotDeltaToken object
type VolumeSnapshotDeltaTokenReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=cbt.storage.k8s.io,resources=volumesnapshotdeltatokens,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=cbt.storage.k8s.io,resources=volumesnapshotdeltatokens/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=cbt.storage.k8s.io,resources=volumesnapshotdeltatokens/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the VolumeSnapshotDeltaToken object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.1/pkg/reconcile
func (r *VolumeSnapshotDeltaTokenReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	vsdt := &cbtv1alpha1.VolumeSnapshotDeltaToken{}
	err := r.Get(ctx, req.NamespacedName, vsdt)
	if err != nil {
		if apierrors.IsNotFound(err) {
			// Object not found, return.  Created objects are automatically garbage collected.
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the req.
		return reconcile.Result{}, err
	}

	return ctrl.Result{}, r.handleEvents(ctx, vsdt, logger)
}

// SetupWithManager sets up the controller with the Manager.
func (r *VolumeSnapshotDeltaTokenReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&cbtv1alpha1.VolumeSnapshotDeltaToken{}).
		Complete(r)
}

// Create creates a new version of a resource.
func (r *VolumeSnapshotDeltaTokenReconciler) handleEvents(
	ctx context.Context,
	obj *cbtv1alpha1.VolumeSnapshotDeltaToken,
	logger logr.Logger) error {

	//casted.SetCreationTimestamp(metav1.Now())
	//out := casted.DeepCopy()
	if obj.Status.SessionState == cbtv1alpha1.SessionStateTypeReady ||
		obj.Status.SessionState == cbtv1alpha1.SessionStateTypeFailed {
		return nil
	}

	//reqID := uuid.New().String()
	//token := NewToken(reqID)
	//ca, err := fetchCABundle()
	//ca := "xxxxxxxxx"
	//if err != nil {
	//	return nil, err
	//}
	//obj.Status = cbtv1alpha1.VolumeSnapshotDeltaTokenStatus{
	//	SessionState: cbtv1alpha1.SessionStateTypeReady,
	//	SessionToken:        token.Token,
	//	SessionURL:          token.URL,
	//	CABundle:     []byte(ca),
	//}
	status, err := fetchSessionToken(ctx, obj.Spec.BaseVolumeSnapshotName, obj.Spec.TargetVolumeSnapshotName)
	if err != nil {
		return err
	}
	obj.Status = status
	err = r.Status().Update(ctx, obj)
	if err != nil {
		return err
	}
	logger.Info(fmt.Sprintf("created VolumeSnapshotDeltaToken: %s", obj.GetName()))

	return nil
}

//func fetchCABundle() ([]byte, error) {
//	cacertFile := os.Getenv("CBT_SERVER_CA_BUNDLE")
//	if cacertFile == "" {
//		return nil, errors.New("Failed to read CA Bundle from " + cacertFile)
//	}
//	return os.ReadFile(cacertFile)
//}

// TODO: Set the correct state of the request - InProgress,
// SessionResponse, session state and error
func fetchSessionToken(ctx context.Context, baseSnapName, targetSnapName string) (cbtv1alpha1.VolumeSnapshotDeltaTokenStatus, error) {
	csiEndpoint := os.Getenv(CSIEndpointEnvName)
	fmt.Println("Invoking gRPC on", csiEndpoint)
	client := NewCSIClient(csiEndpoint)
	tokenResp, err := client.FetchSessionToken(ctx, baseSnapName, targetSnapName)
	if err != nil {
		return cbtv1alpha1.VolumeSnapshotDeltaTokenStatus{}, err
	}
	return cbtv1alpha1.VolumeSnapshotDeltaTokenStatus{
		SessionState: cbtv1alpha1.SessionStateTypeReady,
		SessionToken: tokenResp.SessionToken,
		SessionURL:   tokenResp.SessionUrl,
		CACert:       []byte(tokenResp.CaCert),
	}, nil
}
