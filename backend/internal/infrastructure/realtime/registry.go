package realtime

import (
	"bytes"
	"context"
)

type Config struct {
	MaxConnections  int
	OutboundBuffer  int
	MaxPayloadBytes int
}

type Registry struct {
	commands        chan registryCommand
	done            chan struct{}
	parentDone      <-chan struct{}
	maxPayloadBytes int
}

type registryCommand interface {
	execute(state *registryState) bool
}

type registryState struct {
	config      Config
	connections map[string]*Connection
}

type openResult struct {
	connection *Connection
	err        error
}

type openCommand struct {
	descriptor Descriptor
	response   chan openResult
}

func (command openCommand) execute(state *registryState) bool {
	if _, exists := state.connections[command.descriptor.ID]; exists {
		command.response <- openResult{err: operationError("open", ErrorCodeDuplicateConnection)}
		return false
	}
	if len(state.connections) >= state.config.MaxConnections {
		command.response <- openResult{err: operationError("open", ErrorCodeConnectionLimit)}
		return false
	}

	connection := newConnection(command.descriptor, state.config.OutboundBuffer)
	state.connections[command.descriptor.ID] = connection
	command.response <- openResult{connection: connection}
	return false
}

type countCommand struct {
	response chan int
}

func (command countCommand) execute(state *registryState) bool {
	command.response <- len(state.connections)
	return false
}

type deliverCommand struct {
	connectionID string
	payload      []byte
	response     chan error
}

func (command deliverCommand) execute(state *registryState) bool {
	connection, exists := state.connections[command.connectionID]
	if !exists {
		command.response <- operationError("deliver", ErrorCodeConnectionNotFound)
		return false
	}

	select {
	case connection.outbound <- command.payload:
		command.response <- nil
	default:
		delete(state.connections, command.connectionID)
		connection.terminate(Closure{Code: CloseCodeSlowConsumer, Reason: "outbound queue full"})
		command.response <- operationError("deliver", ErrorCodeSlowConsumer)
	}
	return false
}

type closeCommand struct {
	connectionID string
	closure      Closure
	response     chan error
}

func (command closeCommand) execute(state *registryState) bool {
	connection, exists := state.connections[command.connectionID]
	if !exists {
		command.response <- operationError("close", ErrorCodeConnectionNotFound)
		return false
	}

	delete(state.connections, command.connectionID)
	connection.terminate(command.closure)
	command.response <- nil
	return false
}

type shutdownCommand struct {
	response chan struct{}
}

func (command shutdownCommand) execute(state *registryState) bool {
	for id, connection := range state.connections {
		connection.terminate(Closure{Code: CloseCodeServerShutdown, Reason: "server shutting down"})
		delete(state.connections, id)
	}
	command.response <- struct{}{}
	return true
}

func NewRegistry(parent context.Context, config Config) (*Registry, error) {
	if config.MaxConnections <= 0 || config.OutboundBuffer <= 0 || config.MaxPayloadBytes <= 0 {
		return nil, operationError("create", ErrorCodeInvalidConfig)
	}
	if parent == nil {
		return nil, operationError("create", ErrorCodeInvalidConfig)
	}

	registry := &Registry{
		commands:        make(chan registryCommand),
		done:            make(chan struct{}),
		parentDone:      parent.Done(),
		maxPayloadBytes: config.MaxPayloadBytes,
	}
	go registry.run(parent, config)

	return registry, nil
}

func (registry *Registry) run(parent context.Context, config Config) {
	defer close(registry.done)

	state := &registryState{
		config:      config,
		connections: make(map[string]*Connection, config.MaxConnections),
	}
	stop := func() {
		shutdownCommand{response: make(chan struct{}, 1)}.execute(state)
	}

	for {
		select {
		case <-parent.Done():
			stop()
			return
		default:
		}

		select {
		case <-parent.Done():
			stop()
			return
		case command := <-registry.commands:
			if parent.Err() != nil {
				stop()
				return
			}
			if command.execute(state) {
				return
			}
		}
	}
}

func (registry *Registry) Open(ctx context.Context, descriptor Descriptor) (*Connection, error) {
	if !descriptor.valid() {
		return nil, operationError("open", ErrorCodeInvalidConnection)
	}

	response := make(chan openResult, 1)
	if err := registry.submit(ctx, openCommand{descriptor: descriptor, response: response}); err != nil {
		return nil, err
	}

	select {
	case result := <-response:
		return result.connection, result.err
	case <-registry.done:
		select {
		case result := <-response:
			return result.connection, result.err
		default:
			return nil, operationError("open", ErrorCodeRegistryClosed)
		}
	}
}

func (registry *Registry) Count(ctx context.Context) (int, error) {
	response := make(chan int, 1)
	if err := registry.submit(ctx, countCommand{response: response}); err != nil {
		return 0, err
	}

	select {
	case count := <-response:
		return count, nil
	case <-registry.done:
		select {
		case count := <-response:
			return count, nil
		default:
			return 0, operationError("count", ErrorCodeRegistryClosed)
		}
	}
}

func (registry *Registry) Deliver(ctx context.Context, connectionID string, payload []byte) error {
	if len(payload) > registry.maxPayloadBytes {
		return operationError("deliver", ErrorCodePayloadTooLarge)
	}

	response := make(chan error, 1)
	command := deliverCommand{
		connectionID: connectionID,
		payload:      bytes.Clone(payload),
		response:     response,
	}
	if err := registry.submit(ctx, command); err != nil {
		return err
	}

	select {
	case err := <-response:
		return err
	case <-registry.done:
		select {
		case err := <-response:
			return err
		default:
			return operationError("deliver", ErrorCodeRegistryClosed)
		}
	}
}

func (registry *Registry) Close(ctx context.Context, connectionID string, closure Closure) error {
	if !closure.valid() {
		return operationError("close", ErrorCodeInvalidClosure)
	}

	response := make(chan error, 1)
	command := closeCommand{
		connectionID: connectionID,
		closure:      closure,
		response:     response,
	}
	if err := registry.submit(ctx, command); err != nil {
		return err
	}

	select {
	case err := <-response:
		return err
	case <-registry.done:
		select {
		case err := <-response:
			return err
		default:
			return operationError("close", ErrorCodeRegistryClosed)
		}
	}
}

func (registry *Registry) Shutdown(ctx context.Context) error {
	select {
	case <-registry.done:
		return nil
	default:
	}

	response := make(chan struct{}, 1)
	if err := registry.submit(ctx, shutdownCommand{response: response}); err != nil {
		if registry.isClosed() {
			return nil
		}
		return err
	}

	select {
	case <-response:
		<-registry.done
		return nil
	case <-registry.done:
		return nil
	}
}

// submit honors caller cancellation until the actor accepts the command. After
// acceptance, public operations wait for the actor's result so a successful
// state change can never be reported as canceled and left without a handle.
func (registry *Registry) submit(ctx context.Context, command registryCommand) error {
	if ctx == nil {
		return context.Canceled
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	select {
	case <-registry.parentDone:
		return operationError("command", ErrorCodeRegistryClosed)
	default:
	}

	select {
	case registry.commands <- command:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	case <-registry.parentDone:
		return operationError("command", ErrorCodeRegistryClosed)
	case <-registry.done:
		return operationError("command", ErrorCodeRegistryClosed)
	}
}

func (registry *Registry) isClosed() bool {
	select {
	case <-registry.done:
		return true
	default:
		return false
	}
}
