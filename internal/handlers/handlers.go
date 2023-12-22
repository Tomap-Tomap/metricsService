package handlers

import (
	"net/http"

	"github.com/DarkOmap/metricsService/internal/storage"
)

func Update(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(res, "only POST requests are allowed!", http.StatusMethodNotAllowed)
		return
	}

	url := req.URL.RequestURI()
	su, err := storage.NewStorageUnit(url)

	if err != nil {
		http.Error(res, err.Error(), http.StatusNotFound)
		return
	}

	err = storage.Storage.AddUnit(su)

	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
	}
}
