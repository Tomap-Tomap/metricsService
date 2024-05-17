package certmanager

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/require"
)

func TestNewEncryptManager(t *testing.T) {
	t.Run("positive test", func(t *testing.T) {
		_, err := NewEncryptManager("./testdata/test_public")
		require.NoError(t, err)
	})

	t.Run("error read file", func(t *testing.T) {
		_, err := NewEncryptManager("./testdata/empty")
		require.Error(t, err)
	})

	t.Run("error parse", func(t *testing.T) {
		_, err := NewEncryptManager("./testdata/test_public_error_key")
		require.Error(t, err)
	})
}

func TestEncryptManager_EncryptMessage(t *testing.T) {
	t.Run("positive test", func(t *testing.T) {
		em, _ := NewEncryptManager("./testdata/test_public")
		testMessage := []byte("testMessage")
		encryptMessage, err := em.EncryptMessage(testMessage)
		require.NoError(t, err)
		require.NotEqual(t, testMessage, encryptMessage)
	})

	t.Run("test short error", func(t *testing.T) {
		em, _ := NewEncryptManager("./testdata/test_short_public")

		testMessage := make([]byte, 1024)
		for i := 0; i < 1024; i++ {
			testMessage[i] = byte(i)
		}

		_, err := em.EncryptMessage(testMessage)
		require.Error(t, err)
	})
}

func TestNewDecryptManager(t *testing.T) {
	t.Run("positive test", func(t *testing.T) {
		_, err := NewDecryptManager("./testdata/test_private")
		require.NoError(t, err)
	})

	t.Run("error read file", func(t *testing.T) {
		_, err := NewDecryptManager("./testdata/empty")
		require.Error(t, err)
	})

	t.Run("error parse", func(t *testing.T) {
		_, err := NewDecryptManager("./testdata/test_private_error_key")
		require.Error(t, err)
	})
}

func TestDecryptManager_DecryptMessage(t *testing.T) {
	t.Run("positive test", func(t *testing.T) {
		em, err := NewEncryptManager("./testdata/test_public")
		require.NoError(t, err)
		testMessage := []byte("testMessage")
		encryptMessage, err := em.EncryptMessage(testMessage)
		require.NoError(t, err)
		dm, err := NewDecryptManager("./testdata/test_private")
		require.NoError(t, err)
		descryptMessage, err := dm.DecryptMessage(encryptMessage)
		require.NoError(t, err)
		require.Equal(t, testMessage, descryptMessage)
	})

	t.Run("error test", func(t *testing.T) {
		testMessage := []byte("testMessage")
		dm, err := NewDecryptManager("./testdata/test_private")
		require.NoError(t, err)
		_, err = dm.DecryptMessage(testMessage)
		require.Error(t, err)
	})
}

func TestDecryptManager_DecryptHandle(t *testing.T) {
	t.Run("positive test", func(t *testing.T) {
		em, err := NewEncryptManager("./testdata/test_public")
		require.NoError(t, err)
		testMessage := []byte("testMessage")
		encryptMessage, err := em.EncryptMessage(testMessage)
		require.NoError(t, err)

		dm, err := NewDecryptManager("./testdata/test_private")
		require.NoError(t, err)

		webhook := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		handler := dm.DecryptHandle(webhook)

		srv := httptest.NewServer(handler)
		defer srv.Close()

		c := resty.New()

		req := c.R().SetBody(encryptMessage)
		resp, err := req.Post(srv.URL)

		require.NoError(t, err)
		require.Equal(t, resp.StatusCode(), 200)
	})

	t.Run("error test", func(t *testing.T) {
		testMessage := []byte("testMessage")

		dm, err := NewDecryptManager("./testdata/test_private")
		require.NoError(t, err)

		webhook := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		handler := dm.DecryptHandle(webhook)

		srv := httptest.NewServer(handler)
		defer srv.Close()

		c := resty.New()

		req := c.R().SetBody(testMessage)
		resp, err := req.Post(srv.URL)

		require.NoError(t, err)
		require.Equal(t, resp.StatusCode(), 500)
	})
}
