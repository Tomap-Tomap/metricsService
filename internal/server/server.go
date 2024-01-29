package server

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/DarkOmap/metricsService/internal/file"
	"github.com/DarkOmap/metricsService/internal/logger"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

type fileRepository interface {
	EnableWriteEvent()
	GetWriteEventFunc() (func() <-chan struct{}, error)
}

type Server struct {
	fileStoragePath string
	fr              fileRepository
	storeFunc       func() <-chan struct{}
}

func NewServer(fr fileRepository, fileStoragePath string, storeInterval uint, restoreFromFile bool) (*Server, error) {
	srv := &Server{
		fileStoragePath: fileStoragePath,
		fr:              fr,
	}

	if err := srv.initializeStoreFunc(storeInterval); err != nil {
		return nil, fmt.Errorf("initialize store func: %w", err)
	}

	if restoreFromFile {
		if err := srv.restoreFromFile(); err != nil {
			return nil, fmt.Errorf("restore from file: %w", err)
		}
	}

	return srv, nil
}

func (s *Server) initializeStoreFunc(storeInterval uint) error {
	if storeInterval == 0 {
		s.fr.EnableWriteEvent()
		sf, err := s.fr.GetWriteEventFunc()

		if err != nil {
			return fmt.Errorf("setup of save to a file: %w", err)
		}

		s.storeFunc = sf
	} else {
		s.storeFunc = func() <-chan struct{} {
			ch := make(chan struct{})
			go func() {
				<-time.After(time.Duration(storeInterval) * time.Second)
				ch <- struct{}{}
			}()

			return ch
		}
	}

	return nil
}

func (s *Server) restoreFromFile() error {
	consumer, err := file.NewConsumer(s.fileStoragePath)

	if err != nil {
		return fmt.Errorf("initializing new consumer: %w", err)
	}
	defer consumer.Close()

	if err := consumer.Decoder.Decode(s.fr); err != nil && err != io.EOF {
		return fmt.Errorf("read from file for storage: %w", err)
	}

	return nil
}

func (s *Server) ListenAndServe(addr string, handler http.Handler) error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer cancel()

	eg, egCtx := errgroup.WithContext(ctx)

	syncCtx := s.runSyncFromFile(egCtx, eg)

	httpServer := &http.Server{
		Addr:    addr,
		Handler: handler,
	}

	eg.Go(func() error {
		return httpServer.ListenAndServe()
	})

	eg.Go(func() error {
		<-syncCtx.Done()
		logger.Log.Info("stop serve")
		return httpServer.Shutdown(context.Background())
	})

	if err := eg.Wait(); err != nil && err != http.ErrServerClosed {
		logger.Log.Error("problem with working server", zap.Error(err))
		return fmt.Errorf("problem with working server: %w", err)
	}

	return nil
}

func (s *Server) runSyncFromFile(ctx context.Context, eg *errgroup.Group) context.Context {
	syncCtx, cancel := context.WithCancel(context.Background())
	eg.Go(func() error {
		defer cancel()
		producer, err := file.NewProducer(s.fileStoragePath)

		if err != nil {
			return err
		}

		defer producer.Close()

		for {
			select {
			case <-s.storeFunc():
				producer.Seek()
				producer.Encoder.Encode(s.fr)
			case <-ctx.Done():
				producer.Seek()
				producer.Encoder.Encode(s.fr)
				logger.Log.Info("stop sync")
				return nil
			}
		}
	})

	return syncCtx
}
