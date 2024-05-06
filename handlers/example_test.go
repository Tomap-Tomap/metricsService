package handlers

import (
	"github.com/go-resty/resty/v2"
)

func ExampleServiceHandlers_updateByJSON() {
	client := resty.New()
	req := client.R().SetBody(`{
		"id": "test",
		"type": "gauge",
		"value": 1.1
	}`).SetHeader("Content-Type", "application/json")
	req.Post("localhost:8080/update")

	// Update by URL
	req = client.R().SetHeader("Content-Type", "text/plain")
	req.Post("localhost:8080/update/gauge/test/1.1")
}

func ExampleServiceHandlers_updateByURL() {
	client := resty.New()

	req := client.R().SetHeader("Content-Type", "text/plain")
	req.Post("localhost:8080/update/gauge/test/1.1")
}

func ExampleServiceHandlers_valueByJSON() {
	client := resty.New()
	// Value by JSON
	req := client.R().SetBody(`{
		"id": "test",
		"type": "gauge"
	}`).SetHeader("Content-Type", "application/json")
	req.Post("localhost:8080/value")
}

func ExampleServiceHandlers_valueByURL() {
	client := resty.New()

	// Value by URL
	req := client.R().SetHeader("Content-Type", "text/plain")
	req.Get("localhost:8080/value/gauge/test")
}

func ExampleServiceHandlers_ping() {
	client := resty.New()

	req := client.R().SetHeader("Content-Type", "text/plain")
	req.Get("localhost:8080/ping")
}

func ExampleServiceHandlers_updates() {
	client := resty.New()
	// Update by JSON
	req := client.R().SetBody(`[
		{
			"id": "test",
			"type": "gauge",
			"value": 1.1
		},
		{
			"id": "test2",
			"type": "gauge",
			"value": 1.2
		}
	]`).SetHeader("Content-Type", "application/json")
	req.Post("localhost:8080/updates")
}

func ExampleServiceHandlers_all() {
	client := resty.New()

	req := client.R().SetHeader("Content-Type", "text/plain")
	req.Get("localhost:8080")
}
