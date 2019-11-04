package main

import (
	"context"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ csi.NodeServer = &nodeServer{}

type nodeServer struct{}

func newNodeServer() *nodeServer {
	return &nodeServer{}
}

func (s *nodeServer) NodePublishVolume(ctx context.Context, req *csi.NodePublishVolumeRequest) (*csi.NodePublishVolumeResponse, error) {
	// _, span := trace.StartSpan(context.Background(), "xyz.alexeldeib.csi.node.NodePublishVolume")
	// defer span.End()
	return nil, status.Error(codes.Unimplemented, "")
}

func (s *nodeServer) NodeUnpublishVolume(ctx context.Context, req *csi.NodeUnpublishVolumeRequest) (*csi.NodeUnpublishVolumeResponse, error) {
	// _, span := trace.StartSpan(context.Background(), "xyz.alexeldeib.csi.node.NodeUnpublishVolume")
	// defer span.End()
	return nil, status.Error(codes.Unimplemented, "")
}

func (s *nodeServer) NodeStageVolume(context.Context, *csi.NodeStageVolumeRequest) (*csi.NodeStageVolumeResponse, error) {
	// _, span := trace.StartSpan(context.Background(), "xyz.alexeldeib.csi.node.NodeStageVolume")
	// defer span.End()
	return nil, status.Error(codes.Unimplemented, "")
}

func (s *nodeServer) NodeUnstageVolume(context.Context, *csi.NodeUnstageVolumeRequest) (*csi.NodeUnstageVolumeResponse, error) {
	// _, span := trace.StartSpan(context.Background(), "xyz.alexeldeib.csi.node.NodeUnstageVolume")
	// defer span.End()
	return nil, status.Error(codes.Unimplemented, "")
}

func (s *nodeServer) NodeGetVolumeStats(context.Context, *csi.NodeGetVolumeStatsRequest) (*csi.NodeGetVolumeStatsResponse, error) {
	// _, span := trace.StartSpan(context.Background(), "xyz.alexeldeib.csi.node.NodeGetVolumeStats")
	// defer span.End()
	return nil, status.Error(codes.Unimplemented, "")
}

func (s *nodeServer) NodeExpandVolume(context.Context, *csi.NodeExpandVolumeRequest) (*csi.NodeExpandVolumeResponse, error) {
	// _, span := trace.StartSpan(context.Background(), "xyz.alexeldeib.csi.node.NodeExpandVolume")
	// defer span.End()
	return nil, status.Error(codes.Unimplemented, "")
}

func (s *nodeServer) NodeGetCapabilities(context.Context, *csi.NodeGetCapabilitiesRequest) (*csi.NodeGetCapabilitiesResponse, error) {
	// _, span := trace.StartSpan(context.Background(), "xyz.alexeldeib.csi.node.NodeGetCapabilities")
	// defer span.End()
	return nil, status.Error(codes.Unimplemented, "")
}

func (s *nodeServer) NodeGetInfo(context.Context, *csi.NodeGetInfoRequest) (*csi.NodeGetInfoResponse, error) {
	// _, span := trace.StartSpan(context.Background(), "xyz.alexeldeib.csi.node.NodeGetInfo")
	// defer span.End()
	return nil, status.Error(codes.Unimplemented, "")
}
