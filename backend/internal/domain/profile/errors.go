package profile

import "fmt"

type ErrorCode string

const (
	ErrorCodeInvalidID           ErrorCode = "invalid_id"
	ErrorCodeInvalidClerkSubject ErrorCode = "invalid_clerk_subject"
	ErrorCodeInvalidDisplayName  ErrorCode = "invalid_display_name"
	ErrorCodeInvalidAvatarURL    ErrorCode = "invalid_avatar_url"
	ErrorCodeInvalidTimestamp    ErrorCode = "invalid_timestamp"
)

type Error struct {
	Code      ErrorCode
	Operation string
}

func (err *Error) Error() string {
	if err.Operation == "" {
		return fmt.Sprintf("profile: %s", err.Code)
	}

	return fmt.Sprintf("profile %s: %s", err.Operation, err.Code)
}

func (err *Error) Is(target error) bool {
	other, ok := target.(*Error)
	return ok && err.Code == other.Code
}

var (
	ErrInvalidID           = &Error{Code: ErrorCodeInvalidID}
	ErrInvalidClerkSubject = &Error{Code: ErrorCodeInvalidClerkSubject}
	ErrInvalidDisplayName  = &Error{Code: ErrorCodeInvalidDisplayName}
	ErrInvalidAvatarURL    = &Error{Code: ErrorCodeInvalidAvatarURL}
	ErrInvalidTimestamp    = &Error{Code: ErrorCodeInvalidTimestamp}
)

func profileError(operation string, code ErrorCode) *Error {
	return &Error{Code: code, Operation: operation}
}
