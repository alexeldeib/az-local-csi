package main

import (
	"net/http"

	"contrib.go.opencensus.io/exporter/jaeger"
	"contrib.go.opencensus.io/exporter/prometheus"
	"github.com/pkg/errors"
	"go.opencensus.io/plugin/ochttp"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/trace"
	"go.opencensus.io/zpages"
	"go.uber.org/zap"
)

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
