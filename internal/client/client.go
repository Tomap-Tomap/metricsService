// The package client defines a structure that sends data to the server.
package client

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"syscall"
	"time"

	"github.com/DarkOmap/metricsService/internal/compresses"
	"github.com/DarkOmap/metricsService/internal/hasher"
	"github.com/DarkOmap/metricsService/internal/logger"
	"github.com/DarkOmap/metricsService/internal/models"
	"github.com/go-resty/resty/v2"
	"go.uber.org/zap"
)

// Agent it's structure witch send hashed data to server.
type Client struct {
	addr        string
	restyClient *resty.Client
	hasher      hasher.Hasher
	jobs        chan func() error
	gp          *compresses.GzipPool
}

func NewClient(addr, key string, rateLimit uint) *Client {
	client := resty.New().
		AddRetryCondition(func(r *resty.Response, err error) bool {
			return errors.Is(err, syscall.ECONNREFUSED)
		}).
		SetRetryCount(3).
		SetRetryWaitTime(1 * time.Second).
		SetRetryMaxWaitTime(9 * time.Second)

	jobs := make(chan func() error, rateLimit)

	c := &Client{
		addr,
		client,
		hasher.NewHasher([]byte(key)),
		jobs,
		compresses.NewGzipPool(rateLimit),
	}

	for w := 1; w <= cap(jobs); w++ {
		go c.worker(jobs)
	}

	return c
}

// SendGauge send float64 value to server.
func (c *Client) SendGauge(ctx context.Context, name string, value float64) {
	c.jobs <- func() error {
		return c.sendGauge(ctx, name, value)
	}
}

// SendCounter send int64 value to server.
func (c *Client) SendCounter(ctx context.Context, name string, delta int64) {
	c.jobs <- func() error {
		return c.sendCounter(ctx, name, delta)
	}
}

// SendBatch send batch data to server.
func (c *Client) SendBatch(ctx context.Context, batch map[string]float64) {
	c.jobs <- func() error {
		return c.sendBatch(ctx, batch)
	}
}

func (c *Client) Close() {
	close(c.jobs)
}

func (c *Client) sendGauge(ctx context.Context, name string, value float64) error {
	m := models.NewMetricsForGauge(name, value)

	b, err := c.gp.GetCompressedJSON(m)

	if err != nil {
		return fmt.Errorf("failed compress model name %s value %f: %w", name, value, err)
	}

	req := c.restyClient.R().SetBody(b).
		SetHeader("Content-Type", "application/json").
		SetHeader("Content-Encoding", "gzip").
		SetContext(ctx)
	c.hasher.HashingRequest(req, b)
	resp, err := req.Post("http://" + c.addr + "/update")

	if err != nil {
		return fmt.Errorf("send gauge name %s value %f: %w", name, value, err)
	}

	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("status not 200, current status %d", resp.StatusCode())
	}

	return nil
}

func (c *Client) sendCounter(ctx context.Context, name string, delta int64) error {
	m := models.NewMetricsForCounter(name, delta)

	b, err := c.gp.GetCompressedJSON(m)

	if err != nil {
		return fmt.Errorf("failed compress model name %s delta %d: %w", name, delta, err)
	}

	req := c.restyClient.R().SetBody(b).
		SetHeader("Content-Type", "application/json").
		SetHeader("Content-Encoding", "gzip").
		SetContext(ctx)
	c.hasher.HashingRequest(req, b)
	resp, err := req.Post("http://" + c.addr + "/update")

	if err != nil {
		return fmt.Errorf("send counter name %s delta %d: %w", name, delta, err)
	}

	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("status not 200, current status %d", resp.StatusCode())
	}

	return nil
}

func (c *Client) sendBatch(ctx context.Context, batch map[string]float64) error {
	m := models.GetGaugesSliceByMap(batch)

	b, err := c.gp.GetCompressedJSON(m)

	if err != nil {
		return fmt.Errorf("failed compress batch: %w", err)
	}

	req := c.restyClient.R().SetBody(b).
		SetHeader("Content-Type", "application/json").
		SetHeader("Content-Encoding", "gzip").
		SetContext(ctx)
	c.hasher.HashingRequest(req, b)
	resp, err := req.Post("http://" + c.addr + "/updates")

	if err != nil {
		return fmt.Errorf("send batch: %w", err)
	}

	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("status not 200, current status %d", resp.StatusCode())
	}

	return nil
}

func (c *Client) worker(jobs <-chan func() error) {
	for j := range jobs {
		err := j()

		if err != nil {
			logger.Log.Warn(
				"Error on sending to server",
				zap.Error(err),
			)
		}
	}
}
