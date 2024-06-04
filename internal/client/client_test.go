package client

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DarkOmap/metricsService/internal/hasher"
	"github.com/DarkOmap/metricsService/internal/parameters"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type CompresserMockedObject struct {
	mock.Mock
}

func (c *CompresserMockedObject) GetCompressedJSON(m any) ([]byte, error) {
	args := c.Called()

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).([]byte), args.Error(1)
}

func (c *CompresserMockedObject) Close() {

}

type EncrypterMockedObject struct {
	mock.Mock
}

func (e *EncrypterMockedObject) EncryptMessage(m []byte) ([]byte, error) {
	args := e.Called()

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).([]byte), args.Error(1)
}

func TestSendGauge(t *testing.T) {
	t.Run("not OK answer", func(t *testing.T) {
		t.Parallel()
		cmo := new(CompresserMockedObject)
		cmo.On("GetCompressedJSON").Return([]byte("test"), nil)

		emo := new(EncrypterMockedObject)
		emo.On("EncryptMessage").Return([]byte("test"), nil)

		ts := httptest.NewServer(
			http.HandlerFunc(
				func(w http.ResponseWriter, r *http.Request) {
					http.Error(w, "test error", http.StatusBadRequest)
				},
			),
		)
		defer ts.Close()

		h := hasher.NewHasher(make([]byte, 0), 1)
		c := ClientHTTP{
			gp:        cmo,
			encrypter: emo,
			h:         h,
			addr:      strings.TrimPrefix(ts.URL, "http://"),
		}
		c.setRestyClient()
		err := c.SendGauge(context.Background(), "test", 1.1)

		require.Error(t, err)
	})

	t.Run("OK answer", func(t *testing.T) {
		t.Parallel()
		cmo := new(CompresserMockedObject)
		cmo.On("GetCompressedJSON").Return([]byte("test"), nil)

		emo := new(EncrypterMockedObject)
		emo.On("EncryptMessage").Return([]byte("test"), nil)

		ts := httptest.NewServer(
			http.HandlerFunc(
				func(w http.ResponseWriter, r *http.Request) {
				},
			),
		)
		defer ts.Close()

		h := hasher.NewHasher(make([]byte, 0), 1)
		c := ClientHTTP{
			gp:        cmo,
			encrypter: emo,
			h:         h,
			addr:      strings.TrimPrefix(ts.URL, "http://"),
		}
		c.setRestyClient()
		err := c.SendGauge(context.Background(), "test", 1.1)

		require.NoError(t, err)
	})

	t.Run("test broken server", func(t *testing.T) {
		t.Parallel()
		cmo := new(CompresserMockedObject)
		cmo.On("GetCompressedJSON").Return([]byte("test"), nil)

		emo := new(EncrypterMockedObject)
		emo.On("EncryptMessage").Return([]byte("test"), nil)

		h := hasher.NewHasher(make([]byte, 0), 1)
		c := ClientHTTP{
			gp:        cmo,
			encrypter: emo,
			h:         h,
			addr:      "test",
		}
		c.setRestyClient()
		err := c.SendGauge(context.Background(), "test", 1.1)

		assert.Error(t, err)
	})

	t.Run("test error compressed", func(t *testing.T) {
		t.Parallel()
		cmo := new(CompresserMockedObject)
		cmo.On("GetCompressedJSON").Return(nil, fmt.Errorf("test error"))

		emo := new(EncrypterMockedObject)
		emo.On("EncryptMessage").Return([]byte("test"), nil)

		h := hasher.NewHasher(make([]byte, 0), 1)
		c := ClientHTTP{
			gp:        cmo,
			encrypter: emo,
			h:         h,
			addr:      "test",
		}
		c.setRestyClient()
		err := c.SendGauge(context.Background(), "test", 1.1)

		assert.Error(t, err)
	})

	t.Run("test error encrypt", func(t *testing.T) {
		t.Parallel()
		cmo := new(CompresserMockedObject)
		cmo.On("GetCompressedJSON").Return([]byte("test"), nil)

		emo := new(EncrypterMockedObject)
		emo.On("EncryptMessage").Return(nil, fmt.Errorf("test error"))

		h := hasher.NewHasher(make([]byte, 0), 1)
		c := ClientHTTP{
			gp:        cmo,
			encrypter: emo,
			h:         h,
			addr:      "test",
		}
		c.setRestyClient()
		err := c.SendGauge(context.Background(), "test", 1.1)

		assert.Error(t, err)
	})

	t.Run("hasher error", func(t *testing.T) {
		t.Parallel()
		cmo := new(CompresserMockedObject)
		cmo.On("GetCompressedJSON").Return([]byte("test"), nil)

		emo := new(EncrypterMockedObject)
		emo.On("EncryptMessage").Return([]byte("test"), nil)

		ts := httptest.NewServer(
			http.HandlerFunc(
				func(w http.ResponseWriter, r *http.Request) {
				},
			),
		)
		defer ts.Close()

		h := hasher.NewHasher([]byte("test"), 0)
		h.Close()
		c := ClientHTTP{
			gp:        cmo,
			encrypter: emo,
			h:         h,
			addr:      strings.TrimPrefix(ts.URL, "http://"),
		}
		c.setRestyClient()
		err := c.SendGauge(context.Background(), "test", 1.1)

		require.Error(t, err)
	})
}

