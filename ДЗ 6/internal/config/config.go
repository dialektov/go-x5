package config

import "os"

func Channel() string {
	if value := os.Getenv("NOTIFY_CHANNEL"); value != "" {
		return value
	}

	return "email"
}
