package server_test

import (
	"github.com/clambin/videoConvertor/internal/server"
	"github.com/clambin/videoConvertor/internal/server/scanner"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestServer_HTTPServer(t *testing.T) {
	cfg := server.Config{
		Addr: ":8080",
		ScannerConfig: scanner.Config{
			RootDir: "/tmp",
			Profile: "hevc-max",
		},
		RemoveConverted: false,
	}
	s, err := server.New(cfg, slog.Default())
	require.NoError(t, err)

	for _, path := range []string{"/convertor/active/on", "/convertor/active/off"} {
		req, _ := http.NewRequest(http.MethodPost, path, nil)
		resp := httptest.NewRecorder()
		s.HTTPServer.Handler.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusOK, resp.Code)
	}
}

func TestServer_Health(t *testing.T) {
	cfg := server.Config{
		Addr: ":8080",
		ScannerConfig: scanner.Config{
			RootDir: "/tmp",
			Profile: "hevc-max",
		},
		RemoveConverted: false,
	}
	s, err := server.New(cfg, slog.Default())
	require.NoError(t, err)

	req, _ := http.NewRequest(http.MethodGet, "/health", nil)
	resp := httptest.NewRecorder()
	s.Health(resp, req)
	require.Equal(t, http.StatusOK, resp.Code)
	assert.Equal(t, `{
 "Convertor": {
  "Active": false,
  "Processing": []
 },
 "Scanner": {
  "Queued": []
 }
}
`, resp.Body.String())
}
