package main

import (
	"encoding/gob"
	"fmt"
	"github.com/Heanthor/quill-secure/boot"
	"github.com/Heanthor/quill-secure/db"
	"github.com/Heanthor/quill-secure/leader/api"
	"github.com/Heanthor/quill-secure/leader/net"
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

func main() {
	initConfig()

	env := viper.GetString("env")
	if env == "" {
		panic("missing 'env' config")
	}
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

	log.Info().Str("env", env).Msg("QuillSecure Leader booting...")
	gob.Register(sensor.Data{})

	d, err := db.NewDB(viper.GetString("dbFile"))
	if err != nil {
		log.Fatal().Err(err).Msg("Error initializing database")
	}
	log.Info().Msg("Database initialized")

	n, err := net.NewLeaderNet(viper.GetInt("leaderPort"),
		viper.GetInt("nodePingTimeoutSecs"),
		d,
	)
	if err != nil {
		log.Fatal().Err(err).Msg("Error initializing listener")
	}

	a := api.NewRouter(env, d, n.ActiveNodesFunc(), viper.GetInt("api.dashboardStatsDays"))
	go func() {
		port := viper.GetInt("api.port")
		log.Info().Int("port", port).Msg("API initialized")
		if err := a.Listen(port); err != nil {
			log.Fatal().Err(err).Msg("Error in API")
		}
	}()

	registerCloseHandler(n, d)

	log.Info().Msg("QuillSecure Leader booted")

	if err := n.StartListening(); err != nil {
		n.Close()
		log.Fatal().Err(err).Msg("Error in listener")
	}
}

func registerCloseHandler(net *net.LeaderNet, d *db.DB) {
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
		viper.AddConfigPath("/usr/local/bin/quillsecure")
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
