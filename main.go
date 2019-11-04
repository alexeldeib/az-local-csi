package main

import (
	// DEBUG/PROFILING ONLY

	"log"
	"math/rand"
	"os"
	"os/signal"
	"regexp"
	"syscall"
	"time"

	"go.uber.org/zap"
)

var (
	csiSocket            = "/csi/csi.sock"
	prometheusPort       = ":9090"
	agentEndpointURI     = "localhost:6831"
	collectorEndpointURI = "http://localhost:14268/api/traces"
	zpageAddress         = "127.0.0.1:8081"
	serverAddress        = "0.0.0.0:8080"
	pattern              = regexp.MustCompile(`\/dev\/nvme[0-9]+n[0-9]+`)
)

func main() {
	// Seed prng
	rand.Seed(time.Now().UTC().UnixNano())

	// Setup logger
	logger, err := zap.NewDevelopment() // or NewProduction
	if err != nil {
		log.Fatalf(err.Error())
	}
	defer logger.Sync()
	sugar := logger.Sugar()

	// Enable metrics and traces via prometheus/jaeger
	enableObservabilityAndExporters(sugar)

	// Signal handler
	var interrupt = make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(interrupt)

	httpServer := newHTTPServer()
	grpcServer := newGRPCServer(logger)
	server := newServer(httpServer, grpcServer)

	// block until SIGINT/SIGTERM or an error...
	sugar.Info("starting server")
	if err := server.Start(interrupt); err != nil {
		sugar.Fatal(err)
	}
}
