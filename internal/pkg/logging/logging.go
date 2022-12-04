package logging

import (
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type LogConfig struct {
	LogPretty bool   `arg:"--log-pretty,env:LOG_PRETTY" default:"false" help:"log pretty"`
	LogLevel  string `arg:"--log-level,env:LOG_LEVEL" default:"info" help:"log level" placeholder:"LEVEL"`
}

func MustInit(config LogConfig) {
	level, err := zerolog.ParseLevel(config.LogLevel)
	if err != nil {
		log.Panic().Err(err).Send()
	}

	if config.LogPretty {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}

	zerolog.TimeFieldFormat = "2006-01-02T15:04:05.000Z07:00"
	zerolog.TimestampFunc = func() time.Time { return time.Now() }

	log.Logger = log.Logger.Level(level).With().Timestamp().Caller().Logger()
}
