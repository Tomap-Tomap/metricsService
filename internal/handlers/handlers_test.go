package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUpdate(t *testing.T) {
	globalCT := "text/plain; charset=utf-8"
	type want struct {
		code        int
		contentType string
	}
	tests := []struct {
		name string
		req  *http.Request
		want want
	}{
		{
			name: "get method test",
			req:  httptest.NewRequest(http.MethodGet, "/gauge/test/123", nil),
			want: want{405, globalCT},
		},
		{
			name: "short url test",
			req:  httptest.NewRequest(http.MethodPost, "/gauge/test", nil),
			want: want{404, globalCT},
		},
		{
			name: "wrong value test",
			req:  httptest.NewRequest(http.MethodPost, "/gauge/test/wrong", nil),
			want: want{400, globalCT},
		},
		{
			name: "positive test",
			req:  httptest.NewRequest(http.MethodPost, "/gauge/test/12", nil),
			want: want{200, globalCT},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			Update(w, tt.req)

			res := w.Result()

			assert.Equal(t, tt.want.code, res.StatusCode)
			assert.Equal(t, tt.want.contentType, res.Header.Get("Content-Type"))
		})
	}
}
