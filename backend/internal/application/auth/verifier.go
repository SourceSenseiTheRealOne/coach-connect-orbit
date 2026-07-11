package auth

import (
	"context"
	"fmt"
	"strings"
)

type VerifiedIdentity struct {
	Subject     string
	DisplayName string
	AvatarURL   *string
}

type Verifier interface {
	VerifyBearer(ctx context.Context, bearerToken string) (VerifiedIdentity, error)
}

type FailClosedVerifier struct{}

func (FailClosedVerifier) VerifyBearer(context.Context, string) (VerifiedIdentity, error) {
	return VerifiedIdentity{}, ErrInvalidBearerToken
}

type ErrorCode string

const (
	ErrorCodeMissingBearerToken ErrorCode = "missing_bearer_token"
	ErrorCodeInvalidBearerToken ErrorCode = "invalid_bearer_token"
	ErrorCodeInvalidConfig      ErrorCode = "invalid_config"
)

type Error struct {
	Code ErrorCode
}

func (err *Error) Error() string {
	return fmt.Sprintf("auth: %s", err.Code)
}

func (err *Error) Is(target error) bool {
	other, ok := target.(*Error)
	return ok && err.Code == other.Code
}

var (
	ErrMissingBearerToken = &Error{Code: ErrorCodeMissingBearerToken}
	ErrInvalidBearerToken = &Error{Code: ErrorCodeInvalidBearerToken}
	ErrInvalidConfig      = &Error{Code: ErrorCodeInvalidConfig}
)

func ExtractBearerToken(header string) (string, error) {
	const prefix = "Bearer "
	if !strings.HasPrefix(header, prefix) {
		return "", ErrMissingBearerToken
	}

	token := strings.TrimSpace(strings.TrimPrefix(header, prefix))
	if token == "" || strings.ContainsAny(token, "\r\n\t ") {
		return "", ErrInvalidBearerToken
	}

	return token, nil
}
