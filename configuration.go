package main

import (
	"fmt"
	"os"
)

type Configuration struct {
	Port string
}

func LoadConfiguration() Configuration {
	return Configuration{
		Port: loadEnv("PORT", "10000"),
	}
}

func loadEnv(key string, defaultValue string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		fmt.Printf(fmt.Sprintf("Env variable '%s' not set. Use default value '%s'\n", key, defaultValue))
		return defaultValue
	}
	return value
}
