package identity

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/SourceSenseiTheRealOne/coach-connect/mvp-v2/backend/internal/domain/profile"
)

func TestServiceReconcileCreatesProfileFromVerifiedIdentity(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 7, 11, 10, 0, 0, 0, time.UTC)
	avatar := "https://cdn.example.com/avatar.png"
	repository := newMemoryProfileRepository()
	service := newTestService(t, repository, []string{"profile_123"}, now)

	author, err := service.Reconcile(context.Background(), verifiedIdentity{
		subject:     "user_123",
		displayName: "Ada Coach",
		avatarURL:   &avatar,
	})
	if err != nil {
		t.Fatalf("reconcile identity: %v", err)
	}

	if author.ID() != "profile_123" {
		t.Fatalf("expected generated profile id, got %q", author.ID())
	}
	if author.ClerkSubject() != "user_123" {
		t.Fatalf("expected verified subject, got %q", author.ClerkSubject())
	}
	if author.DisplayName() != "Ada Coach" {
		t.Fatalf("expected server-sourced display name, got %q", author.DisplayName())
	}
	if got, ok := author.AvatarURL(); !ok || got != avatar {
		t.Fatalf("expected server-sourced avatar %q, got %q (ok=%v)", avatar, got, ok)
	}
	if !author.CreatedAt().Equal(now) || !author.UpdatedAt().Equal(now) {
		t.Fatalf("expected timestamps %s, got created=%s updated=%s", now, author.CreatedAt(), author.UpdatedAt())
	}
	if repository.createCalls != 1 {
		t.Fatalf("expected one create call, got %d", repository.createCalls)
	}
}

func TestServiceReconcileIsIdempotentByClerkSubject(t *testing.T) {
	t.Parallel()

	createdAt := time.Date(2026, 7, 11, 10, 0, 0, 0, time.UTC)
	now := createdAt.Add(time.Hour)
	avatar := "https://cdn.example.com/avatar.png"
	existing := mustProfile(t, "profile_existing", "user_123", "Ada Coach", &avatar, createdAt, createdAt)
	repository := newMemoryProfileRepository(existing)
	service := newTestService(t, repository, []string{"profile_new"}, now)

	author, err := service.Reconcile(context.Background(), verifiedIdentity{
		subject:     "user_123",
		displayName: "Ada Coach",
		avatarURL:   &avatar,
	})
	if err != nil {
		t.Fatalf("reconcile existing identity: %v", err)
	}

	if author.ID() != existing.ID() {
		t.Fatalf("expected stable profile id %q, got %q", existing.ID(), author.ID())
	}
	if !author.CreatedAt().Equal(createdAt) {
		t.Fatalf("expected created timestamp preserved, got %s", author.CreatedAt())
	}
	if repository.createCalls != 0 {
		t.Fatalf("expected no create call for existing subject, got %d", repository.createCalls)
	}
	if repository.updateCalls != 0 {
		t.Fatalf("expected no update call when public identity is unchanged, got %d", repository.updateCalls)
	}
}

func TestServiceReconcileUpdatesServerSourcedPublicIdentity(t *testing.T) {
	t.Parallel()

	createdAt := time.Date(2026, 7, 11, 10, 0, 0, 0, time.UTC)
	now := createdAt.Add(time.Hour)
	oldAvatar := "https://cdn.example.com/old.png"
	newAvatar := "https://cdn.example.com/new.png"
	existing := mustProfile(t, "profile_existing", "user_123", "Ada Coach", &oldAvatar, createdAt, createdAt)
	repository := newMemoryProfileRepository(existing)
	service := newTestService(t, repository, nil, now)

	author, err := service.Reconcile(context.Background(), verifiedIdentity{
		subject:     "user_123",
		displayName: "Ada Tactician",
		avatarURL:   &newAvatar,
	})
	if err != nil {
		t.Fatalf("reconcile changed public identity: %v", err)
	}

	if author.ID() != existing.ID() {
		t.Fatalf("expected stable profile id %q, got %q", existing.ID(), author.ID())
	}
	if author.ClerkSubject() != existing.ClerkSubject() {
		t.Fatalf("expected verified subject preserved, got %q", author.ClerkSubject())
	}
	if author.DisplayName() != "Ada Tactician" {
		t.Fatalf("expected updated display name, got %q", author.DisplayName())
	}
	if got, ok := author.AvatarURL(); !ok || got != newAvatar {
		t.Fatalf("expected updated avatar %q, got %q (ok=%v)", newAvatar, got, ok)
	}
	if !author.CreatedAt().Equal(createdAt) {
		t.Fatalf("expected created timestamp preserved, got %s", author.CreatedAt())
	}
	if !author.UpdatedAt().Equal(now) {
		t.Fatalf("expected update timestamp %s, got %s", now, author.UpdatedAt())
	}
	if repository.updateCalls != 1 {
		t.Fatalf("expected one update call, got %d", repository.updateCalls)
	}
}

