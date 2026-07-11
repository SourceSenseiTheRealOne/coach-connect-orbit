package profile

import (
	"errors"
	"strings"
	"testing"
	"time"
)

func TestNewProfileAcceptsValidAuthorIdentity(t *testing.T) {
	t.Parallel()

	avatar := "https://cdn.example.com/avatars/user_123.png"
	createdAt := time.Date(2026, 7, 11, 10, 0, 0, 0, time.UTC)
	updatedAt := createdAt.Add(time.Minute)

	author, err := NewProfile("profile_123", "user_123", "  Ada Coach  ", &avatar, createdAt, updatedAt)
	if err != nil {
		t.Fatalf("create profile: %v", err)
	}

	if author.ID() != "profile_123" {
		t.Fatalf("expected stable internal id, got %q", author.ID())
	}
	if author.ClerkSubject() != "user_123" {
		t.Fatalf("expected clerk subject, got %q", author.ClerkSubject())
	}
	if author.DisplayName() != "Ada Coach" {
		t.Fatalf("expected trimmed display name, got %q", author.DisplayName())
	}
	if got, ok := author.AvatarURL(); !ok || got != avatar {
		t.Fatalf("expected avatar %q, got %q (ok=%v)", avatar, got, ok)
	}
	if !author.CreatedAt().Equal(createdAt) {
		t.Fatalf("expected created timestamp %s, got %s", createdAt, author.CreatedAt())
	}
	if !author.UpdatedAt().Equal(updatedAt) {
		t.Fatalf("expected updated timestamp %s, got %s", updatedAt, author.UpdatedAt())
	}

	avatar = "https://cdn.example.com/avatars/changed.png"
	if got, _ := author.AvatarURL(); got == avatar {
		t.Fatal("profile must not expose caller-owned avatar pointer")
	}
}

func TestNewProfileAllowsMissingAvatar(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 7, 11, 10, 0, 0, 0, time.UTC)
	author, err := NewProfile("profile_123", "user_123", "Ada Coach", nil, now, now)
	if err != nil {
		t.Fatalf("create profile without avatar: %v", err)
	}

	if avatar, ok := author.AvatarURL(); ok || avatar != "" {
		t.Fatalf("expected no avatar, got %q (ok=%v)", avatar, ok)
	}
}

func TestNewProfileRejectsUnsafeFields(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 7, 11, 10, 0, 0, 0, time.UTC)
	validAvatar := "https://cdn.example.com/avatar.png"
	longSubject := "user_" + strings.Repeat("x", 257)
	longName := strings.Repeat("A", 81)

	tests := []struct {
		name        string
		id          string
		subject     string
		displayName string
		avatarURL   *string
		createdAt   time.Time
		updatedAt   time.Time
		code        ErrorCode
	}{
		{name: "missing id", id: "", subject: "user_123", displayName: "Ada Coach", avatarURL: &validAvatar, createdAt: now, updatedAt: now, code: ErrorCodeInvalidID},
		{name: "blank subject", id: "profile_123", subject: " ", displayName: "Ada Coach", avatarURL: &validAvatar, createdAt: now, updatedAt: now, code: ErrorCodeInvalidClerkSubject},
		{name: "subject with whitespace", id: "profile_123", subject: " user_123 ", displayName: "Ada Coach", avatarURL: &validAvatar, createdAt: now, updatedAt: now, code: ErrorCodeInvalidClerkSubject},
		{name: "subject with control", id: "profile_123", subject: "user_\n123", displayName: "Ada Coach", avatarURL: &validAvatar, createdAt: now, updatedAt: now, code: ErrorCodeInvalidClerkSubject},
		{name: "subject too long", id: "profile_123", subject: longSubject, displayName: "Ada Coach", avatarURL: &validAvatar, createdAt: now, updatedAt: now, code: ErrorCodeInvalidClerkSubject},
		{name: "blank display name", id: "profile_123", subject: "user_123", displayName: " ", avatarURL: &validAvatar, createdAt: now, updatedAt: now, code: ErrorCodeInvalidDisplayName},
		{name: "display name with control", id: "profile_123", subject: "user_123", displayName: "Ada\nCoach", avatarURL: &validAvatar, createdAt: now, updatedAt: now, code: ErrorCodeInvalidDisplayName},
		{name: "display name too long", id: "profile_123", subject: "user_123", displayName: longName, avatarURL: &validAvatar, createdAt: now, updatedAt: now, code: ErrorCodeInvalidDisplayName},
		{name: "avatar is not https", id: "profile_123", subject: "user_123", displayName: "Ada Coach", avatarURL: stringPtr("http://cdn.example.com/avatar.png"), createdAt: now, updatedAt: now, code: ErrorCodeInvalidAvatarURL},
		{name: "avatar has credentials", id: "profile_123", subject: "user_123", displayName: "Ada Coach", avatarURL: stringPtr("https://user:pass@cdn.example.com/avatar.png"), createdAt: now, updatedAt: now, code: ErrorCodeInvalidAvatarURL},
		{name: "avatar has fragment", id: "profile_123", subject: "user_123", displayName: "Ada Coach", avatarURL: stringPtr("https://cdn.example.com/avatar.png#token"), createdAt: now, updatedAt: now, code: ErrorCodeInvalidAvatarURL},
		{name: "zero created", id: "profile_123", subject: "user_123", displayName: "Ada Coach", avatarURL: &validAvatar, updatedAt: now, code: ErrorCodeInvalidTimestamp},
		{name: "zero updated", id: "profile_123", subject: "user_123", displayName: "Ada Coach", avatarURL: &validAvatar, createdAt: now, code: ErrorCodeInvalidTimestamp},
		{name: "updated before created", id: "profile_123", subject: "user_123", displayName: "Ada Coach", avatarURL: &validAvatar, createdAt: now, updatedAt: now.Add(-time.Nanosecond), code: ErrorCodeInvalidTimestamp},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			author, err := NewProfile(test.id, test.subject, test.displayName, test.avatarURL, test.createdAt, test.updatedAt)
			if err == nil {
				t.Fatalf("expected validation error, got profile %#v", author)
			}
			assertProfileErrorCode(t, err, test.code)
		})
	}
}

