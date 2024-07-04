package common

import (
	"os"
	"regexp"

	log "github.com/sirupsen/logrus"
)

const (
	levelWarning = "warning"
	levelError   = "error"
	levelFatal   = "fatal"
	levelPanic   = "panic"
)

var (
	levelRex *regexp.Regexp
)

func init() {
	var err error
	levelRex, err = regexp.Compile("level=([a-z]+)")
	if err != nil {
		log.WithError(err).Fatal("unable to setup log level")
	}
}

// StdLogWriter is a custom log writer.
type StdLogWriter struct{}

func (w *StdLogWriter) Write(p []byte) (int, error) {
	var level string

	matches := levelRex.FindStringSubmatch(string(p))
	if len(matches) > 1 {
		level = matches[1]
	}

	if level == levelWarning || level == levelError || level == levelFatal || level == levelPanic {
		return os.Stderr.Write(p)
	}
	return os.Stdout.Write(p)
}

// ExtraLogVars contains extra log variables that are added to every log message.
type ExtraLogVars struct {
	service string
	env     string
}

// NewExtraLogVars return a new ExtraLogVars struct.
func NewExtraLogVars(service string, env string) *ExtraLogVars {
	return &ExtraLogVars{
		service: service,
		env:     env,
	}
}

// Levels returns log levels that will contain the extra log variables.
func (e *ExtraLogVars) Levels() []log.Level {
	return log.AllLevels
}

// Fire sets the extra log values in the log message.
func (e *ExtraLogVars) Fire(entry *log.Entry) error {
	entry.Data["service"] = e.service
	entry.Data["env"] = e.env
	return nil
}
