// Package client defines a structure that sends data to the server.
package client

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"syscall"
	"time"

	"github.com/DarkOmap/metricsService/internal/certmanager"
	"github.com/DarkOmap/metricsService/internal/compresses"
	"github.com/DarkOmap/metricsService/internal/hasher"
	"github.com/DarkOmap/metricsService/internal/ip"
	"github.com/DarkOmap/metricsService/internal/logger"
	"github.com/DarkOmap/metricsService/internal/models"
	"github.com/DarkOmap/metricsService/internal/parameters"
	"github.com/go-resty/resty/v2"
)

type Compresser interface {
	GetCompressedJSON(m any) ([]byte, error)
	Close()
}

type Encrypter interface {
	EncryptMessage(m []byte) ([]byte, error)
}

// ClientHTTP it's structure witch send hashed data to server.
type ClientHTTP struct {
	restyClient *resty.Client
	gp          Compresser
	encrypter   Encrypter
	h           *hasher.Hasher
	addr        string
}

func NewClientHTTP(p parameters.AgentParameters) (*ClientHTTP, error) {
	logger.Log.Info("Create encrypt manager")
	em, err := certmanager.NewEncryptManager(p.CryptoKeyPath)

	if err != nil {
		return nil, fmt.Errorf("create encrypt manager: %w", err)
	}

	logger.Log.Info("Create hasher pool")
	h := hasher.NewHasher([]byte(p.HashKey), p.RateLimit)

	logger.Log.Info("Create gzip pool")
	pool := compresses.NewGzipPool(p.RateLimit)

	c := &ClientHTTP{
		gp:        pool,
		encrypter: em,
		h:         h,
		addr:      p.ListenAddr,
	}

	c.setRestyClient()

	return c, nil
}

func (c *ClientHTTP) Close() {
	c.gp.Close()
	c.h.Close()
}

// SendGauge send float64 value to server.
func (c *ClientHTTP) SendGauge(ctx context.Context, name string, value float64) error {
	m := models.NewMetricsForGauge(name, value)

	b, err := c.gp.GetCompressedJSON(m)

	if err != nil {
		return fmt.Errorf("failed compress model name %s value %f: %w", name, value, err)
	}

	req := c.restyClient.R().SetBody(b).
		SetHeader("Content-Type", "application/json").
		SetHeader("Content-Encoding", "gzip").
		SetContext(ctx)

	resp, err := req.Post("http://" + c.addr + "/update")

	if err != nil {
		return fmt.Errorf("send gauge: %w", err)
	}

	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("status not 200, current status %d", resp.StatusCode())
	}

	return nil
}

// SendCounter send int64 value to server.
func (c *ClientHTTP) SendCounter(ctx context.Context, name string, delta int64) error {
	m := models.NewMetricsForCounter(name, delta)

	b, err := c.gp.GetCompressedJSON(m)

	if err != nil {
		return fmt.Errorf("failed compress model name %s delta %d: %w", name, delta, err)
	}

	req := c.restyClient.R().SetBody(b).
		SetHeader("Content-Type", "application/json").
		SetHeader("Content-Encoding", "gzip").
		SetContext(ctx)

	resp, err := req.Post("http://" + c.addr + "/update")

	if err != nil {
		return fmt.Errorf("send counter name %s delta %d: %w", name, delta, err)
	}

	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("status not 200, current status %d", resp.StatusCode())
	}

	return nil
}

// SendBatch send batch data to server.
func (c *ClientHTTP) SendBatch(ctx context.Context, batch map[string]float64) error {
	m := models.GetGaugesSliceByMap(batch)

	b, err := c.gp.GetCompressedJSON(m)

	if err != nil {
		return fmt.Errorf("failed compress batch: %w", err)
	}

	req := c.restyClient.R().SetBody(b).
		SetHeader("Content-Type", "application/json").
		SetHeader("Content-Encoding", "gzip").
		SetContext(ctx)

	resp, err := req.Post("http://" + c.addr + "/updates")

	if err != nil {
		return fmt.Errorf("send batch: %w", err)
	}

	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("status not 200, current status %d", resp.StatusCode())
	}

	return nil
}

func (c *ClientHTTP) setRestyClient() {
	client := resty.New().
		AddRetryCondition(func(r *resty.Response, err error) bool {
			return errors.Is(err, syscall.ECONNREFUSED)
		}).
		SetRetryCount(3).
		SetRetryWaitTime(1 * time.Second).
		SetRetryMaxWaitTime(9 * time.Second).
		OnBeforeRequest(func(rc *resty.Client, r *resty.Request) error {
			b, err := c.encrypter.EncryptMessage(r.Body.([]byte))

			if err != nil {
				return fmt.Errorf("encrypt message: %w", err)
			}

			r.Body = b

			err = c.h.HashingRequest(r, b)

			if err != nil {
				return fmt.Errorf("hashing request: %w", err)
			}

			r.SetHeader("X-Real-IP", ip.GetLocalIP())

			return nil
		})

	c.restyClient = client
}
