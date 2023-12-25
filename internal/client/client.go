package client

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"

	memstats "github.com/DarkOmap/metricsService/internal/memStats"
)

const serviceAddr = "http://localhost:8080/update"

func SendGauge(name, value string) error {
	respString := fmt.Sprintf("%s/gauge/%s/%s", serviceAddr, name, value)
	buf := bytes.NewReader([]byte{})
	resp, err := http.Post(respString, "text/plain", buf)

	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return errors.New("status not OK")
	}

	return nil
}

func SendCounter(name, value string) error {
	respString := fmt.Sprintf("%s/counter/%s/%s", serviceAddr, name, value)
	buf := bytes.NewReader([]byte{})
	resp, err := http.Post(respString, "text/plain", buf)

	if err != nil {
		return err
	}

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
