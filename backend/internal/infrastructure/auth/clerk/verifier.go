package clerk

import (
	"context"
	"strings"
	"sync"
	"time"

	applicationauth "github.com/SourceSenseiTheRealOne/coach-connect/mvp-v2/backend/internal/application/auth"
	sdk "github.com/clerk/clerk-sdk-go/v2"
	clerkjwt "github.com/clerk/clerk-sdk-go/v2/jwt"
	clerkuser "github.com/clerk/clerk-sdk-go/v2/user"
)

const publicIdentityCacheTTL = 5 * time.Minute

type cachedUser struct {
	user      *sdk.User
	expiresAt time.Time
}

type Verifier struct {
	getUser func(context.Context, string) (*sdk.User, error)
	now     func() time.Time

	mu    sync.RWMutex
	users map[string]cachedUser
}

func NewVerifier(secret string) (*Verifier, error) {
	if strings.TrimSpace(secret) == "" {
		return nil, applicationauth.ErrInvalidConfig
	}
	sdk.SetKey(secret)
	return &Verifier{
		getUser: clerkuser.Get,
		now:     time.Now,
		users:   make(map[string]cachedUser),
	}, nil
}

func (verifier *Verifier) VerifyBearer(ctx context.Context, token string) (applicationauth.VerifiedIdentity, error) {
	if verifier == nil || strings.TrimSpace(token) == "" {
		return applicationauth.VerifiedIdentity{}, applicationauth.ErrInvalidBearerToken
	}
	claims, err := clerkjwt.Verify(ctx, &clerkjwt.VerifyParams{
		Token: token,
		CustomClaimsConstructor: func(context.Context) any {
			return &map[string]any{}
		},
	})
	if err != nil || claims.Subject == "" {
		return applicationauth.VerifiedIdentity{}, applicationauth.ErrInvalidBearerToken
	}

	custom := map[string]any{}
	if values, ok := claims.Custom.(*map[string]any); ok && values != nil {
		custom = *values
	}
	identity := identityFrom(custom, claims.Subject, nil)
	if identity.DisplayName != "" && identity.AvatarURL != nil {
		return identity, nil
	}

	return identityFrom(custom, claims.Subject, verifier.cachedUser(ctx, claims.Subject)), nil
}

func (verifier *Verifier) cachedUser(ctx context.Context, subject string) *sdk.User {
	if verifier.getUser == nil || verifier.now == nil {
		return nil
	}
	now := verifier.now()
	verifier.mu.RLock()
	cached, ok := verifier.users[subject]
	verifier.mu.RUnlock()
	if ok && now.Before(cached.expiresAt) {
		return cached.user
	}

	user, err := verifier.getUser(ctx, subject)
	if err != nil || user == nil || user.ID != subject {
		return nil
	}
	verifier.mu.Lock()
	verifier.users[subject] = cachedUser{user: user, expiresAt: now.Add(publicIdentityCacheTTL)}
	verifier.mu.Unlock()
	return user
}

func identityFrom(custom map[string]any, subject string, user *sdk.User) applicationauth.VerifiedIdentity {
	identity := applicationauth.VerifiedIdentity{Subject: subject}
	if name, ok := custom["name"].(string); ok && strings.TrimSpace(name) != "" {
		identity.DisplayName = strings.TrimSpace(name)
	}
	if imageURL, ok := custom["image_url"].(string); ok && strings.TrimSpace(imageURL) != "" {
		value := strings.TrimSpace(imageURL)
		identity.AvatarURL = &value
	}
	if user == nil {
		return identity
	}

	if identity.DisplayName == "" {
		identity.DisplayName = userDisplayName(user)
	}
	if identity.AvatarURL == nil && user.ImageURL != nil && strings.TrimSpace(*user.ImageURL) != "" {
		value := strings.TrimSpace(*user.ImageURL)
		identity.AvatarURL = &value
	}
	return identity
}

func userDisplayName(user *sdk.User) string {
	parts := make([]string, 0, 2)
	if user.FirstName != nil && strings.TrimSpace(*user.FirstName) != "" {
		parts = append(parts, strings.TrimSpace(*user.FirstName))
	}
	if user.LastName != nil && strings.TrimSpace(*user.LastName) != "" {
		parts = append(parts, strings.TrimSpace(*user.LastName))
	}
	if name := strings.Join(parts, " "); name != "" {
		return name
	}
	if user.Username != nil && strings.TrimSpace(*user.Username) != "" {
		return strings.TrimSpace(*user.Username)
	}
	return ""
}
