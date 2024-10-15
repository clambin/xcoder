package configuration

import (
	"flag"
	"github.com/clambin/videoConvertor/internal/profile"
)

var (
	debug        = flag.Bool("debug", false, "switch on debug logging")
	input        = flag.String("input", "/media", "input directory")
	videoProfile = flag.String("profile", "hevc-high", "conversion profile")
	active       = flag.Bool("active", false, "start convertor in active mode")
	remove       = flag.Bool("remove", false, "remove source files after successful conversion")
	overwrite    = flag.Bool("overwrite", false, "overwrite existing files")
)

type Configuration struct {
	Input                string
	ProfileName          string
	Profile              profile.Profile
	Debug                bool
	Active               bool
	RemoveSource         bool
	OverwriteNewerTarget bool
}

func GetConfiguration() (Configuration, error) {
	flag.Parse()

	configuration := Configuration{
		Input:                *input,
		ProfileName:          *videoProfile,
		Debug:                *debug,
		Active:               *active,
		RemoveSource:         *remove,
		OverwriteNewerTarget: *overwrite,
	}

	var err error
	if configuration.Profile, err = profile.GetProfile(configuration.ProfileName); err != nil {
		return Configuration{}, err
	}

	return configuration, nil
}