func TestSendCounter(t *testing.T) {
	t.Run("not OK answer", func(t *testing.T) {
		t.Parallel()
		cmo := new(CompresserMockedObject)
		cmo.On("GetCompressedJSON").Return([]byte("test"), nil)

		emo := new(EncrypterMockedObject)
		emo.On("EncryptMessage").Return([]byte("test"), nil)

		ts := httptest.NewServer(
			http.HandlerFunc(
				func(w http.ResponseWriter, r *http.Request) {
					http.Error(w, "test error", http.StatusBadRequest)
				},
			),
		)
		defer ts.Close()

		h := hasher.NewHasher(make([]byte, 0), 1)
		c := ClientHTTP{
			gp:        cmo,
			encrypter: emo,
			h:         h,
			addr:      strings.TrimPrefix(ts.URL, "http://"),
		}
		c.setRestyClient()
		err := c.SendCounter(context.Background(), "test", 1)

		require.Error(t, err)
	})

	t.Run("OK answer", func(t *testing.T) {
		t.Parallel()
		cmo := new(CompresserMockedObject)
		cmo.On("GetCompressedJSON").Return([]byte("test"), nil)

		emo := new(EncrypterMockedObject)
		emo.On("EncryptMessage").Return([]byte("test"), nil)

		ts := httptest.NewServer(
			http.HandlerFunc(
				func(w http.ResponseWriter, r *http.Request) {
				},
			),
		)
		defer ts.Close()

		h := hasher.NewHasher(make([]byte, 0), 1)
		c := ClientHTTP{
			gp:        cmo,
			encrypter: emo,
			h:         h,
			addr:      strings.TrimPrefix(ts.URL, "http://"),
		}
		c.setRestyClient()
		err := c.SendCounter(context.Background(), "test", 1)

		require.NoError(t, err)
	})

	t.Run("test broken server", func(t *testing.T) {
		t.Parallel()
		cmo := new(CompresserMockedObject)
		cmo.On("GetCompressedJSON").Return([]byte("test"), nil)

		emo := new(EncrypterMockedObject)
		emo.On("EncryptMessage").Return([]byte("test"), nil)

		h := hasher.NewHasher(make([]byte, 0), 1)
		c := ClientHTTP{
			gp:        cmo,
			encrypter: emo,
			h:         h,
			addr:      "test",
		}
		c.setRestyClient()
		err := c.SendCounter(context.Background(), "test", 1)

		assert.Error(t, err)
	})

	t.Run("test error compressed", func(t *testing.T) {
		t.Parallel()
		cmo := new(CompresserMockedObject)
		cmo.On("GetCompressedJSON").Return(nil, fmt.Errorf("test error"))

		emo := new(EncrypterMockedObject)
		emo.On("EncryptMessage").Return([]byte("test"), nil)

		h := hasher.NewHasher(make([]byte, 0), 1)
		c := ClientHTTP{
			gp:        cmo,
			encrypter: emo,
			h:         h,
			addr:      "test",
		}
		c.setRestyClient()
		err := c.SendCounter(context.Background(), "test", 1)

		assert.Error(t, err)
	})

	t.Run("test error encrypt", func(t *testing.T) {
		t.Parallel()
		cmo := new(CompresserMockedObject)
		cmo.On("GetCompressedJSON").Return([]byte("test"), nil)

		emo := new(EncrypterMockedObject)
		emo.On("EncryptMessage").Return(nil, fmt.Errorf("test error"))

		h := hasher.NewHasher(make([]byte, 0), 1)
		c := ClientHTTP{
			gp:        cmo,
			encrypter: emo,
			h:         h,
			addr:      "test",
		}
		c.setRestyClient()
		err := c.SendCounter(context.Background(), "test", 1)

		assert.Error(t, err)
	})

	t.Run("hasher error", func(t *testing.T) {
		t.Parallel()
		cmo := new(CompresserMockedObject)
		cmo.On("GetCompressedJSON").Return([]byte("test"), nil)

		emo := new(EncrypterMockedObject)
		emo.On("EncryptMessage").Return([]byte("test"), nil)

		ts := httptest.NewServer(
			http.HandlerFunc(
				func(w http.ResponseWriter, r *http.Request) {
				},
			),
		)
		defer ts.Close()

		h := hasher.NewHasher([]byte("test"), 0)
		h.Close()
		c := ClientHTTP{
			gp:        cmo,
			encrypter: emo,
			h:         h,
			addr:      strings.TrimPrefix(ts.URL, "http://"),
		}
		c.setRestyClient()
		err := c.SendCounter(context.Background(), "test", 1)

		require.Error(t, err)
	})
}

