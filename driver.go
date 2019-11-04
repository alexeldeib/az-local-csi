package main

import (
	"context"
	"net"
	"net/http"
	"os"
	"time"

	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
)

type driver struct {
	httpServer *http.Server
	grpcServer *grpc.Server
	log        *zap.SugaredLogger
	errgroup   *errgroup.Group
	interrupt  chan os.Signal
}

func newDriver(log *zap.SugaredLogger) *driver {
	return &driver{
		httpServer: newHTTPServer(),
		grpcServer: newGRPCServer(),
		log:        log,
		interrupt:  make(chan os.Signal, 1),
	}
}

func (d *driver) start(ctx context.Context) context.Context {
	// errgroup for grpc, http, and health servers
	g, ctx := errgroup.WithContext(ctx)
	d.errgroup = g
	d.log.Info("starting http server")
	g.Go(func() error {
		if err := d.httpServer.ListenAndServe(); err != http.ErrServerClosed {
			d.log.Error(err)
			return err
		}
		return nil
	})

	d.log.Info("starting grpc server")
	g.Go(func() error {
		listener, err := net.Listen("unix", csiSocket)
		if err != nil {
			d.log.Error(err)
			return err
		}
		return d.grpcServer.Serve(listener)
	})
	return ctx
}

func (d *driver) block(ctx context.Context, cancel context.CancelFunc) {
	select {
	case <-d.interrupt:
		break
	case <-ctx.Done():
		break
	}
	d.log.Info("received shutdown signal")
	cancel()
}

func (d *driver) stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if d.httpServer != nil {
		_ = d.httpServer.Shutdown(ctx)
	}

	if d.grpcServer != nil {
		stopped := make(chan struct{})
		go func() {
			d.grpcServer.GracefulStop()
			close(stopped)
		}()

		select {
		case <-ctx.Done():
			d.grpcServer.Stop()
		case <-stopped:
		}
	}

	return d.errgroup.Wait()
}
