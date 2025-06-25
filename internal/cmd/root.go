package cmd

import (
	"runtime/debug"

	"codeberg.org/clambin/go-common/charmer"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	rootCmd = cobra.Command{
		Use:   "xcoder",
		Short: "Transcode media files",
		PreRun: func(cmd *cobra.Command, args []string) {
			charmer.SetTextLogger(cmd, viper.GetBool("log.debug"))
		},
		Run: func(cmd *cobra.Command, args []string) {

		},
	}
)

func Execute() error {
	return rootCmd.Execute()
}

var arguments charmer.Arguments

func init() {
	buildInfo, ok := debug.ReadBuildInfo()
	if !ok {
		panic("could not read build info")
	}
	rootCmd.Version = buildInfo.Main.Version

	// cobra.OnInitialize(initConfig)
	// rootCmd.Flags().StringVar(&configFilename, "config", "", "Configuration file")
	_ = charmer.SetPersistentFlags(&rootCmd, viper.GetViper(), arguments)
	_ = charmer.SetDefaults(viper.GetViper(), arguments)
}

/*
func initConfig() {
	if configFilename != "" {
		viper.SetConfigFile(configFilename)
	} else {
		viper.AddConfigPath("/etc/mediamon/")
		viper.AddConfigPath("$HOME/.mediamon")
		viper.AddConfigPath(".")
		viper.SetConfigName("config")
	}

	viper.SetEnvPrefix("XCODE")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		slog.Error("failed to read config file", "err", err)
	}
}
*/