func TestServiceReconcileUsesSafePreOnboardingNameWhenClerkNameIsBlank(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 7, 11, 10, 0, 0, 0, time.UTC)
	repository := newMemoryProfileRepository()
	service := newTestService(t, repository, []string{"profile_123"}, now)

	author, err := service.Reconcile(context.Background(), verifiedIdentity{
		subject:     "user_123",
		displayName: " ",
	})
	if err != nil {
		t.Fatalf("reconcile identity without Clerk public name: %v", err)
	}
	if author.DisplayName() != "Coach Connect member" {
		t.Fatalf("expected privacy-safe pre-onboarding name, got %q", author.DisplayName())
	}
}

func TestServiceReconcileRejectsUnsafeVerifiedIdentityBeforeRepositoryUse(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 7, 11, 10, 0, 0, 0, time.UTC)
	unsafeAvatar := "javascript:alert(1)"

	tests := []struct {
		name     string
		identity verifiedIdentity
	}{
		{name: "blank subject", identity: verifiedIdentity{subject: " ", displayName: "Ada Coach"}},
		{name: "unsafe avatar", identity: verifiedIdentity{subject: "user_123", displayName: "Ada Coach", avatarURL: &unsafeAvatar}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			repository := newMemoryProfileRepository()
			service := newTestService(t, repository, []string{"profile_123"}, now)

			author, err := service.Reconcile(context.Background(), test.identity)
			if err == nil {
				t.Fatalf("expected invalid verified identity error, got profile %#v", author)
			}
			assertIdentityErrorCode(t, err, ErrorCodeInvalidVerifiedIdentity)
			if repository.findCalls != 0 || repository.createCalls != 0 || repository.updateCalls != 0 {
				t.Fatalf("expected repository unused, got find=%d create=%d update=%d", repository.findCalls, repository.createCalls, repository.updateCalls)
			}
		})
	}
}

func TestServiceReconcileMapsRepositoryFailuresToSafeErrors(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 7, 11, 10, 0, 0, 0, time.UTC)
	repository := newMemoryProfileRepository()
	repository.findErr = errors.New("sql: password=secret host=db.internal")
	service := newTestService(t, repository, []string{"profile_123"}, now)

	author, err := service.Reconcile(context.Background(), verifiedIdentity{subject: "user_123", displayName: "Ada Coach"})
	if err == nil {
		t.Fatalf("expected repository error, got profile %#v", author)
	}
	assertIdentityErrorCode(t, err, ErrorCodeRepositoryUnavailable)
	if strings.Contains(err.Error(), "password") || strings.Contains(err.Error(), "db.internal") {
		t.Fatalf("error leaked infrastructure detail: %q", err.Error())
	}
}

func TestServiceReconcileHandlesCreateConflictByReconcilingStableProfile(t *testing.T) {
	t.Parallel()

	createdAt := time.Date(2026, 7, 11, 10, 0, 0, 0, time.UTC)
	now := createdAt.Add(time.Hour)
	existing := mustProfile(t, "profile_existing", "user_123", "Stale Name", nil, createdAt, createdAt)
	repository := newMemoryProfileRepository()
	repository.createErr = ErrProfileConflict
	repository.afterCreateConflict = &existing
	service := newTestService(t, repository, []string{"profile_new"}, now)

	author, err := service.Reconcile(context.Background(), verifiedIdentity{subject: "user_123", displayName: "Ada Coach"})
	if err != nil {
		t.Fatalf("reconcile after create conflict: %v", err)
	}

	if author.ID() != existing.ID() {
		t.Fatalf("expected conflicted create to preserve stable profile id %q, got %q", existing.ID(), author.ID())
	}
	if author.DisplayName() != "Ada Coach" {
		t.Fatalf("expected latest verified display name after conflict, got %q", author.DisplayName())
	}
	if repository.createCalls != 1 {
		t.Fatalf("expected one create attempt, got %d", repository.createCalls)
	}
	if repository.findCalls != 2 {
		t.Fatalf("expected initial find plus conflict reload, got %d", repository.findCalls)
	}
	if repository.updateCalls != 1 {
		t.Fatalf("expected conflict reload to reconcile public identity once, got %d", repository.updateCalls)
	}
}

