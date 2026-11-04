package config

import (
	"strings"
	"time"

	"github.com/fastenhealth/fasten-onprem/backend/pkg/errors"
)

// UserSessionDuration is the absolute lifetime for newly issued sign-in/sign-up
// tokens. An omitted/empty setting retains the historical one-hour lifetime.
func UserSessionDuration(cfg Interface) (time.Duration, error) {
	value := strings.TrimSpace(cfg.GetString("jwt.session.duration"))
	if value == "" {
		return time.Hour, nil
	}
	duration, err := time.ParseDuration(value)
	if err != nil || duration < time.Second {
		return 0, errors.ConfigValidationError("jwt.session.duration must be a duration of at least 1s")
	}
	return duration, nil
}
