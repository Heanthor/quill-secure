package main

import (
	"fmt"
	"github.com/mitchellh/go-homedir"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"os"
	"os/signal"
	"syscall"
)

var (
	cfgFile string
)

func main() {
	initConfig()
	deviceID := uint8(viper.GetInt("deviceID"))
	log.Info().Uint8("deviceID", deviceID).Msg("QuillSecure Node booting...")
	closeHandler()

	sc := SensorCollection{
		deviceID: deviceID,
	}
	// find and activate all sensor connected to device
	sc.RegisterSensors()
	log.Info().Msg("QuillSecure Node booted")

	sc.Poll()
}

func closeHandler() {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		log.Info().Msg("QuillSecure Node shutting down due to interrupt")

		os.Exit(0)
	}()
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name "quillsecure.yaml"
		viper.AddConfigPath(home)
		viper.AddConfigPath(".")
		viper.AddConfigPath("node")
		viper.SetConfigName("quillsecure")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err != nil {
		log.Fatal().Msg("No config file found")
	}

	if viper.GetBool("prettyLogging") {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}
}
