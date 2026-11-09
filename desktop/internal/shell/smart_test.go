package shell

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/karamhouse/fasten-desktop/internal/host"
	"github.com/karamhouse/fasten-desktop/internal/smart"
	"github.com/stretchr/testify/require"
	"github.com/wailsapp/wails/v3/pkg/application"
)

type smartTestWindow struct {
	application.Window
	name string
}

func (w smartTestWindow) Name() string { return w.name }
func smartTestContext(name string) context.Context {
	return context.WithValue(context.Background(), application.WindowKey, smartTestWindow{name: name})
}

type smartTestManager struct {
	connects, patients, clinicals, cancels, disconnects int
	onConnect, onRead                                   func()
}

func (m *smartTestManager) Providers() []smart.ProviderInfo { return nil }
func (m *smartTestManager) Connect(context.Context, string) (smart.Connection, error) {
	m.connects++
	if m.onConnect != nil {
		m.onConnect()
	}
	return smart.Connection{ID: "synthetic-connection", ProviderID: "synthetic-provider", PatientID: "synthetic-patient"}, nil
}
func (m *smartTestManager) ReadPatient(context.Context, string) (string, error) {
	m.patients++
	if m.onRead != nil {
		m.onRead()
	}
	return `{"resourceType":"Patient","id":"synthetic-patient","name":[{"text":"synthetic-private-Patient"}]}`, nil
}
func (m *smartTestManager) ReadClinicalSnapshot(context.Context, string) (string, error) {
	m.clinicals++
	if m.onRead != nil {
		m.onRead()
	}
	return `{"resourceType":"Bundle","type":"collection","synthetic-private-marker":"raw-value"}`, nil
}
func (m *smartTestManager) Cancel(string)     { m.cancels++ }
func (m *smartTestManager) Disconnect(string) { m.disconnects++ }

func TestSMARTServiceRejectsCallsWithoutMainWindow(t *testing.T) {
	// Reject before reaching manager/resolver, not just after data retrieval.
	calls := 0
	service := NewSMARTService(nil, func(context.Context, string) (host.LocalIdentity, error) { calls++; return host.LocalIdentity{}, nil })
	for _, name := range []string{"", "auth-synthetic-provider", "not-main"} {
		ctx := smartTestContext(name)
		_, err := service.Providers(ctx)
		require.Error(t, err)
		_, err = service.Connect(ctx, "epic-test", "synthetic-token")
		require.Error(t, err)
		_, err = service.ReadPatient(ctx, "connection", "synthetic-token")
		require.Error(t, err)
		_, err = service.ReadClinicalSnapshot(ctx, "connection", "synthetic-token")
		require.Error(t, err)
		require.Error(t, service.Cancel(ctx, "epic-test"))
		require.Error(t, service.Disconnect(ctx, "connection"))
	}
	require.Zero(t, calls)
}

func TestSMARTServiceScopesDataToAuthenticatedLocalOwner(t *testing.T) {
	ctx := smartTestContext("main")
	owner := host.LocalIdentity{ID: uuid.New(), Username: "synthetic-owner"}
	other := host.LocalIdentity{ID: uuid.New(), Username: "synthetic-other"}
	active := true
	resolver := func(_ context.Context, token string) (host.LocalIdentity, error) {
		if token == "synthetic-owner-token" && active {
			return owner, nil
		}
		if token == "synthetic-other-token" {
			return other, nil
		}
		return host.LocalIdentity{}, errors.New("synthetic-private-auth-rejection")
	}
	manager := &smartTestManager{}
	service := NewSMARTService(manager, resolver)
	_, err := service.Connect(ctx, "synthetic-provider", "")
	require.Error(t, err)
	_, err = service.Connect(ctx, "synthetic-provider", "invalid")
	require.Error(t, err)
	require.NotContains(t, err.Error(), "synthetic-private-")
	require.Zero(t, manager.connects)
	connection, err := service.Connect(ctx, "synthetic-provider", "synthetic-owner-token")
	require.NoError(t, err)
	require.Equal(t, owner, connection.Owner)
	_, err = service.ReadPatient(ctx, connection.ID, "synthetic-other-token")
	require.Error(t, err)
	_, err = service.ReadClinicalSnapshot(ctx, connection.ID, "synthetic-other-token")
	require.Error(t, err)
	require.Zero(t, manager.patients)
	require.Zero(t, manager.clinicals)
	patient, err := service.ReadPatient(ctx, connection.ID, "synthetic-owner-token")
	require.NoError(t, err)
	require.Contains(t, patient, "synthetic-private-Patient")
	clinical, err := service.ReadClinicalSnapshot(ctx, connection.ID, "synthetic-owner-token")
	require.NoError(t, err)
	require.Contains(t, clinical, "synthetic-private-marker")
	// A valid request may lose its local session during provider I/O. Do not
	// hand any retrieved payload back to the renderer after revalidation fails.
	manager.onRead = func() { active = false }
	raw, err := service.ReadClinicalSnapshot(ctx, connection.ID, "synthetic-owner-token")
	require.Error(t, err)
	require.Empty(t, raw)
	require.NotContains(t, err.Error(), "synthetic-private-")
	// Logout/expiry must not prevent best-effort credential cleanup.
	require.NoError(t, service.Cancel(ctx, "synthetic-provider"))
	require.NoError(t, service.Disconnect(ctx, connection.ID))
	require.Equal(t, 1, manager.cancels)
	require.Equal(t, 1, manager.disconnects)
	active = true
	_, err = service.ReadPatient(ctx, connection.ID, "synthetic-owner-token")
	require.Error(t, err)
}

func TestSMARTServiceDiscardsConnectionIfIdentityChanges(t *testing.T) {
	ctx := smartTestContext("main")
	owner := host.LocalIdentity{ID: uuid.New(), Username: "synthetic-owner"}
	current := owner
	manager := &smartTestManager{onConnect: func() { current = host.LocalIdentity{ID: uuid.New(), Username: owner.Username} }}
	service := NewSMARTService(manager, func(context.Context, string) (host.LocalIdentity, error) { return current, nil })
	connection, err := service.Connect(ctx, "synthetic-provider", "synthetic-token")
	require.Error(t, err)
	require.Empty(t, connection.ID)
	require.Equal(t, 1, manager.disconnects)
	require.Empty(t, service.owners, "same reusable username is not the same immutable identity")
}
