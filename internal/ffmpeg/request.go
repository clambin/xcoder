package ffmpeg

import "errors"

type Request struct {
	Source        string
	Target        string
	VideoCodec    string
	BitsPerSample int
	BitRate       int
	ProgressCB    func(Progress)
}

var ErrMissingFilename = errors.New("missing filename")
var ErrInvalidCodec = errors.New("only hevc supported")
var ErrInvalidBitsPerSample = errors.New("bits per sample must be 8 or 10")
var ErrInvalidBitRate = errors.New("invalid bitrate")

func (r Request) IsValid() error {
	if r.Source == "" || r.Target == "" {
		return ErrMissingFilename
	}
	if r.VideoCodec != "hevc" {
		return ErrInvalidCodec
	}
	if r.BitsPerSample != 8 && r.BitsPerSample != 10 {
		return ErrInvalidBitsPerSample
	}
	if r.BitRate == 0 {
		return ErrInvalidBitRate
	}
	return nil
}
