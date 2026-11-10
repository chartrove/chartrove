package models

import (
	"context"
	"errors"
	"io"
	"os"

	"github.com/fastenhealth/fasten-sources/clients/factory"
	sourceModels "github.com/fastenhealth/fasten-sources/clients/models"
	sourcePkg "github.com/fastenhealth/fasten-sources/pkg"
	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
)

// NewProviderSourceClient owns the configured SDK diagnostic boundary. Successful
// data/credentials/summaries remain unchanged; instance logs and returned operation
// errors cannot acquire raw SDK request/dump content. This does not intercept SDK
// standard-log calls, original parent logger fields or arbitrary successful data.
func NewProviderSourceClient(env sourcePkg.FastenLighthouseEnvType, ctx context.Context, logger logrus.FieldLogger, credential sourceModels.SourceCredential, options ...func(*sourceModels.SourceClientOptions)) (sourceModels.SourceClient, error) {
	ownedOptions := append([]func(*sourceModels.SourceClientOptions){}, options...)
	ownedOptions = append(ownedOptions, WithProviderResponsePrivacy(), WithProviderRequestLifetime())
	client, err := factory.GetSourceClient(env, ctx, providerDiagnosticLogger(logger), credential, ownedOptions...)
	if err != nil {
		return nil, privateProviderOperationError(err)
	}
	return &providerDiagnosticClient{SourceClient: client}, nil
}

// A separate logger also owns WithFields/WithError-derived entries, unlike a
// FieldLogger adapter whose returned *Entry could escape the filter. The pinned
// logrus Entry.Log forwards severity only; Fatal*/Panic* keep their control flow
// in this logger. No shared logger, hooks, formatter or writer is mutated.
func providerDiagnosticLogger(parent logrus.FieldLogger) *logrus.Logger {
	entry := parent.WithField("component", "provider_sdk")
	logger := logrus.New()
	logger.SetLevel(entry.Logger.GetLevel())
	logger.SetOutput(io.Discard)
	logger.SetFormatter(providerDiagnosticFormatter{})
	logger.ExitFunc = entry.Logger.ExitFunc
	logger.AddHook(providerDiagnosticHook{parent: entry})
	return logger
}

type providerDiagnosticFormatter struct{}

func (providerDiagnosticFormatter) Format(*logrus.Entry) ([]byte, error) {
	return []byte("provider SDK diagnostic\n"), nil
}

type providerDiagnosticHook struct{ parent *logrus.Entry }

func (providerDiagnosticHook) Levels() []logrus.Level { return logrus.AllLevels }
func (hook providerDiagnosticHook) Fire(entry *logrus.Entry) error {
	// Reset the whole entry: WithFields can also retain private field-error
	// text in logrus's unexported state, observable in a recovered panic.
	*entry = logrus.Entry{Logger: entry.Logger, Time: entry.Time, Level: entry.Level,
		Data: logrus.Fields{}, Message: "provider SDK diagnostic"}
	hook.parent.WithTime(entry.Time).Log(entry.Level, entry.Message)
	return nil
}

type providerOperationError struct{ cause error }

func privateProviderOperationError(err error) error {
	if err == nil {
		return nil
	}
	return &providerOperationError{cause: err}
}

func (err *providerOperationError) Error() string {
	if errors.Is(err.cause, context.Canceled) {
		return "provider operation canceled"
	}
	if errors.Is(err.cause, context.DeadlineExceeded) {
		return "provider operation deadline exceeded"
	}
	var rejection *oauth2.RetrieveError
	if errors.As(err.cause, &rejection) {
		if code := providerAllowedOAuthCode(rejection.ErrorCode); code != providerFailureCode {
			return "provider operation failed: " + code
		}
	}
	return "provider operation failed"
}
func (err *providerOperationError) Unwrap() error { return err.cause }

type providerDiagnosticClient struct{ sourceModels.SourceClient }

func (client *providerDiagnosticClient) GetRequest(path string, out interface{}) (string, error) {
	location, err := client.SourceClient.GetRequest(path, out)
	return location, privateProviderOperationError(err)
}
func (client *providerDiagnosticClient) GetResourceBundle(path string) (interface{}, error) {
	bundle, err := client.SourceClient.GetResourceBundle(path)
	return bundle, privateProviderOperationError(err)
}
func (client *providerDiagnosticClient) SyncAll(repository sourceModels.DatabaseRepository) (sourceModels.UpsertSummary, error) {
	summary, err := client.SourceClient.SyncAll(repository)
	return summary, privateProviderOperationError(err)
}
func (client *providerDiagnosticClient) SyncAllByResourceName(repository sourceModels.DatabaseRepository, names []string) (sourceModels.UpsertSummary, error) {
	summary, err := client.SourceClient.SyncAllByResourceName(repository, names)
	return summary, privateProviderOperationError(err)
}
func (client *providerDiagnosticClient) SyncAllByPatientEverythingBundle(repository sourceModels.DatabaseRepository, bundle interface{}) (sourceModels.UpsertSummary, error) {
	summary, err := client.SourceClient.SyncAllByPatientEverythingBundle(repository, bundle)
	return summary, privateProviderOperationError(err)
}
func (client *providerDiagnosticClient) SyncAllBundle(repository sourceModels.DatabaseRepository, file *os.File, version sourcePkg.FhirVersion) (sourceModels.UpsertSummary, error) {
	summary, err := client.SourceClient.SyncAllBundle(repository, file, version)
	return summary, privateProviderOperationError(err)
}
func (client *providerDiagnosticClient) ExtractPatientId(file *os.File) (string, sourcePkg.FhirVersion, error) {
	patient, version, err := client.SourceClient.ExtractPatientId(file)
	return patient, version, privateProviderOperationError(err)
}
func (client *providerDiagnosticClient) RefreshAccessToken(options ...func(*sourceModels.SourceClientRefreshOptions)) error {
	return privateProviderOperationError(client.SourceClient.RefreshAccessToken(options...))
}
