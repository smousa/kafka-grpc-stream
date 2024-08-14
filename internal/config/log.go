package config

import (
	"time"

	"github.com/rs/zerolog"
	"github.com/spf13/viper"
)

//nolint:cyclop
func SetupLogging() {
	switch viper.GetString("log.level") {
	case "panic":
		zerolog.SetGlobalLevel(zerolog.PanicLevel)
	case "fatal":
		zerolog.SetGlobalLevel(zerolog.FatalLevel)
	case "error":
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	case "warn":
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	case "info":
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	case "debug":
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	case "trace":
		zerolog.SetGlobalLevel(zerolog.TraceLevel)
	default:
		panic("invalid log level")
	}

	switch viper.GetString("log.timeFieldFormat") {
	case "rfc3339":
		zerolog.TimeFieldFormat = time.RFC3339
	case "unix":
		zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	default:
		panic("invalid time field format")
	}
}
