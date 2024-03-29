package utils

import (
	"chattweiler/internal/configs"
	"fmt"
	"os"
)

func GetEnvOrDefault(config configs.ApplicationConfig) string {
	value := os.Getenv(config.GetKey())
	if len(value) == 0 {
		return config.GetDefaultValue()
	}
	return value
}

func MustGetEnv(config configs.ApplicationConfig) string {
	value := os.Getenv(config.GetKey())
	if len(value) == 0 {
		panic(fmt.Sprintf("%s: variable must be specified", config.GetKey()))
	}
	return value
}
