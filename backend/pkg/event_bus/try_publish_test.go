package event_bus

import (
	"testing"

	"github.com/fastenhealth/fasten-onprem/backend/pkg/models"
	"github.com/stretchr/testify/require"
)

func TestTryPublishMessageAdmission(t *testing.T) {
	bus := &eventBus{message: make(chan EventBusMessage, 1)}
	event := models.NewEventSourceComplete("synthetic-owner", "synthetic-source")
	require.NoError(t, bus.TryPublishMessage(event))
	require.ErrorIs(t, bus.TryPublishMessage(event), ErrNotificationUnavailable, "a full bounded channel cannot create a blocked sender")
	message := <-bus.message
	require.Equal(t, "synthetic-owner", message.UserID)
	require.Contains(t, message.Message, "synthetic-source")
	require.Contains(t, message.Message, "source_complete")
	bus.message = make(chan EventBusMessage)
	require.ErrorIs(t, bus.TryPublishMessage(event), ErrNotificationUnavailable, "an absent dispatcher cannot admit a message")
	noop, ok := NewNoopEventBusServer().(BestEffortPublisher)
	require.True(t, ok)
	require.NoError(t, noop.TryPublishMessage(event))
}
