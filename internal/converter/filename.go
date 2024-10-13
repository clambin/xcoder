package converter

import (
	"github.com/clambin/videoConvertor/internal/worklist"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

func buildTargetFilename(item *worklist.WorkItem, directory, codec, extension string) string {
	if directory == "" {
		directory = filepath.Dir(item.Source)
	}
	elements := make([]string, 0, 5)
	elements = append(elements, getBasename(item.Source))
	if height := item.SourceVideoStats().Height; height > 0 {
		elements = append(elements, strconv.Itoa(height))
	}
	elements = append(elements, codec, extension)

	return filepath.Join(
		directory,
		strings.Join(elements, "."),
	)
}

func getBasename(source string) string {
	source = filepath.Base(source)
	if basename, ok := parseEpisode(source); ok {
		return basename
	}
	if basename, ok := parseMovie(source); ok {
		return basename
	}
	return parseGeneric(source)
}

var regexpSeries = regexp.MustCompile(`^(.+)\.([Ss][0-9]+([Ee][0-9]+)+).*\.(.+?)$`)
var regexpAltSeries = regexp.MustCompile(`^(.+) ([Ss][0-9]+([Ee][0-9]+)+).*\.(.+?)$`)

func parseEpisode(filename string) (string, bool) {
	if match := regexpSeries.FindStringSubmatch(filename); len(match) >= 3 {
		return match[1] + "." + strings.ToLower(match[2]), true
	}
	if match := regexpAltSeries.FindStringSubmatch(filename); len(match) >= 3 {
		return match[1] + "." + strings.ToLower(match[2]), true
	}
	return "", false
}

var regexpMovie = []*regexp.Regexp{
	regexp.MustCompile(`(.+?\.\d{4}).*\.(.+?)$`),
	regexp.MustCompile(`(.+? \d{4}) .+\.(.+?)$`),
	regexp.MustCompile(`(.+? \(?\d{4}\)?).+?\.(.+?)$`),
}

func parseMovie(filename string) (string, bool) {
	for _, re := range regexpMovie {
		match := re.FindStringSubmatch(filename)
		if len(match) != 0 {
			return match[1], true
		}
	}
	return "", false
}

func parseGeneric(filename string) string {
	return strings.TrimSuffix(filename, filepath.Ext(filename))
}
