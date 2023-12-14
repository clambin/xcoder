package quality_test

import (
	"github.com/clambin/videoConvertor/internal/server/scanner/inspector/quality"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestErrSourceRejected(t *testing.T) {
	err := quality.ErrSourceRejected{Reason: "too small"}
	assert.ErrorIs(t, quality.ErrSourceRejected{}, err)

	err = quality.ErrSourceRejected{Reason: "bitrate too low"}
	assert.ErrorIs(t, quality.ErrSourceRejected{}, err)
}