func TestIdentityErrorUsesStableSafeCode(t *testing.T) {
	t.Parallel()

	err := operationError("reconcile", ErrorCodeRepositoryUnavailable)
	if err.Error() != "identity reconcile: repository_unavailable" {
		t.Fatalf("unexpected safe error message %q", err.Error())
	}
	if !errors.Is(err, ErrRepositoryUnavailable) {
		t.Fatal("expected errors.Is to match identity error code")
	}
	if errors.Is(err, ErrInvalidVerifiedIdentity) {
		t.Fatal("different identity error codes must not match")
	}
}

type verifiedIdentity struct {
	subject     string
	displayName string
	avatarURL   *string
}

func (identity verifiedIdentity) ClerkSubject() string {
	return identity.subject
}

func (identity verifiedIdentity) PublicDisplayName() string {
	return identity.displayName
}

func (identity verifiedIdentity) PublicAvatarURL() *string {
	return identity.avatarURL
}

type staticClock struct {
	now time.Time
}

func (clock staticClock) Now() time.Time {
	return clock.now
}

type queuedProfileIDs struct {
	values []string
}

func (generator *queuedProfileIDs) NewProfileID(_ context.Context) (string, error) {
	if len(generator.values) == 0 {
		return "", errors.New("no profile id queued")
	}

	value := generator.values[0]
	generator.values = generator.values[1:]
	return value, nil
}

type memoryProfileRepository struct {
	bySubject           map[string]profile.Profile
	findErr             error
	createErr           error
	updateErr           error
	afterCreateConflict *profile.Profile
	findCalls           int
	createCalls         int
	updateCalls         int
}

func newMemoryProfileRepository(authors ...profile.Profile) *memoryProfileRepository {
	repository := &memoryProfileRepository{bySubject: make(map[string]profile.Profile)}
	for _, author := range authors {
		repository.bySubject[author.ClerkSubject()] = author
	}
	return repository
}

func (repository *memoryProfileRepository) FindByClerkSubject(_ context.Context, subject string) (profile.Profile, error) {
	repository.findCalls++
	if repository.findErr != nil {
		return profile.Profile{}, repository.findErr
	}

	author, ok := repository.bySubject[subject]
	if !ok {
		return profile.Profile{}, ErrProfileNotFound
	}
	return author, nil
}

func (repository *memoryProfileRepository) Create(_ context.Context, author profile.Profile) (profile.Profile, error) {
	repository.createCalls++
	if repository.createErr != nil {
		if errors.Is(repository.createErr, ErrProfileConflict) && repository.afterCreateConflict != nil {
			repository.bySubject[repository.afterCreateConflict.ClerkSubject()] = *repository.afterCreateConflict
		}
		return profile.Profile{}, repository.createErr
	}

	repository.bySubject[author.ClerkSubject()] = author
	return author, nil
}

func (repository *memoryProfileRepository) UpdatePublicIdentity(_ context.Context, author profile.Profile) (profile.Profile, error) {
	repository.updateCalls++
	if repository.updateErr != nil {
		return profile.Profile{}, repository.updateErr
	}

	repository.bySubject[author.ClerkSubject()] = author
	return author, nil
}

func newTestService(t *testing.T, repository *memoryProfileRepository, ids []string, now time.Time) Service {
	t.Helper()

	service, err := NewService(repository, &queuedProfileIDs{values: ids}, staticClock{now: now})
	if err != nil {
		t.Fatalf("create identity service: %v", err)
	}
	return service
}

func mustProfile(t *testing.T, id string, subject string, name string, avatar *string, createdAt time.Time, updatedAt time.Time) profile.Profile {
	t.Helper()

	author, err := profile.NewProfile(id, subject, name, avatar, createdAt, updatedAt)
	if err != nil {
		t.Fatalf("create profile fixture: %v", err)
	}
	return author
}

func assertIdentityErrorCode(t *testing.T, err error, expected ErrorCode) {
	t.Helper()

	var identityErr *Error
	if !errors.As(err, &identityErr) {
		t.Fatalf("expected identity Error, got %T: %v", err, err)
	}
	if identityErr.Code != expected {
		t.Fatalf("expected identity error code %q, got %q", expected, identityErr.Code)
	}
}
