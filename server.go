package main

import (
	"context"
	"net"
	"net/http"
	"net/http/pprof"
	"os"
	"sync"
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
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

type Server struct {
	http *HTTPServer
	grpc *GRPCServer
	errs chan error
}

func newServer(http *HTTPServer, grpc *GRPCServer) *Server {
	return &Server{
		http: http,
		grpc: grpc,
		errs: make(chan error),
	}
}

func (s *Server) Start(stop chan os.Signal) error {
	go s.http.Start(s.errs)
	go s.grpc.Start(s.errs)

	select {
	case <-stop:
		return nil
	case err := <-s.errs:
		return err
	}
}

func (s *Server) Shutdown() {
	var wg = &sync.WaitGroup{}
	wg.Add(2)
	go func() {
		s.http.Shutdown()
		wg.Done()
	}()
	go func() {
		s.grpc.Shutdown()
		wg.Done()
	}()
	wg.Wait()
}

type HTTPServer struct {
	srv *http.Server
}

func (h *HTTPServer) Start(errs chan<- error) {
	if err := h.srv.ListenAndServe(); err != http.ErrServerClosed {
		errs <- err
	}
}

func (h *HTTPServer) Shutdown() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if h.srv != nil {
		_ = h.srv.Shutdown(ctx)
	}
}

func newHTTPServer() *HTTPServer {
	router := http.NewServeMux()
	router.Handle("/healthz", instrument("/healthz", healthz))
	router.Handle("/livez", instrument("/livez", livez))
	router.Handle("/readyz", instrument("/readyz", readyz))

	router.Handle("/debug/pprof/", http.HandlerFunc(pprof.Index))
	router.Handle("/debug/pprof/cmdline", http.HandlerFunc(pprof.Cmdline))
	router.Handle("/debug/pprof/profile", http.HandlerFunc(pprof.Profile))
	router.Handle("/debug/pprof/symbol", http.HandlerFunc(pprof.Symbol))
	router.Handle("/debug/pprof/trace", http.HandlerFunc(pprof.Trace))

	return &HTTPServer{
		srv: &http.Server{
			Addr:         serverAddress,
			ReadTimeout:  30 * time.Second,
			WriteTimeout: 30 * time.Second,
			Handler: &ochttp.Handler{
				Handler: router,
			},
		},
	}
}

type GRPCServer struct {
	srv *grpc.Server
}

func (g *GRPCServer) Start(errs chan<- error) {
	listener, err := net.Listen("unix", csiSocket)
	if err != nil {
		errs <- err
		return
	}
	if err := g.srv.Serve(listener); err != nil {
		errs <- err
	}
}

func (g *GRPCServer) Shutdown() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if g.srv != nil {
		stopped := make(chan struct{})
		go func() {
			g.srv.GracefulStop()
			close(stopped)
		}()

		select {
		case <-ctx.Done():
			g.srv.Stop()
		case <-stopped:
		}
	}
}

func newGRPCServer(logger *zap.Logger) *GRPCServer {
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
	return &GRPCServer{
		srv: grpcServer,
	}
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