func TestClient_SendBatch(t *testing.T) {
	cmo := new(CompresserMockedObject)
	cmo.On("GetCompressedJSON").Return([]byte("test"), nil)

	emo := new(EncrypterMockedObject)
	emo.On("EncryptMessage").Return([]byte("test"), nil)

	t.Run("not OK answer", func(t *testing.T) {
		t.Parallel()
		hf := func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "test error", http.StatusBadRequest)
		}

		ts := httptest.NewServer(http.HandlerFunc(hf))
		defer ts.Close()

		h := hasher.NewHasher(make([]byte, 0), 1)
		c := ClientHTTP{
			gp:        cmo,
			encrypter: emo,
			h:         h,
			addr:      strings.TrimPrefix(ts.URL, "http://"),
		}
		c.setRestyClient()
		err := c.SendBatch(context.Background(), map[string]float64{"test": 44})
		assert.Error(t, err)
	})

	t.Run("OK answer", func(t *testing.T) {
		t.Parallel()
		hf := func(w http.ResponseWriter, r *http.Request) {
		}

		ts := httptest.NewServer(http.HandlerFunc(hf))
		defer ts.Close()

		h := hasher.NewHasher(make([]byte, 0), 1)
		c := ClientHTTP{
			gp:        cmo,
			encrypter: emo,
			h:         h,
			addr:      strings.TrimPrefix(ts.URL, "http://"),
		}
		c.setRestyClient()
		err := c.SendBatch(context.Background(), map[string]float64{"test": 44})
		assert.NoError(t, err)
	})

	t.Run("test broken server", func(t *testing.T) {
		t.Parallel()
		h := hasher.NewHasher(make([]byte, 0), 1)
		c := ClientHTTP{
			gp:        cmo,
			encrypter: emo,
			h:         h,
			addr:      "test",
		}
		c.setRestyClient()
		err := c.SendBatch(context.Background(), map[string]float64{"test": 44})

		assert.Error(t, err)
	})

	t.Run("test error compressed", func(t *testing.T) {
		t.Parallel()
		cmo := new(CompresserMockedObject)
		cmo.On("GetCompressedJSON").Return(nil, fmt.Errorf("test error"))
		h := hasher.NewHasher(make([]byte, 0), 1)
		c := ClientHTTP{
			gp:        cmo,
			encrypter: emo,
			h:         h,
			addr:      "test",
		}
		c.setRestyClient()
		err := c.SendBatch(context.Background(), map[string]float64{"test": 44})

		assert.Error(t, err)
	})

	t.Run("test error encrypt", func(t *testing.T) {
		t.Parallel()

		emo := new(EncrypterMockedObject)
		emo.On("EncryptMessage").Return(nil, fmt.Errorf("test error"))

		h := hasher.NewHasher(make([]byte, 0), 1)
		c := ClientHTTP{
			gp:        cmo,
			encrypter: emo,
			h:         h,
			addr:      "test",
		}
		c.setRestyClient()
		err := c.SendBatch(context.Background(), map[string]float64{"test": 44})

		assert.Error(t, err)
	})

	t.Run("hasher error", func(t *testing.T) {
		t.Parallel()
		hf := func(w http.ResponseWriter, r *http.Request) {
		}

		ts := httptest.NewServer(http.HandlerFunc(hf))
		defer ts.Close()

		h := hasher.NewHasher([]byte("test"), 0)
		h.Close()
		c := ClientHTTP{
			gp:        cmo,
			encrypter: emo,
			h:         h,
			addr:      strings.TrimPrefix(ts.URL, "http://"),
		}
		c.setRestyClient()
		err := c.SendBatch(context.Background(), map[string]float64{"test": 44})
		assert.Error(t, err)
	})
}

func TestNewClientHTTP(t *testing.T) {
	t.Run("positive test", func(t *testing.T) {
		c, err := NewClientHTTP(parameters.AgentParameters{
			CryptoKeyPath: "./testdata/test_public",
			HashKey:       "",
			RateLimit:     1,
			ListenAddr:    ":0",
		})

		require.NoError(t, err)
		defer c.Close()
		require.NotEmpty(t, c)
	})

	t.Run("key path error test", func(t *testing.T) {
		_, err := NewClientHTTP(parameters.AgentParameters{
			CryptoKeyPath: "",
			HashKey:       "",
			RateLimit:     1,
			ListenAddr:    ":0",
		})

		require.Error(t, err)
	})
}
