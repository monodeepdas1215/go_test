package logger

import (
	"github.com/monodeepdas1215/go_test/core/config"
	"github.com/sirupsen/logrus"
	"os"
)

var Logger *logrus.Logger

func init() {

	Logger = &logrus.Logger{
		Out:          os.Stdout,
		Formatter:    new(logrus.TextFormatter),
		Hooks:        make(logrus.LevelHooks),
		Level:        config.AppLogLevel,
		ExitFunc:     os.Exit,
		ReportCaller: false,
	}
}
