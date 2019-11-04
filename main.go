package main

import (
	// DEBUG/PROFILING ONLY
	// _ "net/http/pprof"

	"log"
	"math/rand"
	"net/http"
	"time"

	"contrib.go.opencensus.io/exporter/jaeger"
	"contrib.go.opencensus.io/exporter/prometheus"
	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/pkg/errors"
	"go.opencensus.io/plugin/ocgrpc"
	"go.opencensus.io/plugin/ochttp"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/trace"
	"go.opencensus.io/zpages"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

var (
	prometheusPort       = ":9090"
	agentEndpointURI     = "localhost:6831"
	collectorEndpointURI = "http://localhost:14268/api/traces"
	zpageAddress         = "127.0.0.1:8081"
	serverAddress        = "0.0.0.0:8080"
	sugar                *zap.SugaredLogger
	logger               *zap.Logger
)

func main() {
	rand.Seed(time.Now().UTC().UnixNano())

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

	router := http.NewServeMux()
	router.Handle("/healthz", instrument("/healthz", healthz))
	router.Handle("/livez", instrument("/livez", livez))
	router.Handle("/readyz", instrument("/readyz", readyz))

	srv := grpc.NewServer(grpc.StatsHandler(&ocgrpc.ServerHandler{}))
	csi.RegisterIdentityServer(srv, &identityServer{})
	csi.RegisterControllerServer(srv, &controllerServer{})
	csi.RegisterNodeServer(srv, &nodeServer{})

	sugar.Info("starting server")
	sugar.Fatal(http.ListenAndServe(serverAddress, &ochttp.Handler{
		Handler: router,
	}))
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
