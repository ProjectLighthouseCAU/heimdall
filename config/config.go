package config

import (
	"log"
	"os"
	"strconv"
)

func GetString(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func GetInt(key string, defaultValue int) int {
	if value, exists := os.LookupEnv(key); exists {
		s, err := strconv.Atoi(value)
		if err != nil {
			log.Printf("Found Config %s=%s, but could not parse it (int required)", key, value)
			return defaultValue
		}
		return s
	}
	return defaultValue
}

func GetBool(key string, defaultValue bool) bool {
	if value, exists := os.LookupEnv(key); exists {
		s, err := strconv.ParseBool(value)
		if err != nil {
			log.Printf("Found Config %s=%s, but could not parse it (bool required)", key, value)
			return defaultValue
		}
		return s
	}
	return defaultValue
}
