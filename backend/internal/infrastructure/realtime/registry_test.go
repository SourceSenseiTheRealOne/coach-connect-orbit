package realtime

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestNewRegistryRejectsInvalidConfiguration(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		config Config
	}{
		{name: "zero connection limit", config: Config{MaxConnections: 0, OutboundBuffer: 1, MaxPayloadBytes: 64}},
		{name: "negative connection limit", config: Config{MaxConnections: -1, OutboundBuffer: 1, MaxPayloadBytes: 64}},
		{name: "zero outbound buffer", config: Config{MaxConnections: 1, OutboundBuffer: 0, MaxPayloadBytes: 64}},
		{name: "negative outbound buffer", config: Config{MaxConnections: 1, OutboundBuffer: -1, MaxPayloadBytes: 64}},
		{name: "zero payload limit", config: Config{MaxConnections: 1, OutboundBuffer: 1, MaxPayloadBytes: 0}},
		{name: "negative payload limit", config: Config{MaxConnections: 1, OutboundBuffer: 1, MaxPayloadBytes: -1}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			registry, err := NewRegistry(context.Background(), test.config)
			if registry != nil {
				t.Fatal("expected no registry for invalid configuration")
			}
			assertErrorCode(t, err, ErrorCodeInvalidConfig)
		})
	}
}

func TestRegistryOpenValidatesDescriptor(t *testing.T) {
	t.Parallel()

	registry := newTestRegistry(t, Config{MaxConnections: 2, OutboundBuffer: 1, MaxPayloadBytes: 64})

	tests := []struct {
		name       string
		descriptor Descriptor
	}{
		{name: "missing connection id", descriptor: Descriptor{ConversationID: "conversation-1", UserID: "user-1"}},
		{name: "missing conversation id", descriptor: Descriptor{ID: "connection-1", UserID: "user-1"}},
		{name: "missing user id", descriptor: Descriptor{ID: "connection-1", ConversationID: "conversation-1"}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			connection, err := registry.Open(context.Background(), test.descriptor)
			if connection != nil {
				t.Fatal("expected no connection for invalid descriptor")
			}
			assertErrorCode(t, err, ErrorCodeInvalidConnection)
		})
	}
}

func TestRegistryOpenTracksConnectionAndRejectsDuplicateID(t *testing.T) {
	t.Parallel()

	registry := newTestRegistry(t, Config{MaxConnections: 2, OutboundBuffer: 1, MaxPayloadBytes: 64})
	descriptor := Descriptor{ID: "connection-1", ConversationID: "conversation-1", UserID: "user-1"}

	connection, err := registry.Open(context.Background(), descriptor)
	if err != nil {
		t.Fatalf("open connection: %v", err)
	}
	if connection.Descriptor() != descriptor {
		t.Fatalf("expected descriptor %#v, got %#v", descriptor, connection.Descriptor())
	}
	if cap(connection.Outbound()) != 1 {
		t.Fatalf("expected outbound capacity 1, got %d", cap(connection.Outbound()))
	}

	count, err := registry.Count(context.Background())
	if err != nil {
		t.Fatalf("count connections: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected one connection, got %d", count)
	}

	duplicate, err := registry.Open(context.Background(), descriptor)
	if duplicate != nil {
		t.Fatal("expected duplicate open to return no connection")
	}
	assertErrorCode(t, err, ErrorCodeDuplicateConnection)
}

func TestRegistryOpenEnforcesConnectionLimit(t *testing.T) {
	t.Parallel()

	registry := newTestRegistry(t, Config{MaxConnections: 1, OutboundBuffer: 1, MaxPayloadBytes: 64})
	first := Descriptor{ID: "connection-1", ConversationID: "conversation-1", UserID: "user-1"}
	second := Descriptor{ID: "connection-2", ConversationID: "conversation-1", UserID: "user-2"}

	if _, err := registry.Open(context.Background(), first); err != nil {
		t.Fatalf("open first connection: %v", err)
	}

	connection, err := registry.Open(context.Background(), second)
	if connection != nil {
		t.Fatal("expected connection limit to reject second connection")
	}
	assertErrorCode(t, err, ErrorCodeConnectionLimit)
}

