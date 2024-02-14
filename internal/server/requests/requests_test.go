package requests_test

import (
	"github.com/clambin/videoConvertor/internal/server/requests"
	"github.com/clambin/videoConvertor/pkg/ffmpeg"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRequests(t *testing.T) {
	r := requests.Requests{}

	sources := []string{"1", "2", "3"}
	for _, source := range sources {
		r.Add(requests.Request{Request: ffmpeg.Request{Source: source}})
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
