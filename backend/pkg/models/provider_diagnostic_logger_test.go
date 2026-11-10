package models

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
)

func TestProviderDiagnosticLoggerPreservesSeverityAndControlFlow(t *testing.T) {
	const marker = "synthetic-private-sdk-log-field"
	var output bytes.Buffer
	parent := logrus.New()
	parent.SetLevel(logrus.TraceLevel)
	formatter := &logrus.JSONFormatter{}
	parent.SetFormatter(formatter)
	parent.SetOutput(&output)
	var exits []int
	parent.ExitFunc = func(code int) { exits = append(exits, code) }
	private := providerDiagnosticLogger(parent.WithField("job_id", "synthetic-owned-job"))
	// A function-valued field also exercises logrus's internal field-error text;
	// neither WithFields nor WithError can escape via a returned *Entry.
	private.WithFields(logrus.Fields{marker: func() {}}).WithError(errors.New(marker)).Warnf("%s", marker)
	private.Trace(marker)
	private.Debug(marker)
	private.Info(marker)
	private.Error(marker)
	private.Fatal(marker) // The parent's owned test ExitFunc returns, not os.Exit.
	var recovered any
	func() {
		defer func() { recovered = recover() }()
		private.WithField(marker, func() {}).Panic(marker)
	}()
	require.Equal(t, []int{1}, exits, "Fatal delegates exactly one owned exit")
	panicEntry, ok := recovered.(*logrus.Entry)
	require.True(t, ok, "Panic still panics with a logrus entry")
	require.Equal(t, "provider SDK diagnostic", panicEntry.Message)
	require.Empty(t, panicEntry.Data)
	require.Nil(t, panicEntry.Context)
	require.NotContains(t, fmt.Sprint(panicEntry), marker, "panic formatter cannot expose internal raw-field error text")
	require.NotContains(t, output.String(), marker)
	decoder := json.NewDecoder(bytes.NewReader(output.Bytes()))
	var levels []string
	for {
		var event map[string]any
		err := decoder.Decode(&event)
		if err == io.EOF {
			break
		}
		require.NoError(t, err)
		require.Equal(t, "provider SDK diagnostic", event["msg"])
		require.Equal(t, "provider_sdk", event["component"])
		require.Equal(t, "synthetic-owned-job", event["job_id"])
		levels = append(levels, event["level"].(string))
	}
	require.Equal(t, []string{"warning", "trace", "debug", "info", "error", "fatal", "panic"}, levels)
	require.Equal(t, logrus.TraceLevel, parent.GetLevel())
	require.Same(t, formatter, parent.Formatter)
	require.Same(t, &output, parent.Out)
}

func TestProviderDiagnosticLoggerKeepsParentLevelFiltering(t *testing.T) {
	var output bytes.Buffer
	parent := logrus.New()
	parent.SetOutput(&output)
	parent.SetLevel(logrus.WarnLevel)
	private := providerDiagnosticLogger(parent)
	private.Debug("synthetic suppressed debug")
	require.Empty(t, output.String())
	private.Warn("synthetic warning payload")
	require.Contains(t, output.String(), "provider SDK diagnostic")
	require.NotContains(t, output.String(), "synthetic warning payload")
}

func TestProviderOperationErrorsAreFixedButRemainInspectable(t *testing.T) {
	const marker = "synthetic-private-provider-operation-location"
	safeOAuth := &oauth2.RetrieveError{ErrorCode: "invalid_grant", ErrorDescription: marker, Body: []byte(marker)}
	unknownOAuth := &oauth2.RetrieveError{ErrorCode: marker, ErrorDescription: marker}
	for _, item := range []struct {
		cause error
		want  string
	}{
		{fmt.Errorf("%s: %w", marker, context.Canceled), "provider operation canceled"},
		{fmt.Errorf("%s: %w", marker, context.DeadlineExceeded), "provider operation deadline exceeded"},
		{errors.New(marker), "provider operation failed"},
		{safeOAuth, "provider operation failed: invalid_grant"},
		{unknownOAuth, "provider operation failed"},
	} {
		err := privateProviderOperationError(item.cause)
		require.EqualError(t, err, item.want)
		require.NotContains(t, err.Error(), marker)
		require.ErrorIs(t, err, item.cause)
		require.Same(t, item.cause, errors.Unwrap(err))
	}
	var inspected *oauth2.RetrieveError
	require.ErrorAs(t, privateProviderOperationError(safeOAuth), &inspected)
	require.Same(t, safeOAuth, inspected)
	require.Nil(t, privateProviderOperationError(nil))
}
