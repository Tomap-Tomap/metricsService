package client

import (
	"context"
	"fmt"
	"net/http"

	memstats "github.com/DarkOmap/metricsService/internal/memstats"
	"github.com/go-resty/resty/v2"
)

func SendGauge(ctx context.Context, addr, name, value string) error {
	param := map[string]string{"name": name, "value": value}

	client := resty.New()
	resp, err := client.R().SetPathParams(param).
		SetHeader("Content-Type", "text/plain").
		SetContext(ctx).
		Post("http://" + addr + "/update/gauge/{name}/{value}")

	if err != nil {
		return fmt.Errorf("send error gauge name %s value %s: %w", name, value, err)
	}

	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("status not 200, current status %d", resp.StatusCode())
	}

	return nil
}

func SendCounter(ctx context.Context, addr, name, value string) error {
	param := map[string]string{"name": name, "value": value}

	client := resty.New()

	resp, err := client.R().SetPathParams(param).
		SetHeader("Content-Type", "text/plain").
		SetContext(ctx).
		Post("http://" + addr + "/update/counter/{name}/{value}")

	if err != nil {
		return fmt.Errorf("send error counter name %s value %s: %w", name, value, err)
	}

	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("status not 200, current status %d", resp.StatusCode())
	}

	return nil
}

func PushStats(ctx context.Context, addr string, ms []memstats.StringMS) error {
	for idx, val := range ms {
		err := SendGauge(ctx, addr, val.Name, val.Value)

		if err != nil {
			return fmt.Errorf("push error memstats index %d: %w", idx, err)
		}
	}

	return nil
}
