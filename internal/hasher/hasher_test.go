package hasher

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/interop"
	testgrpc "google.golang.org/grpc/interop/grpc_testing"
	"google.golang.org/grpc/metadata"
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

	t.Run("error decode", func(t *testing.T) {
		body := []byte("test")
		h := NewHasher([]byte("test"), 1)
		webhook := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write(body)
		})
		handler := h.RequestHash(webhook)

		srv := httptest.NewServer(handler)
		defer srv.Close()

		c := resty.New()

		req := c.R().SetBody(body).SetHeader("HashSHA256", "123")

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

func TestHasher_InterceptorAddHashMD(t *testing.T) {
	lis, err := net.Listen("tcp", "localhost:0")
	require.NoError(t, err)

	s := grpc.NewServer()

	testgrpc.RegisterTestServiceServer(
		s,
		interop.NewTestServer(),
	)

	go func() {
		if err := s.Serve(lis); err != nil {
			require.FailNow(t, err.Error())
		}
	}()

	defer s.Stop()
	t.Run("positive test", func(t *testing.T) {
		key := []byte("test")
		h := NewHasher(key, 1)

		conn, err := grpc.NewClient(
			lis.Addr().String(),
			grpc.WithTransportCredentials(insecure.NewCredentials()),
			grpc.WithUnaryInterceptor(h.InterceptorAddHashMD),
		)

		require.NoError(t, err)
		defer conn.Close()

		client := testgrpc.NewTestServiceClient(conn)
		ctx := context.Background()
		_, err = client.EmptyCall(ctx, &testgrpc.Empty{})

		require.NoError(t, err)
	})

	t.Run("empty hash", func(t *testing.T) {
		h := NewHasher(make([]byte, 0), 1)

		conn, err := grpc.NewClient(
			lis.Addr().String(),
			grpc.WithTransportCredentials(insecure.NewCredentials()),
			grpc.WithUnaryInterceptor(h.InterceptorAddHashMD),
		)

		require.NoError(t, err)
		defer conn.Close()

		client := testgrpc.NewTestServiceClient(conn)
		ctx := context.Background()
		_, err = client.EmptyCall(ctx, &testgrpc.Empty{})

		require.NoError(t, err)
	})

	t.Run("closed pool", func(t *testing.T) {
		key := []byte("test")
		h := NewHasher(key, 0)
		h.Close()

		conn, err := grpc.NewClient(
			lis.Addr().String(),
			grpc.WithTransportCredentials(insecure.NewCredentials()),
			grpc.WithUnaryInterceptor(h.InterceptorAddHashMD),
		)

		require.NoError(t, err)
		defer conn.Close()

		client := testgrpc.NewTestServiceClient(conn)
		ctx := context.Background()
		_, err = client.EmptyCall(ctx, &testgrpc.Empty{})

		require.Error(t, err)
	})
}

