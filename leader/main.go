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
	log.Info().Msg("QuillSecure Leader booting...")

	net, err := NewLeaderNet(viper.GetString("leaderHost"), viper.GetInt("leaderPort"))
	if err != nil {
		log.Fatal().Err(err).Msg("Error initializing listener")
	}
	closeHandler(net)

	log.Info().Msg("QuillSecure Leader booted")

	if err := net.StartListening(); err != nil {
		net.Close()
		log.Fatal().Err(err).Msg("Error in listener")
	}
}

func closeHandler(net *LeaderNet) {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		net.Close()
		log.Info().Msg("QuillSecure Leader shutting down due to interrupt")

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
		viper.AddConfigPath("leader")
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
