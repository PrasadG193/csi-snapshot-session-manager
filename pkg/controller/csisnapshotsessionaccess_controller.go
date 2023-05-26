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

	//"github.com/docker/docker/daemon/logger"

	volsnapv1 "github.com/kubernetes-csi/external-snapshotter/client/v4/apis/volumesnapshot/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	cbtv1alpha1 "github.com/PrasadG193/external-snapshot-session-access/pkg/api/cbt/v1alpha1"
	"github.com/go-logr/logr"
)

const CSIEndpointEnvName = "CSI_ENDPOINT"

// CSISnapshotSessionAccessReconciler reconciles a CSISnapshotSessionAccess object
type CSISnapshotSessionAccessReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

type vsInfo struct {
	volumeName string
	volumeID   string
	snapshotID string
	driverName string
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
		// TODO: Return if expired and pending
		// Set object state to Failed/Expired
		return nil
	}

	// TODO: Set initial expiry time and retry

	// 1. Discover driver name
	vsInfo, err := r.volumeSnapshotsInfo(ctx, obj.Spec.Snapshots, obj.GetNamespace())
	if err != nil {
		return err
	}
	// 2. Find CSISnapshotSessionService object for the driver
	driver := vsInfo[obj.Spec.Snapshots[0]].driverName
	sss, err := r.findSnapSessionService(ctx, logger, driver)
	if err != nil {
		return err
	}

	// 3. Generate CSISnapshotSessionData object
	token := newToken()
	_, err1 := r.storeSessionData(ctx, logger, sss.GetNamespace(), token, driver, vsInfo)
	if err1 != nil {
		return err1
	}

	obj.Status = cbtv1alpha1.CSISnapshotSessionAccessStatus{
		ExpiryTime:   nil,
		SessionState: cbtv1alpha1.SessionStateTypeReady,
		SessionToken: []byte(token),
		SessionURL:   sss.Spec.Address,
		CACert:       sss.Spec.CACert,
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

func (r *CSISnapshotSessionAccessReconciler) storeSessionData(ctx context.Context, logger logr.Logger, namespace, token, driver string, vsInfoMap map[string]vsInfo) (*cbtv1alpha1.CSISnapshotSessionData, error) {
	expiry := metav1.NewTime(time.Now().Add(time.Minute * time.Duration(10)))
	ssd := &cbtv1alpha1.CSISnapshotSessionData{
		ObjectMeta: metav1.ObjectMeta{
			Name:      SnapSessionDataNameWithToken(token),
			Namespace: namespace,
		},
		Spec: cbtv1alpha1.CSISnapshotSessionDataSpec{
			Expiry: &expiry,
			// Considering all snapshots point to the same volume, we will just add single element to the volume array
		},
	}
	for vs, info := range vsInfoMap {
		ssd.Spec.Volumes = append(ssd.Spec.Volumes, cbtv1alpha1.Volume{ID: info.volumeID, Name: info.volumeName})
		ssd.Spec.Snapshots = append(ssd.Spec.Snapshots, cbtv1alpha1.Snapshot{ID: info.snapshotID, Name: vs, Volume: info.volumeName})
	}
	err := r.Create(ctx, ssd)
	logger.Info(fmt.Sprintf("created CSISnapshotSessionData: %s/%s", ssd.GetNamespace(), ssd.GetName()))
	return ssd, err
}

func (r *CSISnapshotSessionAccessReconciler) volumeSnapshotsInfo(ctx context.Context, snaps []string, namespace string) (map[string]vsInfo, error) {
	vsInfoMap := make(map[string]vsInfo, len(snaps))
	for _, vsName := range snaps {
		vs := &volsnapv1.VolumeSnapshot{}
		err := r.Get(ctx, types.NamespacedName{Name: vsName, Namespace: namespace}, vs)
		if err != nil {
			return nil, err
		}
		// FIXME: Check if the vs is ready
		vsc := &volsnapv1.VolumeSnapshotContent{}
		err1 := r.Get(ctx, types.NamespacedName{Name: *vs.Status.BoundVolumeSnapshotContentName, Namespace: namespace}, vsc)
		if err1 != nil {
			return nil, err1
		}
		pvc := &corev1.PersistentVolumeClaim{}
		err2 := r.Get(ctx, types.NamespacedName{Name: *vs.Spec.Source.PersistentVolumeClaimName, Namespace: namespace}, pvc)
		if err2 != nil {
			return nil, err2
		}
		// TODO: Verify that all the vs refers to the same driver
		vsInfoMap[vsName] = vsInfo{
			driverName: vsc.Spec.Driver,
			//volumeName: pvc.Spec.VolumeName,
			volumeID:   *vsc.Spec.Source.VolumeHandle,
			snapshotID: *vsc.Status.SnapshotHandle,
		}
	}
	return vsInfoMap, nil
}

func (r *CSISnapshotSessionAccessReconciler) findSnapSessionService(ctx context.Context, logger logr.Logger, driver string) (*cbtv1alpha1.CSISnapshotSessionService, error) {
	sssList := &cbtv1alpha1.CSISnapshotSessionServiceList{}

	sssReq, err := labels.NewRequirement("cbt.storage.k8s.io/driver", selection.Equals, []string{driver})
	if err != nil {
		return nil, err
	}
	err1 := r.List(ctx, sssList, &client.ListOptions{LabelSelector: labels.NewSelector().Add(*sssReq)})
	if err1 != nil {
		return nil, err1
	}

	if len(sssList.Items) == 0 {
		return nil, nil
	}
	// TODO: Handle multiple obj
	logger.Info(fmt.Sprintf("found CSISnapshotSessionService object %s for driver: %s", sssList.Items[0].GetName(), driver))
	return &sssList.Items[0], nil
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
