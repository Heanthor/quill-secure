package main

import (
	"encoding/gob"
	"flag"
	"fmt"
	"github.com/Heanthor/quill-secure/node/sensor"
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

	log.Info().Msg("QuillSecure Node booting...")
	gob.Register(sensor.Data{})

	sc := NewSensorCollection(deviceID,
		viper.GetString("leaderHost"),
		viper.GetInt("leaderPort"),
		viper.GetInt("pingIntervalSecs"),
		viper.GetInt("packetBufferSize"))
	setCloseHandler(&sc)

	// find and activate all sensor connected to device
	sc.RegisterSensors()
	// start periodic health ping to leader
	sc.StartLeaderHealthCheck()

	log.Info().Msg("QuillSecure Node booted")

	// block and poll connected sensors
	sc.Poll()
}

func setCloseHandler(sc *SensorCollection) {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
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
