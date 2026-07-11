package realtime

import "fmt"

type ErrorCode string

const (
	ErrorCodeInvalidConfig       ErrorCode = "invalid_config"
	ErrorCodeInvalidConnection   ErrorCode = "invalid_connection"
	ErrorCodeInvalidClosure      ErrorCode = "invalid_closure"
	ErrorCodeDuplicateConnection ErrorCode = "duplicate_connection"
	ErrorCodeConnectionLimit     ErrorCode = "connection_limit"
	ErrorCodeConnectionNotFound  ErrorCode = "connection_not_found"
	ErrorCodePayloadTooLarge     ErrorCode = "payload_too_large"
	ErrorCodeSlowConsumer        ErrorCode = "slow_consumer"
	ErrorCodeRegistryClosed      ErrorCode = "registry_closed"
)

type Error struct {
	Code      ErrorCode
	Operation string
}

func (err *Error) Error() string {
	if err.Operation == "" {
		return fmt.Sprintf("realtime: %s", err.Code)
	}

	return fmt.Sprintf("realtime %s: %s", err.Operation, err.Code)
}

func (err *Error) Is(target error) bool {
	other, ok := target.(*Error)
	return ok && err.Code == other.Code
}

var ErrRegistryClosed = &Error{Code: ErrorCodeRegistryClosed}

func operationError(operation string, code ErrorCode) *Error {
	return &Error{Code: code, Operation: operation}
}
