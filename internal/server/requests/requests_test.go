package requests_test

import (
	requests2 "github.com/clambin/vidconv/internal/server/requests"
	"github.com/clambin/vidconv/pkg/ffmpeg"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRequests(t *testing.T) {
	r := requests2.Requests{}

	sources := []string{"1", "2", "3"}
	for _, source := range sources {
		r.Add(requests2.Request{Request: ffmpeg.Request{Source: source}})
	}

	assert.Equal(t, len(sources), r.Len())
	assert.Equal(t, sources, r.List())

	for _, source := range sources {
		req, ok := r.GetNext()
		assert.True(t, ok)
		assert.Equal(t, source, req.Source)

	}

	_, ok := r.GetNext()
	assert.False(t, ok)
}
