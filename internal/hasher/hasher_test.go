package hasher

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/require"
)

func TestHasher_HashingRequest(t *testing.T) {
	t.Run("empty hash", func(t *testing.T) {
		h := NewHasher(make([]byte, 0), 1)

		body := []byte("test")
		req := resty.New().R().SetBody(body)

		h.HashingRequest(req, body)

		hashHeader := req.Header.Get("HashSHA256")

		require.Empty(t, hashHeader)
	})

	t.Run("HashSHA256 equal", func(t *testing.T) {
		key := []byte("test")
		h := NewHasher(key, 1)

		body := []byte("test")
		req := resty.New().R().SetBody(body)

		h.HashingRequest(req, body)

		hashHeader := req.Header.Get("HashSHA256")

		require.NotEmpty(t, hashHeader)

		hash := hmac.New(sha256.New, key)
		hash.Write(body)
		dst := hash.Sum(nil)
		hh, err := hex.DecodeString(hashHeader)

		require.NoError(t, err)
		require.True(t, hmac.Equal(hh, dst))
	})

	t.Run("HashSHA256 not equal", func(t *testing.T) {
		key := []byte("test")
		h := NewHasher(key, 1)

		body := []byte("test")
		req := resty.New().R().SetBody(body)

		h.HashingRequest(req, body)

		hashHeader := req.Header.Get("HashSHA256")

		require.NotEmpty(t, hashHeader)

		hash := hmac.New(sha256.New, []byte("test123"))
		hash.Write(body)
		dst := hash.Sum(nil)
		hh, err := hex.DecodeString(hashHeader)

		require.NoError(t, err)
		require.False(t, hmac.Equal(hh, dst))
	})

	t.Run("closed pool", func(t *testing.T) {
		key := []byte("test")
		h := NewHasher(key, 0)
		h.Close()

		body := []byte("test")
		req := resty.New().R().SetBody(body)

		err := h.HashingRequest(req, body)
		require.Error(t, err)
	})
}

func TestHasher_RequestHash(t *testing.T) {
	t.Run("positive test", func(t *testing.T) {
		body := []byte("test")
		h := NewHasher([]byte("test"), 1)
		webhook := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write(body)
		})
		handler := h.RequestHash(webhook)

		srv := httptest.NewServer(handler)
		defer srv.Close()

		c := resty.New()

		req := c.R().SetBody(body)
		h.HashingRequest(req, body)

		resp, err := req.Post(srv.URL)

		require.NoError(t, err)
		require.Equal(t, resp.StatusCode(), 200)
	})

	t.Run("no key test", func(t *testing.T) {
		body := []byte("test")
		h := NewHasher(make([]byte, 0), 1)
		webhook := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write(body)
		})
		handler := h.RequestHash(webhook)

		srv := httptest.NewServer(handler)
		defer srv.Close()

		c := resty.New()

		req := c.R().SetBody(body)
		h.HashingRequest(req, body)

		resp, err := req.Post(srv.URL)

		require.NoError(t, err)
		require.Equal(t, resp.StatusCode(), 200)
	})

	t.Run("no equal test", func(t *testing.T) {
		body := []byte("test")
		h := NewHasher([]byte("test"), 1)
		webhook := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write(body)
		})
		handler := h.RequestHash(webhook)

		srv := httptest.NewServer(handler)
		defer srv.Close()

		c := resty.New()

		req := c.R().SetBody(body)
		h2 := NewHasher([]byte("test2"), 1)
		h2.HashingRequest(req, body)

		resp, err := req.Post(srv.URL)

		require.NoError(t, err)
		require.Equal(t, resp.StatusCode(), 400)
	})

	t.Run("closed pool", func(t *testing.T) {
		body := []byte("test")
		h := NewHasher([]byte("test"), 0)
		webhook := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write(body)
		})
		handler := h.RequestHash(webhook)

		srv := httptest.NewServer(handler)
		defer srv.Close()

		c := resty.New()

		req := c.R().SetBody(body)
		h.HashingRequest(req, body)
		h.Close()

		resp, err := req.Post(srv.URL)

		require.NoError(t, err)
		require.Equal(t, resp.StatusCode(), 500)
	})
}

func BenchmarkHashingRequest(b *testing.B) {
	key := []byte("test")
	h := NewHasher(key, 10)

	body := []byte("test")
	req := resty.New().R().SetBody(body)
	b.ResetTimer()
	for i := 1; i < b.N; i++ {
		h.HashingRequest(req, body)
	}
}

func BenchmarkRequestHash(b *testing.B) {
	body := []byte("test")
	h := NewHasher([]byte("test"), 10)
	webhook := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(body)
	})
	handler := h.RequestHash(webhook)

	srv := httptest.NewServer(handler)
	defer srv.Close()

	c := resty.New()

	req := c.R().SetBody(body)
	h.HashingRequest(req, body)
	b.ResetTimer()
	for i := 1; i < b.N; i++ {
		req.Post(srv.URL)
	}
}
