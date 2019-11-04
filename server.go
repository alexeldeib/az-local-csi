package main

import (
	"net/http"
	"time"

	"github.com/container-storage-interface/spec/lib/go/csi"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	grpc_opentracing "github.com/grpc-ecosystem/go-grpc-middleware/tracing/opentracing"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"go.opencensus.io/plugin/ocgrpc"
	"go.opencensus.io/plugin/ochttp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

func newHTTPServer() *http.Server {
	router := http.NewServeMux()
	router.Handle("/healthz", instrument("/healthz", healthz))
	router.Handle("/livez", instrument("/livez", livez))
	router.Handle("/readyz", instrument("/readyz", readyz))

	return &http.Server{
		Addr:         serverAddress,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		Handler: &ochttp.Handler{
			Handler: router,
		},
	}
}

func newGRPCServer() *grpc.Server {
	grpcServer := grpc.NewServer(
		grpc.KeepaliveParams(keepalive.ServerParameters{MaxConnectionAge: 2 * time.Minute}),
		grpc.StatsHandler(&ocgrpc.ServerHandler{}),
		grpc.StreamInterceptor(grpc_middleware.ChainStreamServer(
			grpc_ctxtags.StreamServerInterceptor(),
			grpc_opentracing.StreamServerInterceptor(),
			grpc_prometheus.StreamServerInterceptor,
			grpc_zap.StreamServerInterceptor(logger),
			grpc_recovery.StreamServerInterceptor(),
		)),
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
			grpc_ctxtags.UnaryServerInterceptor(),
			grpc_opentracing.UnaryServerInterceptor(),
			grpc_prometheus.UnaryServerInterceptor,
			grpc_zap.UnaryServerInterceptor(logger),
			grpc_recovery.UnaryServerInterceptor(),
		)),
	)
	csi.RegisterIdentityServer(grpcServer, newIdentityServer())
	csi.RegisterControllerServer(grpcServer, newControllerServer())
	csi.RegisterNodeServer(grpcServer, newNodeServer())
	return grpcServer
}

func healthz(w http.ResponseWriter, r *http.Request) {
	handleOK(w, r)
}

func livez(w http.ResponseWriter, r *http.Request) {
	handleOK(w, r)
}

func readyz(w http.ResponseWriter, r *http.Request) {
	handleOK(w, r)
}

func instrument(path string, handler http.HandlerFunc) http.Handler {
	return ochttp.WithRouteTag(http.HandlerFunc(handler), path)
}

func setHeaders(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
}

func handleOK(w http.ResponseWriter, r *http.Request) {
	setHeaders(w)
	w.WriteHeader(http.StatusOK)
}
