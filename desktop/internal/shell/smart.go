package shell

import (
	"context"
	"errors"
	"sync"

	"github.com/google/uuid"
	"github.com/karamhouse/fasten-desktop/internal/host"
	"github.com/karamhouse/fasten-desktop/internal/smart"
)

type smartConnectionManager interface {
	Providers() []smart.ProviderInfo
	Connect(context.Context, string) (smart.Connection, error)
	ReadPatient(context.Context, string) (string, error)
	ReadClinicalSnapshot(context.Context, string) (string, error)
	Cancel(string)
	Disconnect(string)
}

// SMARTService requires both the main window and an independently authenticated
// local owner for clinical data. Provider credentials remain only in the manager.
type SMARTService struct {
	manager  smartConnectionManager
	identity host.IdentityResolver
	mu       sync.Mutex
	owners   map[string]host.LocalIdentity
}

type SMARTConnection struct {
	smart.Connection
	Owner host.LocalIdentity `json:"owner"`
}

var errSMARTIdentity = errors.New("direct provider action requires the same active local identity")
var errSMARTData = errors.New("direct provider data could not be retrieved")

func NewSMARTService(manager smartConnectionManager, identity host.IdentityResolver) *SMARTService {
	return &SMARTService{manager: manager, identity: identity, owners: map[string]host.LocalIdentity{}}
}
func authorizeSMART(ctx context.Context) error { return host.AuthorizeWindow(caller(ctx), "info", "") }

func (s *SMARTService) Providers(ctx context.Context) ([]smart.ProviderInfo, error) {
	if err := authorizeSMART(ctx); err != nil {
		return nil, err
	}
	if s.manager == nil {
		return nil, errSMARTData
	}
	return s.manager.Providers(), nil
}

func (s *SMARTService) Connect(ctx context.Context, providerID, token string) (SMARTConnection, error) {
	owner, err := s.authorizeIdentity(ctx, token)
	if err != nil {
		return SMARTConnection{}, err
	}
	connection, err := s.manager.Connect(ctx, providerID)
	if err != nil {
		return SMARTConnection{}, errSMARTData
	}
	verified, err := s.authorizeIdentity(ctx, token)
	if err != nil || owner.ID != verified.ID || ctx.Err() != nil {
		s.manager.Disconnect(connection.ID)
		return SMARTConnection{}, errSMARTIdentity
	}
	s.mu.Lock()
	s.owners[connection.ID] = verified
	s.mu.Unlock()
	return SMARTConnection{Connection: connection, Owner: verified}, nil
}

func (s *SMARTService) ReadPatient(ctx context.Context, connectionID, token string) (string, error) {
	return s.read(ctx, connectionID, token, false)
}

func (s *SMARTService) ReadClinicalSnapshot(ctx context.Context, connectionID, token string) (string, error) {
	return s.read(ctx, connectionID, token, true)
}

func (s *SMARTService) read(ctx context.Context, connectionID, token string, clinical bool) (string, error) {
	if _, err := s.authorizeConnection(ctx, connectionID, token); err != nil {
		return "", err
	}
	var raw string
	var err error
	if clinical {
		raw, err = s.manager.ReadClinicalSnapshot(ctx, connectionID)
	} else {
		raw, err = s.manager.ReadPatient(ctx, connectionID)
	}
	if err != nil {
		return "", errSMARTData
	}
	if _, err := s.authorizeConnection(ctx, connectionID, token); err != nil || ctx.Err() != nil {
		return "", errSMARTIdentity
	}
	return raw, nil
}

func (s *SMARTService) authorizeIdentity(ctx context.Context, token string) (host.LocalIdentity, error) {
	if err := authorizeSMART(ctx); err != nil {
		return host.LocalIdentity{}, err
	}
	if s.manager == nil || s.identity == nil || token == "" {
		return host.LocalIdentity{}, errSMARTIdentity
	}
	owner, err := s.identity(ctx, token)
	if err != nil || owner.ID == uuid.Nil || owner.Username == "" {
		return host.LocalIdentity{}, errSMARTIdentity
	}
	return owner, nil
}

func (s *SMARTService) authorizeConnection(ctx context.Context, id, token string) (host.LocalIdentity, error) {
	owner, err := s.authorizeIdentity(ctx, token)
	if err != nil {
		return host.LocalIdentity{}, err
	}
	s.mu.Lock()
	bound, exists := s.owners[id]
	s.mu.Unlock()
	if !exists || bound.ID != owner.ID {
		return host.LocalIdentity{}, errSMARTIdentity
	}
	return owner, nil
}

func (s *SMARTService) Cancel(ctx context.Context, providerID string) error {
	if err := authorizeSMART(ctx); err != nil {
		return err
	}
	if s.manager == nil {
		return errSMARTData
	}
	// Cleanup is deliberately available after local logout/expiry. It can
	// cancel work but cannot return data or establish a new connection.
	s.manager.Cancel(providerID)
	return nil
}
func (s *SMARTService) Disconnect(ctx context.Context, connectionID string) error {
	if err := authorizeSMART(ctx); err != nil {
		return err
	}
	if s.manager == nil {
		return errSMARTData
	}
	s.mu.Lock()
	delete(s.owners, connectionID)
	s.mu.Unlock()
	s.manager.Disconnect(connectionID)
	return nil
}
