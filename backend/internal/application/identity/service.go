package identity

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/SourceSenseiTheRealOne/coach-connect/mvp-v2/backend/internal/domain/profile"
)

const (
	validationProfileID = "profile_validation"
	defaultDisplayName  = "Coach Connect member"
)

type VerifiedIdentity interface {
	ClerkSubject() string
	PublicDisplayName() string
	PublicAvatarURL() *string
}

type ProfileRepository interface {
	FindByClerkSubject(ctx context.Context, subject string) (profile.Profile, error)
	Create(ctx context.Context, author profile.Profile) (profile.Profile, error)
	UpdatePublicIdentity(ctx context.Context, author profile.Profile) (profile.Profile, error)
}

type ProfileIDGenerator interface {
	NewProfileID(ctx context.Context) (string, error)
}

type Clock interface {
	Now() time.Time
}

type Service struct {
	repository ProfileRepository
	ids        ProfileIDGenerator
	clock      Clock
}

func NewService(repository ProfileRepository, ids ProfileIDGenerator, clock Clock) (Service, error) {
	if repository == nil || ids == nil || clock == nil {
		return Service{}, operationError("create", ErrorCodeInvalidConfig)
	}

	return Service{repository: repository, ids: ids, clock: clock}, nil
}

func (service Service) Reconcile(ctx context.Context, verified VerifiedIdentity) (profile.Profile, error) {
	if verified == nil {
		return profile.Profile{}, operationError("reconcile", ErrorCodeInvalidVerifiedIdentity)
	}

	now := service.clock.Now()
	incoming, err := profile.NewProfile(
		validationProfileID,
		verified.ClerkSubject(),
		publicDisplayName(verified.PublicDisplayName()),
		verified.PublicAvatarURL(),
		now,
		now,
	)
	if err != nil {
		return profile.Profile{}, operationError("reconcile", ErrorCodeInvalidVerifiedIdentity)
	}

	existing, err := service.repository.FindByClerkSubject(ctx, incoming.ClerkSubject())
	if err == nil {
		return service.reconcileExisting(ctx, existing, incoming, now)
	}
	if !errors.Is(err, ErrProfileNotFound) {
		return profile.Profile{}, operationError("reconcile", ErrorCodeRepositoryUnavailable)
	}

	return service.createProfile(ctx, incoming, now)
}

func (service Service) reconcileExisting(ctx context.Context, existing profile.Profile, incoming profile.Profile, now time.Time) (profile.Profile, error) {
	matches, err := existing.PublicIdentityMatches(incoming.DisplayName(), avatarPointer(incoming))
	if err != nil {
		return profile.Profile{}, operationError("reconcile", ErrorCodeInvalidVerifiedIdentity)
	}
	if matches {
		return existing, nil
	}

	updated, err := existing.WithPublicIdentity(incoming.DisplayName(), avatarPointer(incoming), now)
	if err != nil {
		return profile.Profile{}, operationError("reconcile", ErrorCodeInvalidProfile)
	}

	saved, err := service.repository.UpdatePublicIdentity(ctx, updated)
	if err != nil {
		return profile.Profile{}, operationError("reconcile", ErrorCodeRepositoryUnavailable)
	}

	return saved, nil
}

func (service Service) createProfile(ctx context.Context, incoming profile.Profile, now time.Time) (profile.Profile, error) {
	id, err := service.ids.NewProfileID(ctx)
	if err != nil {
		return profile.Profile{}, operationError("reconcile", ErrorCodeRepositoryUnavailable)
	}

	author, err := profile.NewProfile(id, incoming.ClerkSubject(), incoming.DisplayName(), avatarPointer(incoming), now, now)
	if err != nil {
		return profile.Profile{}, operationError("reconcile", ErrorCodeInvalidProfile)
	}

	saved, err := service.repository.Create(ctx, author)
	if err == nil {
		return saved, nil
	}
	if errors.Is(err, ErrProfileConflict) {
		return service.reconcileAfterCreateConflict(ctx, incoming, now)
	}

	return profile.Profile{}, operationError("reconcile", ErrorCodeRepositoryUnavailable)
}

func (service Service) reconcileAfterCreateConflict(ctx context.Context, incoming profile.Profile, now time.Time) (profile.Profile, error) {
	author, err := service.repository.FindByClerkSubject(ctx, incoming.ClerkSubject())
	if err != nil {
		return profile.Profile{}, operationError("reconcile", ErrorCodeRepositoryUnavailable)
	}

	return service.reconcileExisting(ctx, author, incoming, now)
}

func publicDisplayName(value string) string {
	if strings.TrimSpace(value) == "" {
		return defaultDisplayName
	}

	return value
}

func avatarPointer(author profile.Profile) *string {
	value, ok := author.AvatarURL()
	if !ok {
		return nil
	}

	return &value
}
