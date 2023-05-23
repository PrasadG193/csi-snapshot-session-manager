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
	"errors"
	"fmt"
	"os"
	"time"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	cbtv1alpha1 "github.com/PrasadG193/external-snapshot-session-access/pkg/api/cbt/v1alpha1"
	"github.com/go-logr/logr"
	"github.com/google/uuid"
)

const CSIEndpointEnvName = "CSI_ENDPOINT"

// CSISnapshotSessionAccessReconciler reconciles a CSISnapshotSessionAccess object
type CSISnapshotSessionAccessReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=cbt.storage.k8s.io,resources=csisnapshotsessionaccesses,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=cbt.storage.k8s.io,resources=csisnapshotsessionaccesses/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the CSISnapshotSessionAccess object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.1/pkg/reconcile
func (r *CSISnapshotSessionAccessReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	vsdt := &cbtv1alpha1.CSISnapshotSessionAccess{}
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
func (r *CSISnapshotSessionAccessReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&cbtv1alpha1.CSISnapshotSessionAccess{}).
		Complete(r)
}

// Create creates a new version of a resource.
func (r *CSISnapshotSessionAccessReconciler) handleEvents(
	ctx context.Context,
	obj *cbtv1alpha1.CSISnapshotSessionAccess,
	logger logr.Logger) error {

	//casted.SetCreationTimestamp(metav1.Now())
	//out := casted.DeepCopy()
	if obj.Status.SessionState == cbtv1alpha1.SessionStateTypeReady ||
		obj.Status.SessionState == cbtv1alpha1.SessionStateTypeFailed {
		return nil
	}

	// Fake token generation delay
	time.Sleep(10 * time.Second)

	reqID := uuid.New().String()
	token := NewToken(reqID)
	ca, err := fetchCABundle()
	if err != nil {
		return err
	}
	obj.Status = cbtv1alpha1.CSISnapshotSessionAccessStatus{
		SessionState: cbtv1alpha1.SessionStateTypeReady,
		SessionToken: token.Token,
		SessionURL:   token.URL,
		CACert:       []byte(ca),
	}
	//status, err := mockSessionToken(ctx, obj.Spec.BaseVolumeSnapshotName, obj.Spec.TargetVolumeSnapshotName)
	//if err != nil {
	//	return err
	//}
	//obj.Status = status
	err = r.Update(ctx, obj)
	if err != nil {
		return err
	}
	logger.Info(fmt.Sprintf("created CSISnapshotSessionAccess: %s", obj.GetName()))
	return nil
}

func fetchCABundle() ([]byte, error) {
	cacertFile := os.Getenv("CBT_SERVER_CA_BUNDLE")
	if cacertFile == "" {
		return nil, errors.New("Failed to read CA Bundle from " + cacertFile)
	}
	return os.ReadFile(cacertFile)
}

//func mockSessionToken(ctx context.Context, baseSnapName, targetSnapName string) (cbtv1alpha1.CSISnapshotSessionAccessStatus, error) {
//	// TODO: Store the session params in a CR
//	cacert, err := fetchCABundle()
//	if err != nil {
//		return nil, err
//	}
//	return cbtv1alpha1.CSISnapshotSessionAccessStatus{
//		SessionState: cbtv1alpha1.SessionStateTypeReady,
//		SessionToken: uuid.New().String(),
//		SessionURL:   os.Getenv(CSIEndpointEnvName),
//		CACert:       string(cacert),
//	}, nil
//}

// TODO: Set the correct state of the request - InProgress,
// SessionResponse, session state and error
func fetchSessionToken(ctx context.Context, baseSnapName, targetSnapName string) (cbtv1alpha1.CSISnapshotSessionAccessStatus, error) {
	// FIXME Add sleep of 10s to mock CBT session creation delay
	time.Sleep(10 * time.Second)
	csiEndpoint := os.Getenv(CSIEndpointEnvName)
	fmt.Println("Invoking gRPC on", csiEndpoint)
	client := NewCSIClient(csiEndpoint)
	tokenResp, err := client.FetchSessionToken(ctx, baseSnapName, targetSnapName)
	if err != nil {
		return cbtv1alpha1.CSISnapshotSessionAccessStatus{}, err
	}
	return cbtv1alpha1.CSISnapshotSessionAccessStatus{
		SessionState: cbtv1alpha1.SessionStateTypeReady,
		SessionToken: tokenResp.SessionToken,
		SessionURL:   tokenResp.SessionUrl,
		CACert:       []byte(tokenResp.CaCert),
	}, nil
}
