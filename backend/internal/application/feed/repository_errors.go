package feed

import domain "github.com/SourceSenseiTheRealOne/coach-connect/mvp-v2/backend/internal/domain/feed"

var (
	ErrUnauthenticated = domain.ErrUnauthorized
	ErrForbidden       = domain.ErrForbidden
	ErrNotFound        = domain.ErrNotFound
	ErrUnavailable     = domain.ErrUnavailable
)