func TestRegistryDeliverTargetsOneConnectionAndCopiesPayload(t *testing.T) {
	t.Parallel()

	registry := newTestRegistry(t, Config{MaxConnections: 2, OutboundBuffer: 1, MaxPayloadBytes: 64})
	first := openTestConnection(t, registry, Descriptor{ID: "connection-1", ConversationID: "conversation-1", UserID: "user-1"})
	second := openTestConnection(t, registry, Descriptor{ID: "connection-2", ConversationID: "conversation-1", UserID: "user-2"})
	payload := []byte("persisted-message")

	if err := registry.Deliver(context.Background(), first.Descriptor().ID, payload); err != nil {
		t.Fatalf("deliver payload: %v", err)
	}
	payload[0] = 'X'

	select {
	case delivered := <-first.Outbound():
		if string(delivered) != "persisted-message" {
			t.Fatalf("expected copied payload, got %q", delivered)
		}
	default:
		t.Fatal("expected first connection to receive payload")
	}

	select {
	case delivered := <-second.Outbound():
		t.Fatalf("expected second connection to receive nothing, got %q", delivered)
	default:
	}
}

func TestRegistryDeliverRejectsUnknownConnection(t *testing.T) {
	t.Parallel()

	registry := newTestRegistry(t, Config{MaxConnections: 1, OutboundBuffer: 1, MaxPayloadBytes: 64})

	err := registry.Deliver(context.Background(), "missing-connection", []byte("message"))
	assertErrorCode(t, err, ErrorCodeConnectionNotFound)
}

func TestRegistryDeliverRejectsOversizedPayloadWithoutClosingConnection(t *testing.T) {
	t.Parallel()

	registry := newTestRegistry(t, Config{MaxConnections: 1, OutboundBuffer: 1, MaxPayloadBytes: 4})
	connection := openTestConnection(t, registry, Descriptor{ID: "connection-1", ConversationID: "conversation-1", UserID: "user-1"})

	err := registry.Deliver(context.Background(), connection.Descriptor().ID, []byte("12345"))
	assertErrorCode(t, err, ErrorCodePayloadTooLarge)

	count, countErr := registry.Count(context.Background())
	if countErr != nil {
		t.Fatalf("count connections: %v", countErr)
	}
	if count != 1 {
		t.Fatalf("expected oversized payload rejection to preserve connection, got %d connections", count)
	}
}

func TestRegistryDeliverDisconnectsSlowConsumerWithoutBlocking(t *testing.T) {
	t.Parallel()

	registry := newTestRegistry(t, Config{MaxConnections: 1, OutboundBuffer: 1, MaxPayloadBytes: 64})
	connection := openTestConnection(t, registry, Descriptor{ID: "connection-1", ConversationID: "conversation-1", UserID: "user-1"})

	if err := registry.Deliver(context.Background(), connection.Descriptor().ID, []byte("first")); err != nil {
		t.Fatalf("deliver first payload: %v", err)
	}

	err := registry.Deliver(context.Background(), connection.Descriptor().ID, []byte("second"))
	assertErrorCode(t, err, ErrorCodeSlowConsumer)

	closure, ok := <-connection.Closed()
	if !ok {
		t.Fatal("expected a terminal closure notification")
	}
	if closure.Code != CloseCodeSlowConsumer {
		t.Fatalf("expected slow consumer close code, got %q", closure.Code)
	}
	if payload, ok := <-connection.Outbound(); ok {
		t.Fatalf("expected termination to discard queued payloads, got %q", payload)
	}

	count, countErr := registry.Count(context.Background())
	if countErr != nil {
		t.Fatalf("count connections: %v", countErr)
	}
	if count != 0 {
		t.Fatalf("expected slow consumer removal, got %d active connections", count)
	}
}

func TestRegistryCloseRemovesConnectionAndPublishesReason(t *testing.T) {
	t.Parallel()

	registry := newTestRegistry(t, Config{MaxConnections: 1, OutboundBuffer: 1, MaxPayloadBytes: 64})
	connection := openTestConnection(t, registry, Descriptor{ID: "connection-1", ConversationID: "conversation-1", UserID: "user-1"})
	expected := Closure{Code: CloseCodeClientDisconnected, Reason: "peer closed"}

	if err := registry.Close(context.Background(), connection.Descriptor().ID, expected); err != nil {
		t.Fatalf("close connection: %v", err)
	}

	closure, ok := <-connection.Closed()
	if !ok || closure != expected {
		t.Fatalf("expected closure %#v, got %#v (open=%v)", expected, closure, ok)
	}
	if _, ok := <-connection.Closed(); ok {
		t.Fatal("expected closure channel to close after terminal notification")
	}
	if payload, ok := <-connection.Outbound(); ok {
		t.Fatalf("expected outbound channel to close, got %q", payload)
	}

	count, err := registry.Count(context.Background())
	if err != nil {
		t.Fatalf("count connections: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected no active connections, got %d", count)
	}
}

