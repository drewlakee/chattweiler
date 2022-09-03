// Package utils provides useful functions for
// the application needs
package utils

import (
	"chattweiler/pkg/configs"
	"fmt"
	"os"
	"strings"
)

func iterateOverAllPossibleKeys(key string) string {
	value := os.Getenv(key)
	if len(value) != 0 {
		return value
	}

	// my.env.var <-> my_env_var
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
		panic(fmt.Sprintf("%s: variable must be specified", config.GetKey()))
	}
	return value
}
