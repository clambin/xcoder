package ffmpeg

import (
	"encoding/json"
	"fmt"
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

func Parse(input string) (VideoStats, error) {
	var stats map[string]any

	if err := json.Unmarshal([]byte(input), &stats); err != nil {
		return VideoStats{}, fmt.Errorf("json: %w", err)
	}

	var s VideoStats
	var ok bool

	format := stats["format"].(map[string]any)

	// if s.Source, ok = format["filename"].(string); !ok {
	//	return VideoStats{}, fmt.Errorf("json: missing filename")
	//}

	duration, ok := format["duration"].(string)
	if !ok {
		return VideoStats{}, fmt.Errorf("missing duration")
	}
	duration2, err := strconv.ParseFloat(duration, 64)
	if err != nil {
		return VideoStats{}, fmt.Errorf("invalid duration: %w", err)
	}
	s.Duration = time.Duration(duration2*1000) * time.Millisecond

	bitrate, ok := format["bit_rate"].(string)
	if !ok {
		return VideoStats{}, fmt.Errorf("missing bit_rate")
	}
	s.BitRate, err = strconv.Atoi(bitrate)
	if err != nil {
		return VideoStats{}, fmt.Errorf("json: invalid bit_rate: %w", err)
	}

	streams, ok := stats["streams"].([]any)
	if !ok {
		return VideoStats{}, fmt.Errorf("missing streams")
	}

	for _, stream := range streams {
		stream2 := stream.(map[string]any)
		if stream2["codec_type"].(string) != "video" {
			continue
		}
		s.VideoCodec = stream2["codec_name"].(string)
		bitsPerSample := stream2["bits_per_raw_sample"]
		if bitsPerSample == nil {
			s.BitsPerSample = 8
		} else {
			if s.BitsPerSample, err = strconv.Atoi(bitsPerSample.(string)); err != nil {
				return VideoStats{}, fmt.Errorf("invalid bits_per_sample %q: %w", bitsPerSample.(string), err)
			}

		}
		s.Height = int(stream2["height"].(float64))
		s.Width = int(stream2["width"].(float64))
	}
	if s.VideoCodec == "" {
		return VideoStats{}, fmt.Errorf("no video stream found")
	}
	return s, nil
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
