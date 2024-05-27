// Package client defines a structure that sends data to the server.
package client

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"syscall"
	"time"

	"github.com/DarkOmap/metricsService/internal/hasher"
	"github.com/DarkOmap/metricsService/internal/models"
	"github.com/go-resty/resty/v2"
)

type Compresser interface {
	GetCompressedJSON(m any) ([]byte, error)
}

type Encrypter interface {
	EncryptMessage(m []byte) ([]byte, error)
}

// Client it's structure witch send hashed data to server.
type Client struct {
	restyClient *resty.Client
	gp          Compresser
	addr        string
}

func NewClient(compresser Compresser, encrypter Encrypter, h hasher.Hasher, addr string) *Client {
	client := resty.New().
		AddRetryCondition(func(r *resty.Response, err error) bool {
			return errors.Is(err, syscall.ECONNREFUSED)
		}).
		SetRetryCount(3).
		SetRetryWaitTime(1 * time.Second).
		SetRetryMaxWaitTime(9 * time.Second).
		OnBeforeRequest(func(c *resty.Client, r *resty.Request) error {
			b, err := encrypter.EncryptMessage(r.Body.([]byte))

			if err != nil {
				return fmt.Errorf("encrypt message: %w", err)
			}

			r.Body = b

			err = h.HashingRequest(r, b)

			if err != nil {
				return fmt.Errorf("hashing request: %w", err)
			}

			return nil
		})

	c := &Client{
		client,
		compresser,
		addr,
	}

	return c
}

// SendGauge send float64 value to server.
func (c *Client) SendGauge(ctx context.Context, name string, value float64) error {
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
		return fmt.Errorf("send gauge name %s value %f: %w", name, value, err)
	}

	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("status not 200, current status %d", resp.StatusCode())
	}

	return nil
}

// SendCounter send int64 value to server.
func (c *Client) SendCounter(ctx context.Context, name string, delta int64) error {
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
func (c *Client) SendBatch(ctx context.Context, batch map[string]float64) error {
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
