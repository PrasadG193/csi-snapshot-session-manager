package controller

import (
	"context"
	"fmt"
	"log"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/kubernetes-csi/csi-lib-utils/connection"
	"github.com/kubernetes-csi/csi-lib-utils/metrics"
)

type CSIClient struct {
	client csi.ControllerClient
}

func NewCSIClient(url string) *CSIClient {
	fmt.Println("DEBUG:: connecting to plugin")
	metricsManager := metrics.NewCSIMetricsManagerWithOptions("", /* driverName */
		// Will be provided via default gatherer.
		metrics.WithProcessStartTime(false),
		metrics.WithSubsystem(metrics.SubsystemSidecar),
	)
	conn, err := connection.Connect(url, metricsManager)
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	client := csi.NewControllerClient(conn)
	fmt.Println("DEBUG:: Connection successfull")
	return &CSIClient{client: client}
}

func (c *CSIClient) FetchSessionToken(ctx context.Context, baseSnapName, targetSnapName string) (*csi.CreateVolumeSnapshotDeltaSessionTokenResponse, error) {
	req := csi.CreateVolumeSnapshotDeltaSessionTokenRequest{
		BaseVolumeSnapshotName:   baseSnapName,
		TargetVolumeSnapshotName: targetSnapName,
	}
	fmt.Printf("DEBUG:: gRPC Request: %#v\n", req)
	resp, err := c.client.CreateVolumeSnapshotDeltaSessionToken(ctx, &req)
	if err != nil {
		return nil, fmt.Errorf("Insert failure: %w", err)
	}
	return resp, nil
}
