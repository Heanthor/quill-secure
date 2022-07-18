package boot

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"os"
	"time"
)

func SetGlobalLogger() {
	consoleWriter := zerolog.ConsoleWriter{Out: os.Stderr}
	if viper.GetBool("prettyLogging") {
		log.Logger = log.Output(consoleWriter)
	}

	logFileSuffix := viper.GetString("logFileSuffix")
	if logFileSuffix != "" {
		dateStamp := "quillsecure_" + time.Now().Format("2006_01_02") + "_" + logFileSuffix + ".log"

		file, err := os.OpenFile(dateStamp, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0755)
		if err != nil {
			log.Fatal().Str("logFile", dateStamp).Msg("Failed to open log file")
		}

		multi := zerolog.MultiLevelWriter(consoleWriter, file)
		log.Logger = log.Output(multi)
	}
}
