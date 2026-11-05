package config

import (
	"strconv"
	"strings"

	"github.com/fastenhealth/fasten-onprem/backend/pkg/errors"
)

// An omitted setting retains public registration. Empty databases can bootstrap
// one local administrator independently of this subsequent-registration policy.
func PublicRegistrationEnabled(cfg Interface) (bool, error) {
	value := strings.TrimSpace(cfg.GetString("auth.registration.enabled"))
	if value == "" {
		return true, nil
	}
	enabled, err := strconv.ParseBool(value)
	if err != nil {
		return false, errors.ConfigValidationError("auth.registration.enabled must be a boolean")
	}
	return enabled, nil
}
