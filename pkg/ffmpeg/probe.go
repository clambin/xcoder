package ffmpeg

import (
	"log/slog"
	"strconv"
	"time"
)

type Probe struct {
	Format  Format   `json:"format"`
	Streams []Stream `json:"streams"`
}

type Stream struct {
	Disposition struct {
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
	CodecTagString     string `json:"codec_tag_string"`
	CodecName          string `json:"codec_name"`
	CodecType          string `json:"codec_type"`
	RFrameRate         string `json:"r_frame_rate"`
	CodecTag           string `json:"codec_tag"`
	ChannelLayout      string `json:"channel_layout,omitempty"`
	SampleRate         string `json:"sample_rate,omitempty"`
	SampleFmt          string `json:"sample_fmt,omitempty"`
	CodecLongName      string `json:"codec_long_name"`
	Profile            string `json:"profile"`
	StartTime          string `json:"start_time"`
	TimeBase           string `json:"time_base"`
	SampleAspectRatio  string `json:"sample_aspect_ratio,omitempty"`
	DisplayAspectRatio string `json:"display_aspect_ratio,omitempty"`
	PixFmt             string `json:"pix_fmt,omitempty"`
	AvgFrameRate       string `json:"avg_frame_rate"`
	ColorRange         string `json:"color_range,omitempty"`
	ChromaLocation     string `json:"chroma_location,omitempty"`
	FieldOrder         string `json:"field_order,omitempty"`
	CodedWidth         int    `json:"coded_width,omitempty"`
	Refs               int    `json:"refs,omitempty"`
	Level              int    `json:"level,omitempty"`
	HasBFrames         int    `json:"has_b_frames,omitempty"`
	StartPts           int    `json:"start_pts"`
	FilmGrain          int    `json:"film_grain,omitempty"`
	ExtradataSize      int    `json:"extradata_size"`
	ClosedCaptions     int    `json:"closed_captions,omitempty"`
	CodedHeight        int    `json:"coded_height,omitempty"`
	Index              int    `json:"index"`
	Height             int    `json:"height,omitempty"`
	Channels           int    `json:"channels,omitempty"`
	Width              int    `json:"width,omitempty"`
	BitsPerSample      int    `json:"bits_per_sample,omitempty"`
	InitialPadding     int    `json:"initial_padding,omitempty"`
}

type Format struct {
	Tags struct {
		Title            string `json:"title"`
		COMMENT          string `json:"COMMENT"`
		MAJORBRAND       string `json:"MAJOR_BRAND"`
		MINORVERSION     string `json:"MINOR_VERSION"`
		COMPATIBLEBRANDS string `json:"COMPATIBLE_BRANDS"`
		ENCODER          string `json:"ENCODER"`
	} `json:"tags"`
	Filename       string `json:"filename"`
	FormatName     string `json:"format_name"`
	FormatLongName string `json:"format_long_name"`
	StartTime      string `json:"start_time"`
	Duration       string `json:"duration"`
	Size           string `json:"size"`
	BitRate        string `json:"bit_rate"`
	NbStreams      int    `json:"nb_streams"`
	NbPrograms     int    `json:"nb_programs"`
	ProbeScore     int    `json:"probe_score"`
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

func (p Probe) BitsPerSample() int {
	for i := range p.Streams {
		if p.Streams[i].CodecType == "video" {
			return p.Streams[i].BitsPerSample
		}
	}
	return 0
}

func (p Probe) LogValue() slog.Value {
	return slog.GroupValue(
		slog.String("codec", p.VideoCodec()),
		slog.Int("bitrate", p.BitRate()/1024),
		slog.Int("depth", p.BitsPerSample()),
		slog.Int("height", p.Height()),
		slog.Duration("duration", p.Duration()),
	)
}
