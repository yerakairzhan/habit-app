// Package logger provides a zerolog-backed structured logger.
package logger

import (
	"os"
	"time"

	"github.com/rs/zerolog"
)

// New returns a zerolog.Logger configured for the given environment.
// In development the output is human-readable console format.
// In production it emits structured JSON.
func New(env string) zerolog.Logger {
	zerolog.TimeFieldFormat = time.RFC3339

	if env == "development" {
		return zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.Kitchen}).
			Level(zerolog.DebugLevel).
			With().
			Timestamp().
			Caller().
			Logger()
	}

	return zerolog.New(os.Stdout).
		Level(zerolog.InfoLevel).
		With().
		Timestamp().
		Logger()
}
