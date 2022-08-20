package main

import (
	"encoding/gob"
	"flag"
	"fmt"
	"github.com/Heanthor/quill-secure/boot"
	"github.com/Heanthor/quill-secure/node/sensor"
	"github.com/mitchellh/go-homedir"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

var (
	cfgFile string
)

var (
	overrideDeviceID int
)

func init() {
	flag.IntVar(&overrideDeviceID, "deviceID", 0, "Override device ID present in config")
}

func main() {
	flag.Parse()

	initConfig()
	deviceID := uint8(viper.GetInt("deviceID"))
	if overrideDeviceID != 0 {
		deviceID = uint8(overrideDeviceID)
	}
	// inject deviceID into every log
	log.Logger = log.With().Uint8("deviceID", deviceID).Logger()

	logLevelStr := viper.GetString("logLevel")
	if logLevelStr == "" {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
		log.Info().Msg("Setting global log level to info")
	} else {
		level, err := zerolog.ParseLevel(strings.ToLower(logLevelStr))
		if err != nil {
			log.Fatal().Str("level", logLevelStr).Msg("invalid log level")
		}
		log.Info().Msgf("Setting global log level to %s", level.String())
		zerolog.SetGlobalLevel(level)
	}

	log.Info().Msg("QuillSecure Node booting...")
	gob.Register(sensor.Data{})

	sc := NewSensorCollection(deviceID,
		viper.GetString("leaderHost"),
		viper.GetInt("leaderPort"),
		viper.GetInt("pingIntervalSecs"),
		viper.GetInt("packetBufferSize"))
	setCloseHandler(&sc)

	// find and activate all sensor connected to device
	sc.RegisterSensors(
		viper.GetString("sensors.atmospheric.executable"),
		viper.GetInt("sensors.atmospheric.pollFrequencySec"),
	)
	// start periodic health ping to leader
	sc.StartLeaderHealthCheck()

	log.Info().Msg("QuillSecure Node booted")

	// block and poll connected sensors
	sc.Poll()
}

func setCloseHandler(sc *SensorCollection) {
	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-c
		log.Info().Msg("QuillSecure Node shutting down due to interrupt")
		sc.StopPolling()
		// drain outgoing packets before exiting
		// TODO should this have a timeout?
		<-sc.doneChan
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

		// Search config in home directory with name "quillsecure_leader.yaml"
		viper.AddConfigPath(home)
		viper.AddConfigPath(".")
		viper.AddConfigPath("node")
		viper.AddConfigPath("/usr/local/bin/quillsecure")
		viper.SetConfigName("quillsecure_node")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err != nil {
		log.Fatal().Msg("No config file found")
	}

	boot.SetGlobalLogger()
}
