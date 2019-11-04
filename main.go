package main

import (
	// DEBUG/PROFILING ONLY
	// _ "net/http/pprof"

	"context"
	"log"
	"math/rand"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"contrib.go.opencensus.io/exporter/jaeger"
	"contrib.go.opencensus.io/exporter/prometheus"
	"github.com/container-storage-interface/spec/lib/go/csi"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	grpc_opentracing "github.com/grpc-ecosystem/go-grpc-middleware/tracing/opentracing"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/pkg/errors"
	"go.opencensus.io/plugin/ocgrpc"
	"go.opencensus.io/plugin/ochttp"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/trace"
	"go.opencensus.io/zpages"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

var (
	csiSocket            = "/csi/csi.sock"
	prometheusPort       = ":9090"
	grpcHealthPort       = ":8082"
	agentEndpointURI     = "localhost:6831"
	collectorEndpointURI = "http://localhost:14268/api/traces"
	zpageAddress         = "127.0.0.1:8081"
	serverAddress        = "0.0.0.0:8080"
	sugar                *zap.SugaredLogger
	logger               *zap.Logger
	httpServer           *http.Server
	grpcServer           *grpc.Server
	// grpcHealthServer     *grpc.Server
	// healthServer         *health.Server
)

func main() {
	rand.Seed(time.Now().UTC().UnixNano())

	// main context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Signal handler
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(interrupt)

	// errgroup for grpc, http, and health servers
	g, ctx := errgroup.WithContext(ctx)

	// Setup logger
	var err error
	logger, err = zap.NewDevelopment() // or NewProduction
	if err != nil {
		log.Fatalf(err.Error())
	}
	defer logger.Sync()
	sugar = logger.Sugar()

	// Enable metrics and traces via prometheus/jaeger
	enableObservabilityAndExporters(sugar)

	// // Setup gRPC health server
	// g.Go(func() error {
	// 	grpcHealthServer = grpc.NewServer()
	// 	healthServer = health.NewServer()
	// 	healthpb.RegisterHealthServer(grpcHealthServer, healthServer)
	// 	healthListener, err := net.Listen("tcp", grpcHealthPort)
	// 	if err != nil {
	// 		sugar.Error(err, "gRPC Health server: failed to listen")
	// 		os.Exit(2)
	// 	}
	// 	sugar.Info(fmt.Sprintf("gRPC health server serving at %s", grpcHealthPort))
	// 	return grpcHealthServer.Serve(healthListener)
	// })

	router := http.NewServeMux()
	router.Handle("/healthz", instrument("/healthz", healthz))
	router.Handle("/livez", instrument("/livez", livez))
	router.Handle("/readyz", instrument("/readyz", readyz))

	httpServer = &http.Server{
		Addr:         serverAddress,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		Handler: &ochttp.Handler{
			Handler: router,
		},
	}

	// Setup gRPC server
	grpcServer = grpc.NewServer(
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

	csi.RegisterIdentityServer(grpcServer, &identityServer{})
	csi.RegisterControllerServer(grpcServer, &controllerServer{})
	csi.RegisterNodeServer(grpcServer, &nodeServer{})

	listener, err := net.Listen("unix", csiSocket)
	if err != nil {
		sugar.Fatal(err)
	}

	sugar.Info("starting server")
	g.Go(func() error {
		if err := httpServer.ListenAndServe(); err != http.ErrServerClosed {
			return err
		}
		return nil
	})

	g.Go(func() error {
		return grpcServer.Serve(listener)
	})

	select {
	case <-interrupt:
		break
	case <-ctx.Done():
		break
	}

	sugar.Info("received shutdown signal")

	cancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if httpServer != nil {
		_ = httpServer.Shutdown(shutdownCtx)
	}

	if grpcServer != nil {
		stopped := make(chan struct{})
		go func() {
			grpcServer.GracefulStop()
			close(stopped)
		}()

		t := time.NewTimer(30 * time.Second)
		select {
		case <-t.C:
			grpcServer.Stop()
		case <-stopped:
			t.Stop()
		}
	}

	err = g.Wait()
	if err != nil {
		sugar.Fatal(err)
	}
}

func enableObservabilityAndExporters(sugar *zap.SugaredLogger) error {
	// Start z-Pages server.
	go func() {
		mux := http.NewServeMux()
		zpages.Handle(mux, "/debug")
		sugar.Fatal(http.ListenAndServe(zpageAddress, mux))
	}()

	// register opencensus http views
	if err := view.Register(ochttp.DefaultServerViews...); err != nil {
		errors.Wrapf(err, "failed to register server views for HTTP metrics: %v")
	}

	// Stats exporter: Prometheus
	pe, err := prometheus.NewExporter(prometheus.Options{
		Namespace: "appserver",
	})

	if err != nil {
		errors.Wrapf(err, "failed to create the Prometheus stats exporter: %v")
	}

	view.RegisterExporter(pe)
	go func() {
		mux := http.NewServeMux()
		mux.Handle("/metrics", pe)
		sugar.Fatal(http.ListenAndServe(prometheusPort, mux))
	}()

	// Trace exporter: Jaeger
	je, err := jaeger.NewExporter(jaeger.Options{
		AgentEndpoint:     agentEndpointURI,
		CollectorEndpoint: collectorEndpointURI,
		ServiceName:       "demo",
	})
	if err != nil {
		return errors.Wrapf(err, "failed to create the Jaeger exporter: %v")
	}

	trace.RegisterExporter(je)

	trace.ApplyConfig(trace.Config{DefaultSampler: trace.AlwaysSample()})
	return nil
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