func TestWithPublicIdentityUpdatesNameAndAvatarWithoutChangingAuthority(t *testing.T) {
	t.Parallel()

	avatar := "https://cdn.example.com/old.png"
	newAvatar := "https://cdn.example.com/new.png"
	createdAt := time.Date(2026, 7, 11, 10, 0, 0, 0, time.UTC)
	updatedAt := createdAt.Add(time.Hour)
	author := mustProfile(t, "profile_123", "user_123", "Ada Coach", &avatar, createdAt, createdAt)

	next, err := author.WithPublicIdentity("  Ada Tactician  ", &newAvatar, updatedAt)
	if err != nil {
		t.Fatalf("update public identity: %v", err)
	}

	if next.ID() != author.ID() {
		t.Fatalf("expected internal id preserved, got %q", next.ID())
	}
	if next.ClerkSubject() != author.ClerkSubject() {
		t.Fatalf("expected clerk subject preserved, got %q", next.ClerkSubject())
	}
	if !next.CreatedAt().Equal(createdAt) {
		t.Fatalf("expected created timestamp preserved, got %s", next.CreatedAt())
	}
	if next.DisplayName() != "Ada Tactician" {
		t.Fatalf("expected updated display name, got %q", next.DisplayName())
	}
	if got, ok := next.AvatarURL(); !ok || got != newAvatar {
		t.Fatalf("expected updated avatar %q, got %q (ok=%v)", newAvatar, got, ok)
	}
	if !next.UpdatedAt().Equal(updatedAt) {
		t.Fatalf("expected updated timestamp %s, got %s", updatedAt, next.UpdatedAt())
	}
}

func TestProfileErrorUsesStableSafeCode(t *testing.T) {
	t.Parallel()

	err := profileError("create", ErrorCodeInvalidDisplayName)
	if err.Error() != "profile create: invalid_display_name" {
		t.Fatalf("unexpected safe error message %q", err.Error())
	}
	if !errors.Is(err, ErrInvalidDisplayName) {
		t.Fatal("expected errors.Is to match profile error code")
	}
	if errors.Is(err, ErrInvalidAvatarURL) {
		t.Fatal("different profile error codes must not match")
	}
}

func mustProfile(t *testing.T, id string, subject string, name string, avatar *string, createdAt time.Time, updatedAt time.Time) Profile {
	t.Helper()

	author, err := NewProfile(id, subject, name, avatar, createdAt, updatedAt)
	if err != nil {
		t.Fatalf("create profile fixture: %v", err)
	}
	return author
}

func assertProfileErrorCode(t *testing.T, err error, expected ErrorCode) {
	t.Helper()

	var profileErr *Error
	if !errors.As(err, &profileErr) {
		t.Fatalf("expected profile Error, got %T: %v", err, err)
	}
	if profileErr.Code != expected {
		t.Fatalf("expected profile error code %q, got %q", expected, profileErr.Code)
	}
}

func stringPtr(value string) *string {
	return &value
}
