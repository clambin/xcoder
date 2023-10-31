package video

import (
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
	var components []string
	if v.IsSeries {
		components = []string{v.Name, v.Episode, v.Extension}
	} else {
		components = []string{v.Name, v.Extension}
	}
	return strings.Join(components, ".")
}

func (v VideoInfo) IsVideo() bool {
	return v.Extension == "mkv" || v.Extension == "mp4" || v.Extension == "avi"
}

func ParseVideoFilename(filename string) (VideoInfo, bool) {
	if info, ok := parseEpisode(filename); ok {
		return info, true
	}
	if info, ok := parseMovie(filename); ok {
		return info, true
	}
	return parseGeneric(filename)
}

var regexpSeries = regexp.MustCompile(`^(.+)\.([Ss][0-9]+([Ee][0-9]+)+).*\.(.+?)$`)

func parseEpisode(filename string) (VideoInfo, bool) {
	match := regexpSeries.FindStringSubmatch(filename)
	if len(match) == 0 {
		return VideoInfo{}, false
	}
	return VideoInfo{
		Name:      match[1],
		Extension: strings.ToLower(match[len(match)-1]),
		IsSeries:  true,
		Episode:   match[2],
	}, true
}

var regexpMovie = []*regexp.Regexp{
	regexp.MustCompile(`(.+?\.\d{4}).*\.(.+?)$`),
	regexp.MustCompile(`(.+? \d{4}) .+\.(.+?)$`),
	regexp.MustCompile(`(.+? \(?\d{4}\)?).+?\.(.+?)$`),
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

var regexpGeneric = regexp.MustCompile(`^(.+)\.(.+?)$`)

func parseGeneric(filename string) (VideoInfo, bool) {
	match := regexpGeneric.FindStringSubmatch(filename)
	if len(match) == 0 {
		return VideoInfo{}, false
	}
	return VideoInfo{Name: match[1], Extension: match[2]}, true
}
