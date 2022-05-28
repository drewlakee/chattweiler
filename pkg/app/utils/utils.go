package utils

import (
	"chattweiler/pkg/app/configs"
	"github.com/sirupsen/logrus"
	"os"
	"strings"
)

var packageLogFields = logrus.Fields{
	"package": "utils",
}

func iterateOverAllPossibleKeys(key string) string {
	value := os.Getenv(key)
	if len(value) != 0 {
		return value
	}

	// my.env.var -> my_env_var
	value = os.Getenv(strings.ReplaceAll(key, ".", "_"))
	if len(value) != 0 {
		return value
	}

	return ""
}

func GetEnvOrDefault(config configs.ApplicationConfig) string {
	value := iterateOverAllPossibleKeys(config.GetKey())
	if len(value) == 0 {
		return config.GetDefaultValue()
	}
	return value
}

func MustGetEnv(config configs.ApplicationConfig) string {
	value := iterateOverAllPossibleKeys(config.GetKey())
	if len(value) == 0 {
		logrus.WithFields(packageLogFields).WithFields(logrus.Fields{
			"func": "MustGetEnv",
			"key":  config.GetKey(),
		}).Fatal("Couldn't get environment variable's value because there is no such variable")
	}
	return value
}
