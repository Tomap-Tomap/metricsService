package client

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"syscall"
	"time"

	"github.com/DarkOmap/metricsService/internal/compresses"
	"github.com/DarkOmap/metricsService/internal/models"
	"github.com/go-resty/resty/v2"
)

type Client struct {
	addr        string
	restyClient *resty.Client
}

func (c *Client) SendGauge(ctx context.Context, name string, value float64) error {
	m := models.NewMetricsForGauge(name, value)

	b, err := compresses.GetCompressedJSON(m)

	if err != nil {
		return fmt.Errorf("failed compress model name %s value %f: %w", name, value, err)
	}

	req := c.restyClient.R().SetBody(b).
		SetHeader("Content-Type", "application/json").
		SetHeader("Content-Encoding", "gzip").
		SetContext(ctx)
	resp, err := req.Post("http://" + c.addr + "/update")

	if errors.Is(err, syscall.ECONNREFUSED) {
		resp, err = c.doRetryRequest(req)
	}

	if err != nil {
		return fmt.Errorf("send gauge name %s value %f: %w", name, value, err)
	}

	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("status not 200, current status %d", resp.StatusCode())
	}

	return nil
}

func (c *Client) SendCounter(ctx context.Context, name string, delta int64) error {
	m := models.NewMetricsForCounter(name, delta)

	b, err := compresses.GetCompressedJSON(m)

	if err != nil {
		return fmt.Errorf("failed compress model name %s delta %d: %w", name, delta, err)
	}

	req := c.restyClient.R().SetBody(b).
		SetHeader("Content-Type", "application/json").
		SetHeader("Content-Encoding", "gzip").
		SetContext(ctx)
	resp, err := req.Post("http://" + c.addr + "/update")

	if errors.Is(err, syscall.ECONNREFUSED) {
		resp, err = c.doRetryRequest(req)
	}

	if err != nil {
		return fmt.Errorf("send counter name %s delta %d: %w", name, delta, err)
	}

	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("status not 200, current status %d", resp.StatusCode())
	}

	return nil
}

func (c *Client) SendBatch(ctx context.Context, batch map[string]float64) error {
	m := models.GetGaugesSliceByMap(batch)

	b, err := compresses.GetCompressedJSON(m)

	if err != nil {
		return fmt.Errorf("failed compress batch: %w", err)
	}

	req := c.restyClient.R().SetBody(b).
		SetHeader("Content-Type", "application/json").
		SetHeader("Content-Encoding", "gzip").
		SetContext(ctx)

	resp, err := req.Post("http://" + c.addr + "/updates")

	if errors.Is(err, syscall.ECONNREFUSED) {
		resp, err = c.doRetryRequest(req)
	}

	if err != nil {
		return fmt.Errorf("send batch: %w", err)
	}

	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("status not 200, current status %d", resp.StatusCode())
	}

	return nil
}

func (c *Client) doRetryRequest(req *resty.Request) (*resty.Response, error) {
	sleepTime := 1
	var (
		err  error
		resp *resty.Response
	)
	for i := 0; i < 3; i++ {
		<-time.After(time.Duration(sleepTime) * time.Second)

		resp, err = req.Post("http://" + c.addr + "/updates")

		if err == nil {
			return resp, err
		}

		if !errors.Is(err, syscall.ECONNREFUSED) {
			return nil, err
		}

		sleepTime += 2
	}

	return nil, err
}

func NewClient(addr string) *Client {
	return &Client{addr, resty.New()}
}
