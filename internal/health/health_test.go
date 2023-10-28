package health_test

import (
	"context"
	"github.com/clambin/vidconv/internal/health"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/url"
	"testing"
	"time"
)

func TestHealth_Run(t *testing.T) {
	s := health.Health{}

	ch := make(chan error)
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		ch <- s.Run(ctx)
	}()

	time.Sleep(100 * time.Millisecond)

	cancel()
	assert.NoError(t, <-ch)
}

func TestHealth_ServeHTTP(t *testing.T) {
	testCases := []struct {
		name           string
		components     []health.HealthChecker
		wantStatusCode int
		wantBody       string
	}{
		{
			name:           "empty",
			components:     nil,
			wantStatusCode: http.StatusOK,
			wantBody:       "[]\n",
		},
		{
			name:           "simple",
			components:     []health.HealthChecker{compA{}},
			wantStatusCode: http.StatusOK,
			wantBody:       "[10]\n",
		},
		{
			name:           "structured",
			components:     []health.HealthChecker{compB{}},
			wantStatusCode: http.StatusOK,
			wantBody:       "[[{\"Name\":\"foo\",\"Value\":10},{\"Name\":\"bar\",\"Value\":20}]]\n",
		},
		{
			name:           "combined",
			components:     []health.HealthChecker{compA{}, compB{}},
			wantStatusCode: http.StatusOK,
			wantBody:       "[10,[{\"Name\":\"foo\",\"Value\":10},{\"Name\":\"bar\",\"Value\":20}]]\n",
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			s := health.Health{Components: tt.components}
			assert.Equal(t, tt.wantBody, assert.HTTPBody(s.ServeHTTP, http.MethodGet, "/health", url.Values{}))
		})
	}
}

type compA struct{}

func (c compA) Health(_ context.Context) any {
	return 10
}

type compB struct{}

func (c compB) Health(_ context.Context) any {
	return []struct {
		Name  string
		Value int
	}{{Name: "foo", Value: 10}, {Name: "bar", Value: 20}}
}
