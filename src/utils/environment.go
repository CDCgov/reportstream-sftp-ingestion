package utils

import "os"

func EnvironmentName() string {
	environment := os.Getenv("ENV")
	if environment == "" {
		environment = "local"
	}

	return environment
}
