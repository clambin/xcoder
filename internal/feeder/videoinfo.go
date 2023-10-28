package feeder

import (
	"fmt"
	"regexp"
	"strings"
)

type VideoInfo struct {
	Name      string
	Extension string
	IsSeries  bool
	Episode   string
}

func (v VideoInfo) String() string {
	if v.IsSeries {
		return fmt.Sprintf("%s.%s.%s", v.Name, v.Episode, v.Extension)
	}
	return v.Name + "." + v.Extension
}

func (v VideoInfo) IsVideo() bool {
	return v.Extension == "mkv" || v.Extension == "mp4" || v.Extension == "avi"
}

var (
	regexpSeries = regexp.MustCompile(`^(.+)\.([Ss][0-9]+[Ee][0-9]+).*\.(.+?)$`)
	regexpMovie  = []*regexp.Regexp{
		regexp.MustCompile(`(.+\.\d{4}).*\.(.+?)$`),
		regexp.MustCompile(`(.+ \d{4}) .+\.(.+?)$`),
		regexp.MustCompile(`(.+ \(?\d{4}\)?).+?\.(.+?)$`),
	}
	regexpGeneric = regexp.MustCompile(`^(.+)\.(.+?)$`)
)

func parseVideoInfo(filename string) (VideoInfo, bool) {
	if info, ok := parseEpisode(filename); ok {
		return info, true
	}
	if info, ok := parseMovie(filename); ok {
		return info, true
	}
	return parseGeneric(filename)
}

func parseEpisode(filename string) (VideoInfo, bool) {
	match := regexpSeries.FindStringSubmatch(filename)
	if len(match) == 0 {
		return VideoInfo{}, false
	}
	return VideoInfo{
		Name:      match[1],
		Extension: strings.ToLower(match[3]),
		IsSeries:  true,
		Episode:   match[2],
	}, true
}

func parseMovie(filename string) (VideoInfo, bool) {
	for _, re := range regexpMovie {
		match := re.FindStringSubmatch(filename)
		if len(match) != 0 {
			return VideoInfo{
				Name:      match[1],
				Extension: match[2],
			}, true
		}
	}
	return VideoInfo{}, false
}

func parseGeneric(filename string) (VideoInfo, bool) {
	match := regexpGeneric.FindStringSubmatch(filename)
	if len(match) == 0 {
		return VideoInfo{}, false
	}
	return VideoInfo{Name: match[1], Extension: match[2]}, true
}
