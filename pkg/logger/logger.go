package logger

import (
	"context"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github/mlyahmed.io/nominee/pkg/config"
	"github/mlyahmed.io/nominee/version"
)

type (
	loggerKey struct{}
)

const (
	// RFC3339NanoFixed is time.RFC3339Nano with nanoseconds padded using zeros to
	// ensure the formatted time is always the same number of characters.
	RFC3339NanoFixed = "2006-01-02T15:04:05.000000000Z07:00"

	// TextFormat represents the text logging format
	TextFormat = "text"

	// JSONFormat represents the JSON logging format
	JSONFormat = "json"
)

var (
	// G is an alias for GetLogger.
	//
	// We may want to define this locally to a package to get package tagged log
	// messages.
	G = GetLogger

	// S is an alias for the standard logger.
	S = logrus.NewEntry(logrus.StandardLogger()).WithFields(logrus.Fields{
		"buildDate":          version.Date,
		"buildPlatform":      version.Platform,
		"buildSimpleVersion": version.SimpleVersion,
		"buildGitVersion":    version.GitVersion,
		"buildGitCommit":     version.GitCommit,
		"buildImageVersion":  version.ImageVersion,
	})
)

func init() {
	config.SetDefault("NOMINEE_OBS_LOGS_FORMAT", TextFormat)
	format := config.GetStringOrPanic("NOMINEE_OBS_LOGS_FORMAT")
	switch format {
	case TextFormat:
		logrus.SetFormatter(&logrus.TextFormatter{
			TimestampFormat: RFC3339NanoFixed,
			FullTimestamp:   true,
		})
	case JSONFormat:
		logrus.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: RFC3339NanoFixed,
		})
	default:
		panic(errors.Errorf("unknown log format: %s", format))
	}
}

// WithLogger returns a new context with the provided logger. Use in
// combination with logger.WithField(s) for great effect.
func WithLogger(ctx context.Context, logger *logrus.Entry) context.Context {
	return context.WithValue(ctx, loggerKey{}, logger)
}

// GetLogger retrieves the current logger from the context. If no logger is
// available, the default logger is returned.
func GetLogger(ctx context.Context) *logrus.Entry {
	logger := ctx.Value(loggerKey{})

	if logger == nil {
		return S
	}

	return logger.(*logrus.Entry)
}
