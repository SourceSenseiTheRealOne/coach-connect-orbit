package profile

import (
	"net/url"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

const (
	maxIDLength           = 128
	maxClerkSubjectLength = 256
	maxDisplayNameLength  = 80
	maxAvatarURLLength    = 2048
)

type Profile struct {
	id           string
	clerkSubject string
	displayName  string
	avatarURL    *string
	createdAt    time.Time
	updatedAt    time.Time
}

func NewProfile(id string, clerkSubject string, displayName string, avatarURL *string, createdAt time.Time, updatedAt time.Time) (Profile, error) {
	validID, err := validateOpaque("create", id, maxIDLength, ErrorCodeInvalidID)
	if err != nil {
		return Profile{}, err
	}

	validSubject, err := validateOpaque("create", clerkSubject, maxClerkSubjectLength, ErrorCodeInvalidClerkSubject)
	if err != nil {
		return Profile{}, err
	}

	validName, err := validateDisplayName("create", displayName)
	if err != nil {
		return Profile{}, err
	}

	validAvatar, err := validateAvatarURL("create", avatarURL)
	if err != nil {
		return Profile{}, err
	}

	if err := validateTimestamps("create", createdAt, updatedAt); err != nil {
		return Profile{}, err
	}

	return Profile{
		id:           validID,
		clerkSubject: validSubject,
		displayName:  validName,
		avatarURL:    validAvatar,
		createdAt:    createdAt.UTC(),
		updatedAt:    updatedAt.UTC(),
	}, nil
}

func (profile Profile) ID() string {
	return profile.id
}

func (profile Profile) ClerkSubject() string {
	return profile.clerkSubject
}

func (profile Profile) DisplayName() string {
	return profile.displayName
}

func (profile Profile) AvatarURL() (string, bool) {
	if profile.avatarURL == nil {
		return "", false
	}

	return *profile.avatarURL, true
}

func (profile Profile) CreatedAt() time.Time {
	return profile.createdAt
}

func (profile Profile) UpdatedAt() time.Time {
	return profile.updatedAt
}

func (profile Profile) WithPublicIdentity(displayName string, avatarURL *string, updatedAt time.Time) (Profile, error) {
	validName, err := validateDisplayName("update_public_identity", displayName)
	if err != nil {
		return Profile{}, err
	}

	validAvatar, err := validateAvatarURL("update_public_identity", avatarURL)
	if err != nil {
		return Profile{}, err
	}

	if err := validateTimestamps("update_public_identity", profile.createdAt, updatedAt); err != nil {
		return Profile{}, err
	}

	next := profile
	next.displayName = validName
	next.avatarURL = validAvatar
	next.updatedAt = updatedAt.UTC()
	return next, nil
}

func (profile Profile) PublicIdentityMatches(displayName string, avatarURL *string) (bool, error) {
	validName, err := validateDisplayName("compare_public_identity", displayName)
	if err != nil {
		return false, err
	}

	validAvatar, err := validateAvatarURL("compare_public_identity", avatarURL)
	if err != nil {
		return false, err
	}

	return profile.displayName == validName && sameOptionalString(profile.avatarURL, validAvatar), nil
}

func validateOpaque(operation string, value string, maxLength int, code ErrorCode) (string, error) {
	if value == "" || value != strings.TrimSpace(value) || len(value) > maxLength || containsControl(value) || !utf8.ValidString(value) {
		return "", profileError(operation, code)
	}

	return value, nil
}

func validateDisplayName(operation string, value string) (string, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" || len([]rune(trimmed)) > maxDisplayNameLength || containsControl(trimmed) || !utf8.ValidString(trimmed) {
		return "", profileError(operation, ErrorCodeInvalidDisplayName)
	}

	return trimmed, nil
}

func validateAvatarURL(operation string, value *string) (*string, error) {
	if value == nil {
		return nil, nil
	}

	raw := strings.TrimSpace(*value)
	if raw == "" || raw != *value || len(raw) > maxAvatarURLLength || containsControl(raw) || !utf8.ValidString(raw) {
		return nil, profileError(operation, ErrorCodeInvalidAvatarURL)
	}

	parsed, err := url.Parse(raw)
	if err != nil || parsed.Scheme != "https" || parsed.Host == "" || parsed.User != nil || parsed.Fragment != "" {
		return nil, profileError(operation, ErrorCodeInvalidAvatarURL)
	}

	copied := raw
	return &copied, nil
}

func validateTimestamps(operation string, createdAt time.Time, updatedAt time.Time) error {
	if createdAt.IsZero() || updatedAt.IsZero() || updatedAt.Before(createdAt) {
		return profileError(operation, ErrorCodeInvalidTimestamp)
	}

	return nil
}

func containsControl(value string) bool {
	for _, char := range value {
		if unicode.IsControl(char) {
			return true
		}
	}
	return false
}

func sameOptionalString(left *string, right *string) bool {
	if left == nil || right == nil {
		return left == right
	}

	return *left == *right
}
