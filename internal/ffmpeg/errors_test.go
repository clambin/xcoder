package ffmpeg

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestErrInvalidConstantRateFactor_Error(t *testing.T) {
	err := ErrInvalidConstantRateFactor{ConstantRateFactor: -1}
	assert.Equal(t, "invalid Constant Rate Factor: -1", err.Error())
}
