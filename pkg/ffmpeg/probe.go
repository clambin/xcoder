package ffmpeg

import (
	"log/slog"
	"strconv"
	"time"
)

type Probe struct {
	Streams []Stream `json:"streams"`
	Format  Format   `json:"format"`
}

type Stream struct {
	Index              int    `json:"index"`
	CodecName          string `json:"codec_name"`
	CodecLongName      string `json:"codec_long_name"`
	Profile            string `json:"profile"`
	CodecType          string `json:"codec_type"`
	CodecTagString     string `json:"codec_tag_string"`
	CodecTag           string `json:"codec_tag"`
	Width              int    `json:"width,omitempty"`
	Height             int    `json:"height,omitempty"`
	CodedWidth         int    `json:"coded_width,omitempty"`
	CodedHeight        int    `json:"coded_height,omitempty"`
	ClosedCaptions     int    `json:"closed_captions,omitempty"`
	FilmGrain          int    `json:"film_grain,omitempty"`
	HasBFrames         int    `json:"has_b_frames,omitempty"`
	SampleAspectRatio  string `json:"sample_aspect_ratio,omitempty"`
	DisplayAspectRatio string `json:"display_aspect_ratio,omitempty"`
	PixFmt             string `json:"pix_fmt,omitempty"`
	Level              int    `json:"level,omitempty"`
	ColorRange         string `json:"color_range,omitempty"`
	ChromaLocation     string `json:"chroma_location,omitempty"`
	FieldOrder         string `json:"field_order,omitempty"`
	Refs               int    `json:"refs,omitempty"`
	RFrameRate         string `json:"r_frame_rate"`
	AvgFrameRate       string `json:"avg_frame_rate"`
	TimeBase           string `json:"time_base"`
	StartPts           int    `json:"start_pts"`
	StartTime          string `json:"start_time"`
	ExtradataSize      int    `json:"extradata_size"`
	Disposition        struct {
		Default         int `json:"default"`
		Dub             int `json:"dub"`
		Original        int `json:"original"`
		Comment         int `json:"comment"`
		Lyrics          int `json:"lyrics"`
		Karaoke         int `json:"karaoke"`
		Forced          int `json:"forced"`
		HearingImpaired int `json:"hearing_impaired"`
		VisualImpaired  int `json:"visual_impaired"`
		CleanEffects    int `json:"clean_effects"`
		AttachedPic     int `json:"attached_pic"`
		TimedThumbnails int `json:"timed_thumbnails"`
		Captions        int `json:"captions"`
		Descriptions    int `json:"descriptions"`
		Metadata        int `json:"metadata"`
		Dependent       int `json:"dependent"`
		StillImage      int `json:"still_image"`
	} `json:"disposition"`
	Tags struct {
		HANDLERNAME string `json:"HANDLER_NAME"`
		VENDORID    string `json:"VENDOR_ID"`
		ENCODER     string `json:"ENCODER,omitempty"`
		DURATION    string `json:"DURATION"`
		Language    string `json:"language,omitempty"`
	} `json:"tags"`
	SampleFmt      string `json:"sample_fmt,omitempty"`
	SampleRate     string `json:"sample_rate,omitempty"`
	Channels       int    `json:"channels,omitempty"`
	ChannelLayout  string `json:"channel_layout,omitempty"`
	BitsPerSample  int    `json:"bits_per_sample,omitempty"`
	InitialPadding int    `json:"initial_padding,omitempty"`
}

type Format struct {
	Filename       string `json:"filename"`
	NbStreams      int    `json:"nb_streams"`
	NbPrograms     int    `json:"nb_programs"`
	FormatName     string `json:"format_name"`
	FormatLongName string `json:"format_long_name"`
	StartTime      string `json:"start_time"`
	Duration       string `json:"duration"`
	Size           string `json:"size"`
	BitRate        string `json:"bit_rate"`
	ProbeScore     int    `json:"probe_score"`
	Tags           struct {
		Title            string `json:"title"`
		COMMENT          string `json:"COMMENT"`
		MAJORBRAND       string `json:"MAJOR_BRAND"`
		MINORVERSION     string `json:"MINOR_VERSION"`
		COMPATIBLEBRANDS string `json:"COMPATIBLE_BRANDS"`
		ENCODER          string `json:"ENCODER"`
	} `json:"tags"`
}

func (p Probe) VideoCodec() string {
	for i := range p.Streams {
		if p.Streams[i].CodecType == "video" {
			return p.Streams[i].CodecName
		}
	}
	return ""
}

func (p Probe) BitRate() int {
	value, _ := strconv.Atoi(p.Format.BitRate)
	return value
}

func (p Probe) Height() int {
	for i := range p.Streams {
		if p.Streams[i].CodecType == "video" {
			return p.Streams[i].Height
		}
	}
	return 0
}

func (p Probe) Duration() time.Duration {
	seconds, _ := strconv.ParseFloat(p.Format.Duration, 64)
	return time.Duration(seconds) * time.Second
}

func (p Probe) LogValue() slog.Value {
	return slog.GroupValue(
		slog.String("codec", p.VideoCodec()),
		slog.Int("bitrate", p.BitRate()/1024),
		slog.Int("height", p.Height()),
		slog.Duration("duration", p.Duration()),
	)
}
