package ffmpeg

import (
	"log/slog"
	"strconv"
	"time"
)

type VideoStats struct {
	Streams []Stream `json:"streams"`
	Format  Format   `json:"format"`
}

type Stream struct {
	Index              int    `json:"index"`
	CodecName          string `json:"codec_name"`
	CodecLongName      string `json:"codec_long_name"`
	Profile            string `json:"profile,omitempty"`
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
	ChromaLocation     string `json:"chroma_location,omitempty"`
	FieldOrder         string `json:"field_order,omitempty"`
	Refs               int    `json:"refs,omitempty"`
	IsAvc              string `json:"is_avc,omitempty"`
	NalLengthSize      string `json:"nal_length_size,omitempty"`
	Id                 string `json:"id"`
	RFrameRate         string `json:"r_frame_rate"`
	AvgFrameRate       string `json:"avg_frame_rate"`
	TimeBase           string `json:"time_base"`
	StartPts           int    `json:"start_pts"`
	StartTime          string `json:"start_time"`
	DurationTs         int    `json:"duration_ts"`
	Duration           string `json:"duration"`
	BitRate            string `json:"bit_rate,omitempty"`
	BitsPerRawSample   string `json:"bits_per_raw_sample,omitempty"`
	NbFrames           string `json:"nb_frames,omitempty"`
	ExtradataSize      int    `json:"extradata_size,omitempty"`
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
		NonDiegetic     int `json:"non_diegetic"`
		Captions        int `json:"captions"`
		Descriptions    int `json:"descriptions"`
		Metadata        int `json:"metadata"`
		Dependent       int `json:"dependent"`
		StillImage      int `json:"still_image"`
		Multilayer      int `json:"multilayer"`
	} `json:"disposition"`
	Tags struct {
		CreationTime string `json:"creation_time"`
		Language     string `json:"language"`
		VendorId     string `json:"vendor_id,omitempty"`
		HandlerName  string `json:"handler_name,omitempty"`
	} `json:"tags,omitempty"`
	SampleFmt      string `json:"sample_fmt,omitempty"`
	SampleRate     string `json:"sample_rate,omitempty"`
	Channels       int    `json:"channels,omitempty"`
	ChannelLayout  string `json:"channel_layout,omitempty"`
	BitsPerSample  int    `json:"bits_per_sample,omitempty"`
	InitialPadding int    `json:"initial_padding,omitempty"`
	SideDataList   []struct {
		SideDataType string `json:"side_data_type"`
		ServiceType  int    `json:"service_type"`
	} `json:"side_data_list,omitempty"`
	ColorRange string `json:"color_range,omitempty"`
	ColorSpace string `json:"color_space,omitempty"`
}

type Format struct {
	Filename       string `json:"filename"`
	NbStreams      int    `json:"nb_streams"`
	NbPrograms     int    `json:"nb_programs"`
	NbStreamGroups int    `json:"nb_stream_groups"`
	FormatName     string `json:"format_name"`
	FormatLongName string `json:"format_long_name"`
	StartTime      string `json:"start_time"`
	Duration       string `json:"duration"`
	Size           string `json:"size"`
	BitRate        string `json:"bit_rate"`
	ProbeScore     int    `json:"probe_score"`
	Tags           struct {
		MajorBrand       string `json:"major_brand"`
		MinorVersion     string `json:"minor_version"`
		CompatibleBrands string `json:"compatible_brands"`
		CreationTime     string `json:"creation_time"`
		ITunEXTC         string `json:"iTunEXTC"`
		ITunMOVI         string `json:"iTunMOVI"`
		Title            string `json:"title"`
		Artist           string `json:"artist"`
		Genre            string `json:"genre"`
		Date             string `json:"date"`
		Synopsis         string `json:"synopsis"`
		HdVideo          string `json:"hd_video"`
		MediaType        string `json:"media_type"`
	} `json:"tags"`
}

func NewVideoStats(codec string, height int, rate int) VideoStats {
	return VideoStats{
		Format: Format{BitRate: strconv.Itoa(rate)},
		Streams: []Stream{
			{CodecType: "video", CodecName: codec, Height: height},
		},
	}
}

func (s VideoStats) String() string {
	output := s.VideoCodec()
	if output == "" {
		return ""
	}
	if height := s.Height(); height > 0 {
		output += "/" + strconv.Itoa(height)
	}
	if bitRate := s.BitRate(); bitRate > 0 {
		output += "/" + Bits(bitRate).Format(2)
	}
	return output
}

func (s VideoStats) VideoCodec() string {
	for i := range s.Streams {
		if s.Streams[i].CodecType == "video" {
			return s.Streams[i].CodecName
		}
	}
	return ""
}

func (s VideoStats) BitRate() int {
	value, _ := strconv.Atoi(s.Format.BitRate)
	return value
}

func (s VideoStats) Height() int {
	for i := range s.Streams {
		if s.Streams[i].CodecType == "video" {
			return s.Streams[i].Height
		}
	}
	return 0
}

func (s VideoStats) Width() any {
	for i := range s.Streams {
		if s.Streams[i].CodecType == "video" {
			return s.Streams[i].Width
		}
	}
	return 0
}

func (s VideoStats) Duration() time.Duration {
	seconds, _ := strconv.ParseFloat(s.Format.Duration, 64)
	return time.Duration(seconds) * time.Second
}

func (s VideoStats) BitsPerSample() int {
	var bitsPerSample int
	for i := range s.Streams {
		if s.Streams[i].CodecType == "video" {
			bitsPerSample, _ = strconv.Atoi(s.Streams[i].BitsPerRawSample)
			if bitsPerSample == 0 {
				bitsPerSample = 8
			}
			return bitsPerSample
		}
	}
	return 0
}

func (s VideoStats) LogValue() slog.Value {
	return slog.GroupValue(
		slog.String("codec", s.VideoCodec()),
		slog.Int("bitrate", s.BitRate()/1024),
		slog.Int("depth", s.BitsPerSample()),
		slog.Int("height", s.Height()),
		slog.Duration("duration", s.Duration()),
	)
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
