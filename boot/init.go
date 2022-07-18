package boot

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"os"
)

func SetGlobalLogger() {
	consoleWriter := zerolog.ConsoleWriter{Out: os.Stderr}
	if viper.GetBool("prettyLogging") {
		log.Logger = log.Output(consoleWriter)
	}

	logFile := viper.GetString("logFile")
	if logFile != "" {
		file, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0755)
		if err != nil {
			log.Fatal().Str("logFile", logFile).Msg("Failed to open log file")
		}

		multi := zerolog.MultiLevelWriter(consoleWriter, file)
		log.Logger = log.Output(multi)
	}
}
