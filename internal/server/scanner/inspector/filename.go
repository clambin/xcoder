package inspector

import (
	"path/filepath"
	"regexp"
	"strings"
)

func makeTargetFilename(source, directory, codec, extension string) string {
	if directory == "" {
		directory = filepath.Dir(source)
	}
	return filepath.Join(
		directory,
		strings.Join([]string{getBasename(source), codec, extension}, "."),
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

func parseEpisode(filename string) (string, bool) {
	match := regexpSeries.FindStringSubmatch(filename)
	if len(match) == 0 {
		return "", false
	}
	return match[1] + "." + strings.ToLower(match[2]), true
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