func TestHasher_InterceptorCheckHash(t *testing.T) {
	t.Run("empty hash", func(t *testing.T) {
		h := NewHasher(make([]byte, 0), 1)

		lis, err := net.Listen("tcp", "localhost:0")
		require.NoError(t, err)

		s := grpc.NewServer(grpc.UnaryInterceptor(h.InterceptorCheckHash))

		testgrpc.RegisterTestServiceServer(
			s,
			interop.NewTestServer(),
		)

		go func() {
			if err := s.Serve(lis); err != nil {
				require.FailNow(t, err.Error())
			}
		}()

		defer s.Stop()

		conn, err := grpc.NewClient(
			lis.Addr().String(),
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		)

		require.NoError(t, err)
		defer conn.Close()

		client := testgrpc.NewTestServiceClient(conn)
		ctx := context.Background()
		_, err = client.EmptyCall(ctx, &testgrpc.Empty{})

		require.NoError(t, err)
	})

	t.Run("closed pool", func(t *testing.T) {
		key := []byte("test")
		h := NewHasher(key, 1)

		lis, err := net.Listen("tcp", "localhost:0")
		require.NoError(t, err)

		s := grpc.NewServer(grpc.UnaryInterceptor(h.InterceptorCheckHash))

		testgrpc.RegisterTestServiceServer(
			s,
			interop.NewTestServer(),
		)

		go func() {
			if err := s.Serve(lis); err != nil {
				require.FailNow(t, err.Error())
			}
		}()

		defer s.Stop()

		h.Close()

		conn, err := grpc.NewClient(
			lis.Addr().String(),
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		)

		require.NoError(t, err)
		defer conn.Close()

		client := testgrpc.NewTestServiceClient(conn)
		ctx := metadata.AppendToOutgoingContext(
			context.Background(),
			"HashSHA256", "123",
		)
		_, err = client.EmptyCall(ctx, &testgrpc.Empty{})

		require.Error(t, err)
	})

	key := []byte("test")
	h := NewHasher(key, 1)

	lis, err := net.Listen("tcp", "localhost:0")
	require.NoError(t, err)

	s := grpc.NewServer(grpc.UnaryInterceptor(h.InterceptorCheckHash))

	testgrpc.RegisterTestServiceServer(
		s,
		interop.NewTestServer(),
	)

	go func() {
		if err := s.Serve(lis); err != nil {
			require.FailNow(t, err.Error())
		}
	}()

	defer s.Stop()

	t.Run("empty md HashSHA256", func(t *testing.T) {
		conn, err := grpc.NewClient(
			lis.Addr().String(),
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		)

		require.NoError(t, err)
		defer conn.Close()

		client := testgrpc.NewTestServiceClient(conn)
		ctx := context.Background()
		_, err = client.EmptyCall(ctx, &testgrpc.Empty{})

		require.NoError(t, err)
	})

	t.Run("empty HashSHA256", func(t *testing.T) {
		conn, err := grpc.NewClient(
			lis.Addr().String(),
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		)

		require.NoError(t, err)
		defer conn.Close()

		client := testgrpc.NewTestServiceClient(conn)

		ctx := metadata.AppendToOutgoingContext(
			context.Background(),
			"HashSHA256", "",
		)

		_, err = client.EmptyCall(ctx, &testgrpc.Empty{})

		require.NoError(t, err)
	})

	t.Run("error decode", func(t *testing.T) {
		conn, err := grpc.NewClient(
			lis.Addr().String(),
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		)

		require.NoError(t, err)
		defer conn.Close()

		client := testgrpc.NewTestServiceClient(conn)

		ctx := metadata.AppendToOutgoingContext(
			context.Background(),
			"HashSHA256", "123",
		)

		_, err = client.EmptyCall(ctx, &testgrpc.Empty{})

		require.Error(t, err)
	})

	t.Run("not equal hash", func(t *testing.T) {
		key2 := []byte("test2")
		h2 := NewHasher(key2, 1)

		conn, err := grpc.NewClient(
			lis.Addr().String(),
			grpc.WithTransportCredentials(insecure.NewCredentials()),
			grpc.WithUnaryInterceptor(h2.InterceptorAddHashMD),
		)

		require.NoError(t, err)
		defer conn.Close()

		client := testgrpc.NewTestServiceClient(conn)

		ctx := context.Background()

		_, err = client.EmptyCall(ctx, &testgrpc.Empty{})

		require.Error(t, err)
	})

	t.Run("positive test", func(t *testing.T) {
		conn, err := grpc.NewClient(
			lis.Addr().String(),
			grpc.WithTransportCredentials(insecure.NewCredentials()),
			grpc.WithUnaryInterceptor(h.InterceptorAddHashMD),
		)

		require.NoError(t, err)
		defer conn.Close()

		client := testgrpc.NewTestServiceClient(conn)

		ctx := context.Background()

		_, err = client.EmptyCall(ctx, &testgrpc.Empty{})

		require.NoError(t, err)
	})
}
