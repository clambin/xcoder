package ffmpeg

import (
	"context"
	"encoding/json"
	"fmt"
	ffmpeg "github.com/u2takey/ffmpeg-go"
	"strconv"
	"time"
)

type VideoStats struct {
	Duration      time.Duration
	VideoCodec    string
	BitRate       int
	BitsPerSample int
	Height        int
	Width         any
}

func (p Processor) Scan(_ context.Context, path string) (VideoStats, error) {
	var probe VideoStats

	output, err := ffmpeg.Probe(path)
	if err != nil {
		return probe, fmt.Errorf("probe: %w", err)
	}

	return parse(output)
}

func parse(input string) (VideoStats, error) {
	var stats struct {
		Streams []struct {
			CodecName        string `json:"codec_name,omitempty"`
			CodecType        string `json:"codec_type"`
			BitsPerRawSample string `json:"bits_per_raw_sample,omitempty"`
			Height           int    `json:"height,omitempty"`
			Width            int    `json:"width,omitempty"`
		} `json:"streams"`
		Format struct {
			Filename string `json:"filename"`
			Duration string `json:"duration"`
			BitRate  string `json:"bit_rate"`
		} `json:"format"`
	}

	if err := json.Unmarshal([]byte(input), &stats); err != nil {
		return VideoStats{}, fmt.Errorf("json: %w", err)
	}

	var videoStats VideoStats
	duration, err := strconv.ParseFloat(stats.Format.Duration, 64)
	if err != nil {
		return VideoStats{}, fmt.Errorf("invalid duration: %w", err)
	}
	videoStats.Duration = time.Duration(duration*1000) * time.Millisecond
	bitrate, err := strconv.Atoi(stats.Format.BitRate)
	if err != nil {
		return VideoStats{}, fmt.Errorf("invalid bit_rate: %w", err)
	}
	videoStats.BitRate = bitrate

	for _, stream := range stats.Streams {
		if stream.CodecType == "video" {
			videoStats.VideoCodec = stream.CodecName
			videoStats.Height = stream.Height
			videoStats.Width = stream.Width
			switch stream.BitsPerRawSample {
			case "", "8":
				videoStats.BitsPerSample = 8
			case "10":
				videoStats.BitsPerSample = 10
			default:
				return VideoStats{}, fmt.Errorf("invalid bits_per_raw_sample %q", stream.BitsPerRawSample)
			}
		}
	}

	if videoStats.VideoCodec == "" {
		return VideoStats{}, fmt.Errorf("no video stream found")
	}

	return videoStats, nil
}

func (s VideoStats) String() string {
	output := s.VideoCodec
	if output == "" {
		return ""
	}
	if height := s.Height; height > 0 {
		output += "/" + strconv.Itoa(height)
	}
	if bitRate := s.BitRate; bitRate > 0 {
		output += "/" + Bits(bitRate).Format(2)
	}
	return output
}

type Bits int

func (b Bits) Format(decimals int) string {
	floatBits := float64(b)
	unit := "b"
	if floatBits > 1000 {
		floatBits /= 1000
		unit = "kb"
	}
	if floatBits > 1000 {
		floatBits /= 1000
		unit = "mb"
	}
	return strconv.FormatFloat(floatBits, 'f', decimals, 64) + " " + unit + "ps"
}