func TestRegistryCloseRejectsInvalidOrUnknownConnection(t *testing.T) {
	t.Parallel()

	registry := newTestRegistry(t, Config{MaxConnections: 1, OutboundBuffer: 1, MaxPayloadBytes: 64})

	err := registry.Close(context.Background(), "connection-1", Closure{})
	assertErrorCode(t, err, ErrorCodeInvalidClosure)

	err = registry.Close(context.Background(), "missing", Closure{Code: CloseCodeClientDisconnected})
	assertErrorCode(t, err, ErrorCodeConnectionNotFound)
}

func TestRegistryShutdownClosesConnectionsAndRejectsLaterOperations(t *testing.T) {
	t.Parallel()

	registry := newTestRegistry(t, Config{MaxConnections: 2, OutboundBuffer: 1, MaxPayloadBytes: 64})
	first := openTestConnection(t, registry, Descriptor{ID: "connection-1", ConversationID: "conversation-1", UserID: "user-1"})
	second := openTestConnection(t, registry, Descriptor{ID: "connection-2", ConversationID: "conversation-1", UserID: "user-2"})

	if err := registry.Shutdown(context.Background()); err != nil {
		t.Fatalf("shutdown registry: %v", err)
	}
	if err := registry.Shutdown(context.Background()); err != nil {
		t.Fatalf("repeat shutdown: %v", err)
	}

	for _, connection := range []*Connection{first, second} {
		closure, ok := <-connection.Closed()
		if !ok || closure.Code != CloseCodeServerShutdown {
			t.Fatalf("expected server shutdown closure, got %#v (open=%v)", closure, ok)
		}
	}

	_, err := registry.Open(context.Background(), Descriptor{ID: "connection-3", ConversationID: "conversation-1", UserID: "user-3"})
	assertErrorCode(t, err, ErrorCodeRegistryClosed)

	err = registry.Deliver(context.Background(), first.Descriptor().ID, []byte("message"))
	assertErrorCode(t, err, ErrorCodeRegistryClosed)

	_, err = registry.Count(context.Background())
	assertErrorCode(t, err, ErrorCodeRegistryClosed)
}

func TestRegistryParentCancellationClosesConnections(t *testing.T) {
	t.Parallel()

	parent, cancel := context.WithCancel(context.Background())
	defer cancel()
	registry, err := NewRegistry(parent, Config{MaxConnections: 1, OutboundBuffer: 1, MaxPayloadBytes: 64})
	if err != nil {
		t.Fatalf("create registry: %v", err)
	}
	connection := openTestConnection(t, registry, Descriptor{ID: "connection-1", ConversationID: "conversation-1", UserID: "user-1"})

	cancel()

	select {
	case closure := <-connection.Closed():
		if closure.Code != CloseCodeServerShutdown {
			t.Fatalf("expected server shutdown closure, got %q", closure.Code)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for parent cancellation cleanup")
	}
}

type blockingCommand struct {
	started chan struct{}
	release chan struct{}
}

func (command blockingCommand) execute(_ *registryState) bool {
	close(command.started)
	<-command.release
	return false
}

func TestRegistryParentCancellationReleasesBlockedSubmitters(t *testing.T) {
	t.Parallel()

	const submitterCount = 32
	parent, cancel := context.WithCancel(context.Background())
	defer cancel()
	registry, err := NewRegistry(parent, Config{MaxConnections: submitterCount, OutboundBuffer: 1, MaxPayloadBytes: 64})
	if err != nil {
		t.Fatalf("create registry: %v", err)
	}

	started := make(chan struct{})
	release := make(chan struct{})
	defer close(release)
	if err := registry.submit(context.Background(), blockingCommand{started: started, release: release}); err != nil {
		t.Fatalf("submit blocking command: %v", err)
	}
	<-started

	results := make(chan error, submitterCount)
	for index := range submitterCount {
		go func() {
			_, openErr := registry.Open(context.Background(), Descriptor{
				ID:             fmt.Sprintf("connection-%d", index),
				ConversationID: "conversation-1",
				UserID:         fmt.Sprintf("user-%d", index),
			})
			results <- openErr
		}()
	}

	cancel()

	for range submitterCount {
		select {
		case openErr := <-results:
			assertErrorCode(t, openErr, ErrorCodeRegistryClosed)
		case <-time.After(250 * time.Millisecond):
			t.Fatal("parent cancellation did not release blocked submitter")
		}
	}
}

func TestRegistryDoesNotApplyCommandFromAlreadyCanceledContext(t *testing.T) {
	t.Parallel()

	registry := newTestRegistry(t, Config{MaxConnections: 1, OutboundBuffer: 1, MaxPayloadBytes: 64})
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	connection, err := registry.Open(ctx, Descriptor{ID: "connection-1", ConversationID: "conversation-1", UserID: "user-1"})
	if connection != nil {
		t.Fatal("expected canceled open to return no connection")
	}
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context canceled, got %v", err)
	}

	count, err := registry.Count(context.Background())
	if err != nil {
		t.Fatalf("count connections: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected canceled command to have no side effect, got %d connections", count)
	}
}

