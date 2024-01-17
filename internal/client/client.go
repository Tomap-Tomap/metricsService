package client

import (
	"context"
	"fmt"
	"net/http"

	memstats "github.com/DarkOmap/metricsService/internal/memstats"
	"github.com/go-resty/resty/v2"
)

type Client struct {
	addr        string
	restyClient *resty.Client
}

func (c Client) SendGauge(ctx context.Context, name, value string) error {
	param := map[string]string{"name": name, "value": value}

	resp, err := c.restyClient.R().SetPathParams(param).
		SetHeader("Content-Type", "text/plain").
		SetContext(ctx).
		Post("http://" + c.addr + "/update/gauge/{name}/{value}")

	if err != nil {
		return fmt.Errorf("send gauge name %s value %s: %w", name, value, err)
	}

	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("status not 200, current status %d", resp.StatusCode())
	}

	return nil
}

func (c Client) SendCounter(ctx context.Context, name, value string) error {
	param := map[string]string{"name": name, "value": value}

	resp, err := c.restyClient.R().SetPathParams(param).
		SetHeader("Content-Type", "text/plain").
		SetContext(ctx).
		Post("http://" + c.addr + "/update/counter/{name}/{value}")

	if err != nil {
		return fmt.Errorf("send counter name %s value %s: %w", name, value, err)
	}

	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("status not 200, current status %d", resp.StatusCode())
	}

	return nil
}

func (c Client) PushStats(ctx context.Context, ms []memstats.StringMS) error {
	for idx, val := range ms {
		err := c.SendGauge(ctx, val.Name, val.Value)

		if err != nil {
			return fmt.Errorf("push memstats index %d: %w", idx, err)
		}
	}

	return nil
}

func NewClient(addr string) Client {
	return Client{addr, resty.New()}
}
