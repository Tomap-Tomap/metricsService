// Package server contains structures and methods for running the server
package server

import (
	"context"
	"fmt"
	"net"
	"net/http"

	"github.com/DarkOmap/metricsService/handlers"
	"github.com/DarkOmap/metricsService/internal/certmanager"
	"github.com/DarkOmap/metricsService/internal/compresses"
	"github.com/DarkOmap/metricsService/internal/hasher"
	"github.com/DarkOmap/metricsService/internal/ip"
	"github.com/DarkOmap/metricsService/internal/logger"
	"github.com/DarkOmap/metricsService/internal/parameters"
	"github.com/DarkOmap/metricsService/internal/proto"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
)

// Server this is a metrics storage server
type Server struct {
	httpServer *http.Server
	Listener   net.Listener
	grpsServer *grpc.Server
}

// OptionFunc this is a function for configuring the server
type OptionFunc func(*Server) error

// NewServer create Server
func NewServer(opts ...OptionFunc) (*Server, error) {
	s := &Server{}

	for _, opt := range opts {
		err := opt(s)
		if err != nil {
			return nil, err
		}
	}
	return s, nil
}

// Run runs server
func (s *Server) Run(ctx context.Context) error {
	eg, egCtx := errgroup.WithContext(ctx)

	if s.httpServer != nil {
		s.runHTTPServer(egCtx, eg)
	}

	if s.grpsServer != nil {
		s.runGRPCServer(egCtx, eg)
	}

	if err := eg.Wait(); err != nil {
		return fmt.Errorf("unexpected server shutdown: %w", err)
	}

	return nil
}

func (s *Server) runHTTPServer(ctx context.Context, eg *errgroup.Group) {
	eg.Go(func() error {
		logger.Log.Info("Run serve")
		err := s.httpServer.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			return err
		}
		return nil
	})

	eg.Go(func() error {
		<-ctx.Done()
		logger.Log.Info("Stop serve")
		return s.httpServer.Shutdown(context.Background())
	})
}

func (s *Server) runGRPCServer(ctx context.Context, eg *errgroup.Group) {
	eg.Go(func() error {
		logger.Log.Info("Run grpc serve")
		err := s.grpsServer.Serve(s.Listener)
		if err != nil && err != http.ErrServerClosed {
			return err
		}
		return nil
	})

	eg.Go(func() error {
		<-ctx.Done()
		logger.Log.Info("Stop grpc serve")
		s.grpsServer.GracefulStop()
		return nil
	})
}

// WithHTTP returns a functional option that adds http handlers to the server
func WithHTTP(r handlers.Repository, ipc *ip.Checker, h *hasher.Hasher, gp *compresses.GzipPool, p parameters.ServerParameters) OptionFunc {
	return func(s *Server) error {
		logger.Log.Info("Create decrypt manager")

		dm, err := certmanager.NewDecryptManager(p.CryptoKeyPath)
		if err != nil {
			return fmt.Errorf("create descrypt manager: %w", err)
		}

		logger.Log.Info("Create handlers")
		sh := handlers.NewServiceHandlers(r)

		logger.Log.Info("Create routers")
		router := handlers.ServiceRouter(gp, h, sh, dm, ipc)

		logger.Log.Info("Create server")
		s.httpServer = &http.Server{
			Addr:    p.FlagRunAddr,
			Handler: router,
		}

		return nil
	}
}

// WithGRPC returns a functional option that adds grpc handlers to the server
func WithGRPC(r handlers.Repository, ipc *ip.Checker, h *hasher.Hasher, p parameters.ServerParameters) OptionFunc {
	return func(s *Server) error {
		logger.Log.Info("Create grpc server")

		listen, err := net.Listen("tcp", p.FlagRunGRPCAddr)
		if err != nil {
			return fmt.Errorf("create listener: %w", err)
		}

		gs := grpc.NewServer(grpc.ChainUnaryInterceptor(
			logger.InterceptorLogger,
			ipc.InterceptorIPCheck,
			h.InterceptorCheckHash,
		))

		proto.RegisterMetricsServer(gs, handlers.NewMetricsServer(r))

		s.Listener = listen
		s.grpsServer = gs

		return nil
	}
}
