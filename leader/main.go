package main

import (
	"encoding/gob"
	"fmt"
	"github.com/Heanthor/quill-secure/boot"
	"github.com/Heanthor/quill-secure/node/sensor"
	"github.com/mitchellh/go-homedir"
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
	gob.Register(sensor.Data{})

	db, err := NewDB(viper.GetString("dbFile"))
	if err != nil {
		log.Fatal().Err(err).Msg("Error initializing database")
	}
	log.Info().Msg("Database initialized")

	net, err := NewLeaderNet(viper.GetString("leaderHost"),
		viper.GetInt("leaderPort"),
		viper.GetInt("nodePingTimeoutSecs"),
		db,
	)
	if err != nil {
		log.Fatal().Err(err).Msg("Error initializing listener")
	}
	registerCloseHandler(net, db)

	log.Info().Msg("QuillSecure Leader booted")

	if err := net.StartListening(); err != nil {
		net.Close()
		log.Fatal().Err(err).Msg("Error in listener")
	}
}

func registerCloseHandler(net *LeaderNet, d *DB) {
	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-c
		net.Close()
		d.Close()
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

		// Search config in home directory with name "quillsecure_leader.yaml"
		viper.AddConfigPath(home)
		viper.AddConfigPath(".")
		viper.AddConfigPath("leader")
		viper.SetConfigName("quillsecure_leader")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err != nil {
		log.Fatal().Msg("No config file found")
	}

	boot.SetGlobalLogger()
}
