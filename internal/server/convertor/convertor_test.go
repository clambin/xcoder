package convertor_test

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/clambin/vidconv/internal/server/convertor"
	"github.com/clambin/vidconv/internal/server/requests"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestConvertor_Router(t *testing.T) {
	testCases := []struct {
		name     string
		method   string
		path     string
		body     string
		wantCode int
		wantBody string
	}{
		{
			name:     "convertor status",
			method:   http.MethodGet,
			path:     "/convertor/active",
			wantCode: http.StatusOK,
			wantBody: `{"Active":false}
`,
		},
		{
			name:     "convertor on",
			method:   http.MethodPost,
			path:     "/convertor/active/on",
			wantCode: http.StatusOK,
		},
		{
			name:     "convertor off",
			method:   http.MethodPost,
			path:     "/convertor/active/off",
			wantCode: http.StatusOK,
		},
	}

	for _, tt := range testCases {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			c := convertor.New(&requests.Requests{}, false, slog.Default())
			m := chi.NewMux()
			m.Route("/convertor", c.Router)

			r, _ := http.NewRequest(tt.method, tt.path, bytes.NewBufferString(tt.body))
			r.RemoteAddr = "localhost:12335"
			w := httptest.NewRecorder()

			m.ServeHTTP(w, r)
			assert.Equal(t, tt.wantCode, w.Code)
			assert.Equal(t, tt.wantBody, w.Body.String())
		})
	}
}

func TestConvertor_Health(t *testing.T) {
	c := convertor.New(&requests.Requests{}, false, slog.Default())

	var health bytes.Buffer
	err := json.NewEncoder(&health).Encode(c.Health(context.Background()))
	require.NoError(t, err)

	assert.Equal(t, `{"Active":false,"Processing":[]}
`, health.String())
}
