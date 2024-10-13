package ffmpeg

import "errors"

type Request struct {
	Source      string
	Target      string
	TargetStats VideoStats
	ProgressCB  func(Progress)
}

var ErrMissingFilename = errors.New("missing filename")
var ErrInvalidCodec = errors.New("only hevc supported")
var ErrInvalidBitsPerSample = errors.New("bits per sample must be 8 or 10")
var ErrInvalidBitRate = errors.New("invalid bitrate")

func (r Request) IsValid() error {
	if r.Source == "" || r.Target == "" {
		return ErrMissingFilename
	}
	if r.TargetStats.VideoCodec != "hevc" {
		return ErrInvalidCodec
	}
	if r.TargetStats.BitsPerSample != 8 && r.TargetStats.BitsPerSample != 10 {
		return ErrInvalidBitsPerSample
	}
	if r.TargetStats.BitRate == 0 {
		return ErrInvalidBitRate
	}
	return nil
}
