package ip

import (
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/require"
)

func TestGetLocalIP(t *testing.T) {
	t.Run("check singleton", func(t *testing.T) {
		want := GetLocalIP()
		got := GetLocalIP()

		require.Equal(t, want, got)
	})
}

func TestIPChecker_RequsetIPCheck(t *testing.T) {
	t.Run("positive test", func(t *testing.T) {
		_, ts, err := net.ParseCIDR("192.168.1.0/24")

		require.NoError(t, err)

		ipc := NewIPChecker(ts)

		webhook := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		handler := ipc.RequsetIPCheck(webhook)

		srv := httptest.NewServer(handler)
		defer srv.Close()

		c := resty.New()

		req := c.R().SetHeader("X-Real-IP", "192.168.1.0")
		resp, err := req.Post(srv.URL)

		require.NoError(t, err)
		require.Equal(t, resp.StatusCode(), 200)
	})

	t.Run("test nil ipNET", func(t *testing.T) {
		ipc := NewIPChecker(nil)

		webhook := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		handler := ipc.RequsetIPCheck(webhook)

		srv := httptest.NewServer(handler)
		defer srv.Close()

		c := resty.New()

		req := c.R()
		resp, err := req.Post(srv.URL)

		require.NoError(t, err)
		require.Equal(t, resp.StatusCode(), 200)
	})

	t.Run("test empty real ip", func(t *testing.T) {
		_, ts, err := net.ParseCIDR("192.168.1.0/24")

		require.NoError(t, err)

		ipc := NewIPChecker(ts)

		webhook := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		handler := ipc.RequsetIPCheck(webhook)

		srv := httptest.NewServer(handler)
		defer srv.Close()

		c := resty.New()

		req := c.R().SetHeader("X-Real-IP", "")
		resp, err := req.Post(srv.URL)

		require.NoError(t, err)
		require.Equal(t, resp.StatusCode(), 403)
	})

	t.Run("test not contain ip", func(t *testing.T) {
		_, ts, err := net.ParseCIDR("192.168.1.0/24")

		require.NoError(t, err)

		ipc := NewIPChecker(ts)

		webhook := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		handler := ipc.RequsetIPCheck(webhook)

		srv := httptest.NewServer(handler)
		defer srv.Close()

		c := resty.New()

		req := c.R().SetHeader("X-Real-IP", "1.1.1.1")
		resp, err := req.Post(srv.URL)

		require.NoError(t, err)
		require.Equal(t, resp.StatusCode(), 403)
	})
}
