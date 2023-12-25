package client

import (
	"errors"
	"fmt"
	"net/http"

	memstats "github.com/DarkOmap/metricsService/internal/memStats"
)

var ServiceAddr string

func init() {
	ServiceAddr = "http://localhost:8080/update"
}

func SendGauge(name, value string) error {
	respString := fmt.Sprintf("%s/gauge/%s/%s", ServiceAddr, name, value)

	resp, err := http.Post(respString, "text/plain", nil)

	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.New("status not OK")
	}

	return nil
}

func SendCounter(name, value string) error {
	respString := fmt.Sprintf("%s/counter/%s/%s", ServiceAddr, name, value)

	resp, err := http.Post(respString, "text/plain", nil)

	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
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
