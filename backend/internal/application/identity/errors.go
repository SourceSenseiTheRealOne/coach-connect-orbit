package identity

import "fmt"

type ErrorCode string

const (
	ErrorCodeInvalidConfig           ErrorCode = "invalid_config"
	ErrorCodeInvalidVerifiedIdentity ErrorCode = "invalid_verified_identity"
	ErrorCodeInvalidProfile          ErrorCode = "invalid_profile"
	ErrorCodeProfileNotFound         ErrorCode = "profile_not_found"
	ErrorCodeProfileConflict         ErrorCode = "profile_conflict"
	ErrorCodeRepositoryUnavailable   ErrorCode = "repository_unavailable"
)

type Error struct {
	Code      ErrorCode
	Operation string
}

func (err *Error) Error() string {
	if err.Operation == "" {
		return fmt.Sprintf("identity: %s", err.Code)
	}

	return fmt.Sprintf("identity %s: %s", err.Operation, err.Code)
}

func (err *Error) Is(target error) bool {
	other, ok := target.(*Error)
	return ok && err.Code == other.Code
}

var (
	ErrInvalidConfig           = &Error{Code: ErrorCodeInvalidConfig}
	ErrInvalidVerifiedIdentity = &Error{Code: ErrorCodeInvalidVerifiedIdentity}
	ErrInvalidProfile          = &Error{Code: ErrorCodeInvalidProfile}
	ErrProfileNotFound         = &Error{Code: ErrorCodeProfileNotFound}
	ErrProfileConflict         = &Error{Code: ErrorCodeProfileConflict}
	ErrRepositoryUnavailable   = &Error{Code: ErrorCodeRepositoryUnavailable}
)

func operationError(operation string, code ErrorCode) *Error {
	return &Error{Code: code, Operation: operation}
}
