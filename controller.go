package main

import (
	"context"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"go.opencensus.io/trace"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ csi.ControllerServer = &controllerServer{}

type controllerServer struct{}

func (s *controllerServer) ControllerGetCapabilities(context.Context, *csi.ControllerGetCapabilitiesRequest) (*csi.ControllerGetCapabilitiesResponse, error) {
	_, span := trace.StartSpan(context.Background(), "xyz.alexeldeib.csi.controller.ControllerGetCapabilities")
	defer span.End()
	return nil, status.Error(codes.Unimplemented, "")
}

func (s *controllerServer) ListVolumes(context.Context, *csi.ListVolumesRequest) (*csi.ListVolumesResponse, error) {
	_, span := trace.StartSpan(context.Background(), "xyz.alexeldeib.csi.controller.ListVolumes")
	defer span.End()
	return nil, status.Error(codes.Unimplemented, "")
}

func (s *controllerServer) ValidateVolumeCapabilities(context.Context, *csi.ValidateVolumeCapabilitiesRequest) (*csi.ValidateVolumeCapabilitiesResponse, error) {
	_, span := trace.StartSpan(context.Background(), "xyz.alexeldeib.csi.controller.ValidateVolumeCapabilities")
	defer span.End()
	return nil, status.Error(codes.Unimplemented, "")
}

func (s *controllerServer) CreateVolume(context.Context, *csi.CreateVolumeRequest) (*csi.CreateVolumeResponse, error) {
	_, span := trace.StartSpan(context.Background(), "xyz.alexeldeib.csi.controller.CreateVolume")
	defer span.End()
	return nil, status.Error(codes.Unimplemented, "")
}

func (s *controllerServer) DeleteVolume(context.Context, *csi.DeleteVolumeRequest) (*csi.DeleteVolumeResponse, error) {
	_, span := trace.StartSpan(context.Background(), "xyz.alexeldeib.csi.controller.DeleteVolume")
	defer span.End()
	return nil, status.Error(codes.Unimplemented, "")
}

func (s *controllerServer) ControllerPublishVolume(context.Context, *csi.ControllerPublishVolumeRequest) (*csi.ControllerPublishVolumeResponse, error) {
	_, span := trace.StartSpan(context.Background(), "xyz.alexeldeib.csi.controller.ControllerPublishVolume")
	defer span.End()
	return nil, status.Error(codes.Unimplemented, "")
}

func (s *controllerServer) ControllerUnpublishVolume(context.Context, *csi.ControllerUnpublishVolumeRequest) (*csi.ControllerUnpublishVolumeResponse, error) {
	_, span := trace.StartSpan(context.Background(), "xyz.alexeldeib.csi.controller.ControllerUnpublishVolume")
	defer span.End()
	return nil, status.Error(codes.Unimplemented, "")
}

func (s *controllerServer) GetCapacity(context.Context, *csi.GetCapacityRequest) (*csi.GetCapacityResponse, error) {
	_, span := trace.StartSpan(context.Background(), "xyz.alexeldeib.csi.controller.GetCapacity")
	defer span.End()
	return nil, status.Error(codes.Unimplemented, "")
}

func (s *controllerServer) CreateSnapshot(context.Context, *csi.CreateSnapshotRequest) (*csi.CreateSnapshotResponse, error) {
	_, span := trace.StartSpan(context.Background(), "xyz.alexeldeib.csi.controller.CreateSnapshot")
	defer span.End()
	return nil, status.Error(codes.Unimplemented, "")
}

func (s *controllerServer) DeleteSnapshot(context.Context, *csi.DeleteSnapshotRequest) (*csi.DeleteSnapshotResponse, error) {
	_, span := trace.StartSpan(context.Background(), "xyz.alexeldeib.csi.controller.DeleteSnapshot")
	defer span.End()
	return nil, status.Error(codes.Unimplemented, "")
}

func (s *controllerServer) ListSnapshots(context.Context, *csi.ListSnapshotsRequest) (*csi.ListSnapshotsResponse, error) {
	_, span := trace.StartSpan(context.Background(), "xyz.alexeldeib.csi.controller.ListSnapshots")
	defer span.End()
	return nil, status.Error(codes.Unimplemented, "")
}

func (s *controllerServer) ControllerExpandVolume(context.Context, *csi.ControllerExpandVolumeRequest) (*csi.ControllerExpandVolumeResponse, error) {
	_, span := trace.StartSpan(context.Background(), "xyz.alexeldeib.csi.controller.ControllerExpandVolume")
	defer span.End()
	return nil, status.Error(codes.Unimplemented, "")
}
