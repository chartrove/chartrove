package models

import (
	"errors"
	"os"
	"testing"

	sourceModels "github.com/fastenhealth/fasten-sources/clients/models"
	sourcePkg "github.com/fastenhealth/fasten-sources/pkg"
	"github.com/stretchr/testify/require"
)

type providerDiagnosticForwardProbe struct {
	sourceModels.SourceClient
	t       *testing.T
	cause   error
	calls   []string
	out     any
	bundle  any
	file    *os.File
	version sourcePkg.FhirVersion
	summary sourceModels.UpsertSummary
	option  bool
}

func (probe *providerDiagnosticForwardProbe) GetRequest(path string, out interface{}) (string, error) {
	probe.calls = append(probe.calls, "request")
	require.Equal(probe.t, "synthetic-original-path", path)
	require.Same(probe.t, probe.out, out)
	return "synthetic-original-location", probe.cause
}
func (probe *providerDiagnosticForwardProbe) GetResourceBundle(path string) (interface{}, error) {
	probe.calls = append(probe.calls, "bundle")
	require.Equal(probe.t, "synthetic-original-path", path)
	return probe.bundle, probe.cause
}
func (probe *providerDiagnosticForwardProbe) SyncAll(repository sourceModels.DatabaseRepository) (sourceModels.UpsertSummary, error) {
	probe.calls = append(probe.calls, "sync")
	require.Nil(probe.t, repository)
	return probe.summary, probe.cause
}
func (probe *providerDiagnosticForwardProbe) SyncAllByResourceName(repository sourceModels.DatabaseRepository, names []string) (sourceModels.UpsertSummary, error) {
	probe.calls = append(probe.calls, "names")
	require.Nil(probe.t, repository)
	require.Equal(probe.t, []string{"Patient", "Observation"}, names)
	return probe.summary, probe.cause
}
func (probe *providerDiagnosticForwardProbe) SyncAllByPatientEverythingBundle(repository sourceModels.DatabaseRepository, bundle interface{}) (sourceModels.UpsertSummary, error) {
	probe.calls = append(probe.calls, "everything")
	require.Nil(probe.t, repository)
	require.Same(probe.t, probe.bundle, bundle)
	return probe.summary, probe.cause
}
func (probe *providerDiagnosticForwardProbe) SyncAllBundle(repository sourceModels.DatabaseRepository, file *os.File, version sourcePkg.FhirVersion) (sourceModels.UpsertSummary, error) {
	probe.calls = append(probe.calls, "manual")
	require.Nil(probe.t, repository)
	require.Same(probe.t, probe.file, file)
	require.Equal(probe.t, probe.version, version)
	return probe.summary, probe.cause
}
func (probe *providerDiagnosticForwardProbe) ExtractPatientId(file *os.File) (string, sourcePkg.FhirVersion, error) {
	probe.calls = append(probe.calls, "patient")
	require.Same(probe.t, probe.file, file)
	return "synthetic-original-patient", probe.version, probe.cause
}
func (probe *providerDiagnosticForwardProbe) RefreshAccessToken(options ...func(*sourceModels.SourceClientRefreshOptions)) error {
	probe.calls = append(probe.calls, "refresh")
	require.Len(probe.t, options, 1)
	options[0](&sourceModels.SourceClientRefreshOptions{})
	return probe.cause
}

func TestProviderDiagnosticClientForwardsEveryOperationWithoutErasingOutputs(t *testing.T) {
	for _, fail := range []bool{false, true} {
		probe := &providerDiagnosticForwardProbe{t: t, version: sourcePkg.FhirVersion("synthetic-original-version"),
			summary: sourceModels.UpsertSummary{UpdatedResources: []string{"synthetic-original-resource"}, TotalResources: 1}}
		if fail {
			probe.cause = errors.New("synthetic-private-original-operation-cause")
		}
		var out map[string]any
		probe.out = &out
		probe.bundle = &map[string]any{"exact": "1.234567890123456789"}
		file, err := os.CreateTemp(t.TempDir(), "synthetic-sdk-forward-*")
		require.NoError(t, err)
		t.Cleanup(func() { _ = file.Close() })
		probe.file = file
		client := &providerDiagnosticClient{SourceClient: probe}
		check := func(err error) {
			if fail {
				require.EqualError(t, err, "provider operation failed")
				require.ErrorIs(t, err, probe.cause)
			} else {
				require.NoError(t, err)
			}
		}
		location, err := client.GetRequest("synthetic-original-path", &out)
		check(err)
		require.Equal(t, "synthetic-original-location", location)
		bundle, err := client.GetResourceBundle("synthetic-original-path")
		check(err)
		require.Same(t, probe.bundle, bundle)
		summary, err := client.SyncAll(nil)
		check(err)
		require.Equal(t, probe.summary, summary)
		summary, err = client.SyncAllByResourceName(nil, []string{"Patient", "Observation"})
		check(err)
		require.Equal(t, probe.summary, summary)
		summary, err = client.SyncAllByPatientEverythingBundle(nil, probe.bundle)
		check(err)
		require.Equal(t, probe.summary, summary)
		summary, err = client.SyncAllBundle(nil, file, probe.version)
		check(err)
		require.Equal(t, probe.summary, summary)
		patient, version, err := client.ExtractPatientId(file)
		check(err)
		require.Equal(t, "synthetic-original-patient", patient)
		require.Equal(t, probe.version, version)
		check(client.RefreshAccessToken(func(*sourceModels.SourceClientRefreshOptions) { probe.option = true }))
		require.True(t, probe.option)
		require.Equal(t, []string{"request", "bundle", "sync", "names", "everything", "manual", "patient", "refresh"}, probe.calls)
	}
}
