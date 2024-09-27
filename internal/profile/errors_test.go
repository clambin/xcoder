package profile

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestErrSourceRejected(t *testing.T) {
	err := ErrSourceRejected{Reason: "bitrate too low"}
	assert.Error(t, err)
	assert.ErrorIs(t, ErrSourceRejected{}, err)
	assert.Equal(t, "bitrate too low", err.Error())
}
