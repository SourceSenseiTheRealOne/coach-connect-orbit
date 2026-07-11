package realtime

type CloseCode string

const (
	CloseCodeClientDisconnected CloseCode = "client_disconnected"
	CloseCodeTransportError     CloseCode = "transport_error"
	CloseCodeServerShutdown     CloseCode = "server_shutdown"
	CloseCodeSlowConsumer       CloseCode = "slow_consumer"
)

type Descriptor struct {
	ID             string
	ConversationID string
	UserID         string
}

func (descriptor Descriptor) valid() bool {
	return descriptor.ID != "" && descriptor.ConversationID != "" && descriptor.UserID != ""
}

type Closure struct {
	Code   CloseCode
	Reason string
}

func (closure Closure) valid() bool {
	switch closure.Code {
	case CloseCodeClientDisconnected, CloseCodeTransportError, CloseCodeServerShutdown, CloseCodeSlowConsumer:
		return true
	default:
		return false
	}
}

type Connection struct {
	descriptor Descriptor
	outbound   chan []byte
	closed     chan Closure
}

func newConnection(descriptor Descriptor, outboundBuffer int) *Connection {
	return &Connection{
		descriptor: descriptor,
		outbound:   make(chan []byte, outboundBuffer),
		closed:     make(chan Closure, 1),
	}
}

func (connection *Connection) Descriptor() Descriptor {
	return connection.descriptor
}

func (connection *Connection) Outbound() <-chan []byte {
	return connection.outbound
}

func (connection *Connection) Closed() <-chan Closure {
	return connection.closed
}

func (connection *Connection) terminate(closure Closure) {
	connection.closed <- closure
	close(connection.closed)
	close(connection.outbound)
}
