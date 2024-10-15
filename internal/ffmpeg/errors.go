package ffmpeg

import (
	"errors"
	"fmt"
)

var ErrMissingFilename = errors.New("missing filename")
var ErrInvalidCodec = errors.New("only hevc supported")
var ErrInvalidBitsPerSample = errors.New("bits per sample must be 8 or 10")
var ErrMissingHeight = errors.New("height must be greater than zero")

type ErrInvalidConstantRateFactor struct {
	ConstantRateFactor int
}

func (e ErrInvalidConstantRateFactor) Error() string {
	return fmt.Sprintf("invalid Constant Rate Factor: %d", e.ConstantRateFactor)
}
