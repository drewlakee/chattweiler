package logging

import (
	"chattweiler/internal/configs"
	"chattweiler/internal/utils"
	"fmt"
	"github.com/sirupsen/logrus"
	"os"
	"strconv"
)

var Log = NewEdgeLogger()

type EdgeLogger struct {
	log *logrus.Logger
}

func NewEdgeLogger() *EdgeLogger {
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})

	logToFile, err := strconv.ParseBool(utils.GetEnvOrDefault(configs.BotLogToFile))
	if err != nil {
		panic(err)
	}

	if logToFile {
		_ = os.Remove("logs")
		err = os.Mkdir("logs", 0755)
		filename := "logs/chattweiler.log"
		_ = os.Remove(filename)
		outputFile, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0666)

		if err != nil {
			panic(err)
		}

		logger.SetOutput(outputFile)
	} else {
		logger.SetOutput(os.Stdout)
	}

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