func TestRegistrySerializesConcurrentConnectionLifecycle(t *testing.T) {
	t.Parallel()

	const connectionCount = 32
	registry := newTestRegistry(t, Config{MaxConnections: connectionCount, OutboundBuffer: 1, MaxPayloadBytes: 64})
	connections := make([]*Connection, connectionCount)
	errorsFound := make(chan error, connectionCount)
	var waitGroup sync.WaitGroup

	for index := range connectionCount {
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			descriptor := Descriptor{
				ID:             fmt.Sprintf("connection-%d", index),
				ConversationID: fmt.Sprintf("conversation-%d", index%4),
				UserID:         fmt.Sprintf("user-%d", index),
			}
			connection, err := registry.Open(context.Background(), descriptor)
			if err != nil {
				errorsFound <- fmt.Errorf("open %s: %w", descriptor.ID, err)
				return
			}
			connections[index] = connection
		}()
	}
	waitGroup.Wait()
	assertNoConcurrentErrors(t, errorsFound)

	count, err := registry.Count(context.Background())
	if err != nil || count != connectionCount {
		t.Fatalf("expected %d connections, got %d: %v", connectionCount, count, err)
	}

	for _, connection := range connections {
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			payload := []byte(connection.Descriptor().ID)
			if err := registry.Deliver(context.Background(), connection.Descriptor().ID, payload); err != nil {
				errorsFound <- fmt.Errorf("deliver %s: %w", connection.Descriptor().ID, err)
				return
			}
			if delivered := <-connection.Outbound(); string(delivered) != string(payload) {
				errorsFound <- fmt.Errorf("deliver %s: got %q", connection.Descriptor().ID, delivered)
			}
		}()
	}
	waitGroup.Wait()
	assertNoConcurrentErrors(t, errorsFound)

	for _, connection := range connections {
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			if err := registry.Close(context.Background(), connection.Descriptor().ID, Closure{Code: CloseCodeClientDisconnected}); err != nil {
				errorsFound <- fmt.Errorf("close %s: %w", connection.Descriptor().ID, err)
			}
		}()
	}
	waitGroup.Wait()
	assertNoConcurrentErrors(t, errorsFound)

	count, err = registry.Count(context.Background())
	if err != nil || count != 0 {
		t.Fatalf("expected no connections, got %d: %v", count, err)
	}
}

func assertNoConcurrentErrors(t *testing.T, errorsFound <-chan error) {
	t.Helper()

	select {
	case err := <-errorsFound:
		t.Fatal(err)
	default:
	}
}

func openTestConnection(t *testing.T, registry *Registry, descriptor Descriptor) *Connection {
	t.Helper()

	connection, err := registry.Open(context.Background(), descriptor)
	if err != nil {
		t.Fatalf("open connection: %v", err)
	}
	return connection
}

func newTestRegistry(t *testing.T, config Config) *Registry {
	t.Helper()

	registry, err := NewRegistry(context.Background(), config)
	if err != nil {
		t.Fatalf("create registry: %v", err)
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		if err := registry.Shutdown(ctx); err != nil && !errors.Is(err, ErrRegistryClosed) {
			t.Errorf("shutdown registry: %v", err)
		}
	})

	return registry
}

func assertErrorCode(t *testing.T, err error, expected ErrorCode) {
	t.Helper()

	var registryError *Error
	if !errors.As(err, &registryError) {
		t.Fatalf("expected realtime Error, got %T: %v", err, err)
	}
	if registryError.Code != expected {
		t.Fatalf("expected error code %q, got %q", expected, registryError.Code)
	}
}
