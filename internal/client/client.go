package client

import (
	"context"
	"fmt"
	"net/http"

	"github.com/DarkOmap/metricsService/internal/compresses"
	"github.com/DarkOmap/metricsService/internal/models"
	"github.com/go-resty/resty/v2"
)

type Client struct {
	addr        string
	restyClient *resty.Client
}

func (c Client) SendGauge(ctx context.Context, name string, value float64) error {
	m := models.NewMetricsForGauge(name, value)

	b, err := compresses.GetCompressJSON(m)

	if err != nil {
		return fmt.Errorf("failed compress model name %s value %f: %w", name, value, err)
	}

	resp, err := c.restyClient.R().SetBody(b).
		SetHeader("Content-Type", "application/json").
		SetHeader("Content-Encoding", "gzip").
		SetContext(ctx).
		Post("http://" + c.addr + "/update")

	if err != nil {
		return fmt.Errorf("send gauge name %s value %f: %w", name, value, err)
	}

	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("status not 200, current status %d", resp.StatusCode())
	}

	return nil
}

func (c Client) SendCounter(ctx context.Context, name string, delta int64) error {
	m := models.NewMetricsForCounter(name, delta)

	b, err := compresses.GetCompressJSON(m)

	if err != nil {
		return fmt.Errorf("failed compress model name %s delta %d: %w", name, delta, err)
	}

	resp, err := c.restyClient.R().SetBody(b).
		SetHeader("Content-Type", "application/json").
		SetHeader("Content-Encoding", "gzip").
		SetContext(ctx).
		Post("http://" + c.addr + "/update")

	if err != nil {
		return fmt.Errorf("send counter name %s delta %d: %w", name, delta, err)
	}

	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("status not 200, current status %d", resp.StatusCode())
	}

	return nil
}

func NewClient(addr string) Client {
	return Client{addr, resty.New()}
}
