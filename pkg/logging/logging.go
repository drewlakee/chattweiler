package logging

import (
	"chattweiler/pkg/configs"
	"chattweiler/pkg/utils"
	"fmt"
	"github.com/sirupsen/logrus"
	"io"
	"os"
)

var Log = NewEdgeLogger()

type EdgeLogger struct {
	log *logrus.Logger
}

func NewEdgeLogger() *EdgeLogger {
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})

	var filename string
	if utils.GetEnvOrDefault(configs.BotLogFile) == "" {
		filename = "chattweiler.log"
	} else {
		filename = utils.GetEnvOrDefault(configs.BotLogFile)
	}

	_ = os.Remove(filename)

	var outputFile *os.File
	var err error
	if utils.GetEnvOrDefault(configs.BotLogFile) == "" {
		outputFile, err = os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0666)
	} else {
		outputFile, err = os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0666)
	}

	if err != nil {
		panic(err)
	}

	logger.SetOutput(io.MultiWriter(os.Stderr, outputFile))
	return &EdgeLogger{logger}
}

func (logger *EdgeLogger) Info(sourcePackage, sourceFunc, messageFormat string, args ...interface{}) {
	logger.log.WithFields(logrus.Fields{
		"package": sourcePackage,
		"func":    sourceFunc,
	}).Info(getMessage(args, messageFormat))
}

func (logger *EdgeLogger) Error(sourcePackage, sourceFunc string, err error, messageFormat string, args ...interface{}) {
	logger.log.WithFields(logrus.Fields{
		"package": sourcePackage,
		"func":    sourceFunc,
		"err":     err,
	}).Error(getMessage(args, messageFormat))
}

func (logger *EdgeLogger) Panic(sourcePackage, sourceFunc string, err error, messageFormat string, args ...interface{}) {
	logger.log.WithFields(logrus.Fields{
		"package": sourcePackage,
		"func":    sourceFunc,
		"err":     err,
	}).Panic(getMessage(args, messageFormat))
}

func (logger *EdgeLogger) Warn(sourcePackage, sourceFunc string, messageFormat string, args ...interface{}) {
	logger.log.WithFields(logrus.Fields{
		"package": sourcePackage,
		"func":    sourceFunc,
	}).Warn(getMessage(args, messageFormat))
}

func getMessage(args []interface{}, messageFormat string) string {
	var message string
	if len(args) == 0 {
		message = messageFormat
	} else {
		message = fmt.Sprintf(messageFormat, args...)
	}
	return message
}
