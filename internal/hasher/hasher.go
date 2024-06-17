// Package hasher defines structures for working with hashed data.
package hasher

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"hash"
	"io"
	"net/http"

	"github.com/go-resty/resty/v2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const (
	headerHashSHA256 = "HashSHA256"
)

// Hasher It's structure witch defines methods for hashing data.
type Hasher struct {
	hasherPool chan hash.Hash
	key        []byte
}

// NewHasher create Hasher
func NewHasher(key []byte, rateLimit uint) *Hasher {
	hp := make(chan hash.Hash, rateLimit)
	return &Hasher{hp, key}
}

// Close closes Hasher
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
	req.SetHeader(headerHashSHA256, hex.EncodeToString(hash.Sum(nil)))

	return nil
}

// InterceptorAddHashMD an interceptor that adds HashSHA256 to the header
func (h *Hasher) InterceptorAddHashMD(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	if len(h.key) == 0 {
		return invoker(ctx, method, req, reply, cc, opts...)
	}

	hash, err := h.getHash()
	if err != nil {
		return fmt.Errorf("get hash: %w", err)
	}

	defer h.putHash(hash)
	hash.Write([]byte(method))

	ctx = metadata.AppendToOutgoingContext(
		ctx,
		headerHashSHA256, hex.EncodeToString(hash.Sum(nil)),
	)

	return invoker(ctx, method, req, reply, cc, opts...)
}

// RequestHash return handler for middleware.
// Handle checks the Hash SHA256 request header for compliance with the specified key.
func (h *Hasher) RequestHash(handler http.Handler) http.Handler {
	logFn := func(w http.ResponseWriter, r *http.Request) {
		hashHeader := r.Header.Get(headerHashSHA256)

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

// InterceptorCheckHash the interceptor for checking the HashSHA256 header
func (h *Hasher) InterceptorCheckHash(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
	if len(h.key) == 0 {
		resp, err = handler(ctx, req)
		return
	}

	md, ok := metadata.FromIncomingContext(ctx)

	if !ok {
		resp, err = handler(ctx, req)
		return
	}

	hashMD := md.Get(headerHashSHA256)

	if len(hashMD) == 0 {
		resp, err = handler(ctx, req)
		return
	}

	hashS := hashMD[0]

	if hashS == "" {
		resp, err = handler(ctx, req)
		return
	}

	hash, err := h.getHash()
	if err != nil {
		err = status.Error(codes.Internal, err.Error())
		return
	}

	defer h.putHash(hash)

	hash.Write([]byte(info.FullMethod))
	dst := hash.Sum(nil)
	hh, err := hex.DecodeString(hashS)
	if err != nil {
		err = status.Error(codes.Internal, err.Error())
		return
	}

	if !hmac.Equal(hh, dst) {
		err = status.Error(codes.Unauthenticated, "hash not equal")
		return
	}

	resp, err = handler(ctx, req)

	return
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

	r.ResponseWriter.Header().Add(headerHashSHA256, hex.EncodeToString(dst))

	size, err := r.ResponseWriter.Write(b)
	r.bytes += size
	return size, err
}
