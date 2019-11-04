package main

import (
	"context"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"go.opencensus.io/trace"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ csi.IdentityServer = &identityServer{}

type identityServer struct{}

func (s *identityServer) GetPluginInfo(context.Context, *csi.GetPluginInfoRequest) (*csi.GetPluginInfoResponse, error) {
	_, span := trace.StartSpan(context.Background(), "xyz.alexeldeib.csi.identity.GetPluginInfo")
	defer span.End()
	return nil, status.Error(codes.Unimplemented, "")
}

func (s *identityServer) GetPluginCapabilities(context.Context, *csi.GetPluginCapabilitiesRequest) (*csi.GetPluginCapabilitiesResponse, error) {
	_, span := trace.StartSpan(context.Background(), "xyz.alexeldeib.csi.identity.GetPluginCapabilities")
	defer span.End()
	return nil, status.Error(codes.Unimplemented, "")
}

func (s *identityServer) Probe(context.Context, *csi.ProbeRequest) (*csi.ProbeResponse, error) {
	_, span := trace.StartSpan(context.Background(), "xyz.alexeldeib.csi.identity.Probe")
	defer span.End()
	return nil, status.Error(codes.Unimplemented, "")
}
