package clerk

import (
	"context"
	"errors"
	"testing"
	"time"

	applicationauth "github.com/SourceSenseiTheRealOne/coach-connect/mvp-v2/backend/internal/application/auth"
	sdk "github.com/clerk/clerk-sdk-go/v2"
)

func TestNewVerifierRequiresSecretKey(t *testing.T) {
	t.Parallel()

	verifier, err := NewVerifier(" ")
	if err == nil {
		t.Fatalf("expected invalid config, got verifier %#v", verifier)
	}
	if !errors.Is(err, applicationauth.ErrInvalidConfig) {
		t.Fatalf("expected invalid config, got %v", err)
	}
}

func TestIdentityFromUsesClerkUserWhenSessionClaimsLackPublicIdentity(t *testing.T) {
	t.Parallel()

	firstName := " Ada "
	lastName := "Coach"
	imageURL := "https://images.example.test/ada.png"
	identity := identityFrom(map[string]any{}, "user_ada", &sdk.User{
		ID:        "user_ada",
		FirstName: &firstName,
		LastName:  &lastName,
		ImageURL:  &imageURL,
	})

	if identity.Subject != "user_ada" || identity.DisplayName != "Ada Coach" {
		t.Fatalf("unexpected identity: %#v", identity)
	}
	if identity.AvatarURL == nil || *identity.AvatarURL != imageURL {
		t.Fatalf("unexpected avatar: %#v", identity.AvatarURL)
	}
}

func TestIdentityFromPrefersVerifiedSessionClaims(t *testing.T) {
	t.Parallel()

	firstName := "Backend"
	imageURL := "https://images.example.test/backend.png"
	identity := identityFrom(map[string]any{
		"name":      "Session Coach",
		"image_url": "https://images.example.test/session.png",
	}, "user_ada", &sdk.User{ID: "user_ada", FirstName: &firstName, ImageURL: &imageURL})

	if identity.DisplayName != "Session Coach" {
		t.Fatalf("display name = %q", identity.DisplayName)
	}
	if identity.AvatarURL == nil || *identity.AvatarURL != "https://images.example.test/session.png" {
		t.Fatalf("unexpected avatar: %#v", identity.AvatarURL)
	}
}

func TestIdentityFromLeavesMissingPublicNameBlankForSafeApplicationFallback(t *testing.T) {
	t.Parallel()

	identity := identityFrom(map[string]any{}, "user_private_subject", &sdk.User{ID: "user_private_subject"})

	if identity.DisplayName != "" {
		t.Fatalf("display name = %q, want blank application fallback", identity.DisplayName)
	}
}

func TestVerifierCachesPublicIdentityLookup(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 7, 11, 12, 0, 0, 0, time.UTC)
	calls := 0
	verifier := &Verifier{
		getUser: func(_ context.Context, subject string) (*sdk.User, error) {
			calls++
			return &sdk.User{ID: subject}, nil
		},
		now:   func() time.Time { return now },
		users: make(map[string]cachedUser),
	}

	if verifier.cachedUser(context.Background(), "user_ada") == nil {
		t.Fatal("first lookup returned nil")
	}
	if verifier.cachedUser(context.Background(), "user_ada") == nil {
		t.Fatal("cached lookup returned nil")
	}
	if calls != 1 {
		t.Fatalf("lookup calls = %d, want 1", calls)
	}

	now = now.Add(publicIdentityCacheTTL)
	if verifier.cachedUser(context.Background(), "user_ada") == nil {
		t.Fatal("refreshed lookup returned nil")
	}
	if calls != 2 {
		t.Fatalf("lookup calls after expiry = %d, want 2", calls)
	}
}
