package main

import (
	// DEBUG/PROFILING ONLY
	// _ "net/http/pprof"

	"context"
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

	// main context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Setup logger
	logger, err := zap.NewDevelopment() // or NewProduction
	if err != nil {
		log.Fatalf(err.Error())
	}
	defer logger.Sync()
	sugar := logger.Sugar()

	// Initialize plugin
	d := newDriver(sugar)

	// Signal handler
	signal.Notify(d.interrupt, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(d.interrupt)

	// Enable metrics and traces via prometheus/jaeger
	enableObservabilityAndExporters(sugar)

	sugar.Info(pattern.Match([]byte("/dev/nvme0n2")))
	sugar.Info(pattern.Match([]byte("/dev/nvme0n112asf")))
	sugar.Info(pattern.Match([]byte("/dev/nvme0112asf")))
	sugar.Info(pattern.Match([]byte("/dev/nvme0n")))

	// start...
	shutdownCtx := d.start(ctx)

	// block until SIGINT/SIGTERM or an error...
	d.block(shutdownCtx, cancel)

	// ...and terminate
	if err := d.stop(); err != nil {
		sugar.Fatal(err)
	}
}
