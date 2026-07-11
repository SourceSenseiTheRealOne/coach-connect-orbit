package feed

import "fmt"

type ErrorCode string

const (
	ErrorCodeInvalidID        ErrorCode = "invalid_id"
	ErrorCodeInvalidBody      ErrorCode = "invalid_body"
	ErrorCodeInvalidTimestamp ErrorCode = "invalid_timestamp"
	ErrorCodeInvalidCursor    ErrorCode = "invalid_cursor"
	ErrorCodeInvalidLimit     ErrorCode = "invalid_limit"
	ErrorCodeNotFound         ErrorCode = "not_found"
	ErrorCodeUnauthorized     ErrorCode = "unauthorized"
	ErrorCodeForbidden        ErrorCode = "forbidden"
	ErrorCodeUnavailable      ErrorCode = "unavailable"
)

type Error struct {
	Code      ErrorCode
	Operation string
}

func (err *Error) Error() string {
	if err.Operation == "" {
		return fmt.Sprintf("feed: %s", err.Code)
	}

	return fmt.Sprintf("feed %s: %s", err.Operation, err.Code)
}

func (err *Error) Is(target error) bool {
	other, ok := target.(*Error)
	return ok && err.Code == other.Code
}

var (
	ErrInvalidID        = &Error{Code: ErrorCodeInvalidID}
	ErrInvalidBody      = &Error{Code: ErrorCodeInvalidBody}
	ErrInvalidTimestamp = &Error{Code: ErrorCodeInvalidTimestamp}
	ErrInvalidCursor    = &Error{Code: ErrorCodeInvalidCursor}
	ErrInvalidLimit     = &Error{Code: ErrorCodeInvalidLimit}
	ErrNotFound         = &Error{Code: ErrorCodeNotFound}
	ErrUnauthorized     = &Error{Code: ErrorCodeUnauthorized}
	ErrForbidden        = &Error{Code: ErrorCodeForbidden}
	ErrUnavailable      = &Error{Code: ErrorCodeUnavailable}
)

func feedError(operation string, code ErrorCode) *Error {
	return &Error{Code: code, Operation: operation}
}
