// Package hasher defines structures for working with hashed data.
package hasher

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"hash"
	"io"
	"net/http"

	"github.com/go-resty/resty/v2"
)

// Hasher It's structure witch defines methods for hashing data.
type Hasher struct {
	hasherPool chan hash.Hash
	key        []byte
}

func NewHasher(key []byte, rateLimit uint) Hasher {
	hp := make(chan hash.Hash, rateLimit)
	return Hasher{hp, key}
}

func (h *Hasher) Close() {
	close(h.hasherPool)
}

// HashingRequest adds HashSHA256 value in header.
// HashSHA256 contains body hashed with key.
func (h *Hasher) HashingRequest(req *resty.Request, body []byte) error {
	if len(h.key) == 0 {
		return nil
	}

	hash, err := h.getHash()

	if err != nil {
		return fmt.Errorf("get hash: %w", err)
	}

	defer h.putHash(hash)
	hash.Write(body)
	req.SetHeader("HashSHA256", hex.EncodeToString(hash.Sum(nil)))

	return nil
}

// RequestHash return handler for middleware.
// Handle checks the Hash SHA256 request header for compliance with the specified key.
func (h *Hasher) RequestHash(handler http.Handler) http.Handler {
	logFn := func(w http.ResponseWriter, r *http.Request) {
		hashHeader := r.Header.Get("HashSHA256")

		if len(h.key) == 0 || hashHeader == "" {
			handler.ServeHTTP(w, r)
			return
		}

		var buf bytes.Buffer
		_, err := buf.ReadFrom(r.Body)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		r.Body = io.NopCloser(&buf)
		hash, err := h.getHash()

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		defer h.putHash(hash)

		hash.Write(buf.Bytes())
		dst := hash.Sum(nil)
		hh, err := hex.DecodeString(hashHeader)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if !hmac.Equal(hh, dst) {
			http.Error(w, "hash not equal", http.StatusBadRequest)
			return
		}

		hw := hashingResponseWriter{
			ResponseWriter: w,
			key:            h.key,
			hasher:         h,
		}

		handler.ServeHTTP(&hw, r)
	}

	return http.HandlerFunc(logFn)
}

func (h *Hasher) getHash() (hash.Hash, error) {
	select {
	case w, ok := <-h.hasherPool:
		if !ok {
			return nil, fmt.Errorf("pool is closed")
		}

		return w, nil
	default:
	}

	return hmac.New(sha256.New, h.key), nil
}

func (h *Hasher) putHash(ph hash.Hash) {
	ph.Reset()
	select {
	case h.hasherPool <- ph:
	default:
	}
}

type hashingResponseWriter struct {
	http.ResponseWriter
	hasher *Hasher
	key    []byte
	bytes  int
}

func (r *hashingResponseWriter) Write(b []byte) (int, error) {
	h, err := r.hasher.getHash()

	if err != nil {
		return 0, fmt.Errorf("get hash: %w", err)
	}

	h.Write(b)
	dst := h.Sum(nil)

	r.ResponseWriter.Header().Add("HashSHA256", hex.EncodeToString(dst))

	size, err := r.ResponseWriter.Write(b)
	r.bytes += size
	return size, err
}
