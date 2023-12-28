package client

import (
	"errors"
	"net/http"

	memstats "github.com/DarkOmap/metricsService/internal/memStats"
	"github.com/go-resty/resty/v2"
)

var ServiceAddr string

func init() {
	ServiceAddr = "http://localhost:8080/update"
}

func SendGauge(name, value string) error {
	param := map[string]string{"name": name, "value": value}

	client := resty.New()

	resp, err := client.R().SetPathParams(param).
		SetHeader("Content-Type", "text/plain").
		Post(ServiceAddr + "/gauge/{name}/{value}")

	if err != nil {
		return err
	}

	if resp.StatusCode() != http.StatusOK {
		return errors.New("status not OK")
	}

	return nil
}

func SendCounter(name, value string) error {
	param := map[string]string{"name": name, "value": value}

	client := resty.New()

	resp, err := client.R().SetPathParams(param).
		SetHeader("Content-Type", "text/plain").
		Post(ServiceAddr + "/counter/{name}/{value}")

	if err != nil {
		return err
	}

	if resp.StatusCode() != http.StatusOK {
		return errors.New("status not OK")
	}

	return nil
}

func PushStats(ms []memstats.StringMS) error {
	for _, val := range ms {
		err := SendGauge(val.Name, val.Value)

		if err != nil {
			return err
		}
	}

	return nil
}
