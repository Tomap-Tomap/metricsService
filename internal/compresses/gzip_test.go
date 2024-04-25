package compresses

// import (
// 	"bytes"
// 	"compress/gzip"
// 	"io"
// 	"net/http"
// 	"net/http/httptest"
// 	"testing"

// 	"github.com/DarkOmap/metricsService/internal/models"
// 	"github.com/go-resty/resty/v2"
// 	"github.com/stretchr/testify/require"
// )

// func TestGzipCompression(t *testing.T) {
// 	requestBody := `{
//         "request": {
//             "type": "SimpleUtterance",
//             "command": "sudo do something"
//         },
//         "version": "1.0"
//     }`

// 	// ожидаемое содержимое тела ответа при успешном запросе
// 	successBody := `{
//         "response": {
//             "text": "Извините, я пока ничего не умею"
//         },
//         "version": "1.0"
//     }`

// 	webhook := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		w.Write([]byte(successBody))
// 	})
// 	handler := CompressHandle(webhook)

// 	srv := httptest.NewServer(handler)
// 	defer srv.Close()

// 	c := resty.New()
// 	t.Run("sends_gzip", func(t *testing.T) {
// 		buf := bytes.NewBuffer(nil)
// 		zb := gzip.NewWriter(buf)
// 		_, err := zb.Write([]byte(requestBody))
// 		require.NoError(t, err)
// 		err = zb.Close()
// 		require.NoError(t, err)

// 		resp, err := c.R().SetBody(buf).
// 			SetHeader("Content-Type", "application/json").
// 			SetHeader("Content-Encoding", "gzip").
// 			Post(srv.URL)

// 		require.NoError(t, err)
// 		require.Equal(t, http.StatusOK, resp.StatusCode())
// 		require.JSONEq(t, successBody, resp.String())
// 	})

// 	t.Run("accepts_gzip", func(t *testing.T) {
// 		buf := bytes.NewBufferString(requestBody)

// 		resp, err := c.R().SetBody(buf).
// 			SetHeader("Content-Type", "application/json").
// 			SetHeader("Accept-Encoding", "gzip").
// 			Post(srv.URL)

// 		require.NoError(t, err)
// 		require.Equal(t, http.StatusOK, resp.StatusCode())
// 		require.JSONEq(t, successBody, resp.String())
// 	})
// }

// func TestGetCompressJSON(t *testing.T) {
// 	type args struct {
// 		m *models.Metrics
// 	}
// 	tests := []struct {
// 		name string
// 		args args
// 		want string
// 	}{
// 		{
// 			name: "test compress json",
// 			args: args{models.NewMetricsForCounter("test", 25)},
// 			want: `{"id": "test", "type": "counter", "delta": 25}`,
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			got, err := GetCompressedJSON(tt.args.m)
// 			require.NoError(t, err)

// 			var buf bytes.Buffer

// 			_, err = buf.Write(got)
// 			require.NoError(t, err)
// 			zr, err := gzip.NewReader(&buf)
// 			require.NoError(t, err)
// 			b, err := io.ReadAll(zr)
// 			require.NoError(t, err)
// 			require.JSONEq(t, tt.want, string(b))
// 		})
// 	}
// }
