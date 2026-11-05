// Package config reads the settings of a process from environment
// variables. Settings come only from the environment (issue #25).
package config

import (
	"fmt"
	"os"
)

// String returns the variable, or fallback when it is empty.
func String(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// Required returns the variable, or an error when it is empty.
func Required(key string) (string, error) {
	v := os.Getenv(key)
	if v == "" {
		return "", fmt.Errorf("environment variable %s is not set (see .env.example)", key)
	}
	return v, nil
}
