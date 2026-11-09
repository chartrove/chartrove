package web

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"runtime"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/fastenhealth/fasten-onprem/backend/pkg/models"
)

const installationRegistrationTimeout = 10 * time.Second
const maxInstallationRegistrationResponse = 64 * 1024

var errOptionalInstallationRegistration = errors.New("optional installation registration unavailable")

// Keep the original local failure inspectable without formatting a possibly
// credential-bearing storage cause into startup logs or CLI output.
type installationSettingsError struct {
	message string
	cause   error
}

func (failure *installationSettingsError) Error() string { return failure.message }
func (failure *installationSettingsError) Unwrap() error { return failure.cause }

func (ae *AppEngine) SetupInstallationRegistration() error {
	return ae.setupInstallationRegistration(http.DefaultClient)
}

func (ae *AppEngine) setupInstallationRegistration(client *http.Client) error {
	settings, err := ae.deviceRepo.LoadSystemSettings(context.Background())
	if err != nil {
		return &installationSettingsError{message: "could not load installation settings", cause: err}
	}
	if settings == nil {
		return errors.New("installation settings unavailable")
	}
	if settings.InstallationID != "" && settings.InstallationSecret != "" {
		return nil
	}
	registration := &models.InstallationRegistrationRequest{SoftwareArchitecture: runtime.GOARCH, SoftwareOS: runtime.GOOS}
	if ae.RelatedVersions != nil {
		registration.FastenSourcesVersion = ae.RelatedVersions["sources"]
		registration.FastenOnpremVersion = ae.RelatedVersions["onprem"]
		registration.FastenDesktopVersion = ae.RelatedVersions["desktop"]
	}
	payload, err := json.Marshal(registration)
	if err != nil {
		return errors.New("could not encode installation registration")
	}
	ctx, cancel := context.WithTimeout(context.Background(), installationRegistrationTimeout)
	defer cancel()
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.platform.fastenhealth.com/v1/installation/register", bytes.NewReader(payload))
	if err != nil {
		return errors.New("could not prepare installation registration")
	}
	request.Header.Set("Content-Type", "application/json")
	bounded := *client
	if bounded.Timeout <= 0 || bounded.Timeout > installationRegistrationTimeout {
		bounded.Timeout = installationRegistrationTimeout
	}
	response, err := bounded.Do(request)
	if err != nil {
		if response != nil && response.Body != nil {
			_ = response.Body.Close()
		}
		return errOptionalInstallationRegistration
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("%w (HTTP %d)", errOptionalInstallationRegistration, response.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(response.Body, maxInstallationRegistrationResponse+1))
	if err != nil || len(body) > maxInstallationRegistrationResponse || !utf8.Valid(body) {
		return errOptionalInstallationRegistration
	}
	var result models.ResponseWrapperTyped[models.InstallationRegistrationResponse]
	if json.Unmarshal(body, &result) != nil || !result.Success || strings.TrimSpace(result.Data.InstallationID) == "" || len(result.Data.InstallationID) > 2048 || len(result.Data.InstallationSecret) > 16384 {
		return errOptionalInstallationRegistration
	}
	// The remote model explicitly permits an omitted secret. Preserve exact
	// accepted strings, rather than inventing/normalizing a credential.
	settings.InstallationID = result.Data.InstallationID
	settings.InstallationSecret = result.Data.InstallationSecret
	if err := ae.deviceRepo.SaveSystemSettings(context.Background(), settings); err != nil {
		return &installationSettingsError{message: "could not save installation settings", cause: err}
	}
	ae.Logger.Info("Saved installation registration")
	return nil
}
