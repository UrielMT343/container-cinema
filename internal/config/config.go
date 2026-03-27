package config

import (
	"os"
	"strconv"
	"time"
)

var CacheTTLMinutes = getEnvAsDurationMinutes("CACHE_TTL_MINUTES", 15)

var HoldTicketTTLMinutes = getEnvAsDurationMinutes("HOLD_TICKET_TTL_MINUTES", 5)

func getEnvAsDurationMinutes(key string, defaultMinutes int) time.Duration {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return time.Duration(defaultMinutes) * time.Minute
	}

	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return time.Duration(defaultMinutes) * time.Minute
	}

	return time.Duration(value) * time.Minute
}
