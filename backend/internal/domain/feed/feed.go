package feed

import (
	"encoding/base64"
	"encoding/json"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/SourceSenseiTheRealOne/coach-connect/mvp-v2/backend/internal/domain/profile"
	"github.com/google/uuid"
)

const (
	MaxPostBodyRunes    = 2000
	MaxCommentBodyRunes = 1000
	DefaultLimit        = 20
	MaxLimit            = 50
)

type AuthorSummary struct {
	ID          string
	DisplayName string
	AvatarURL   *string
}

type ImageMedia struct {
	ID        string
	URL       string
	AltText   string
	MimeType  string
	Width     int
	Height    int
	CreatedAt time.Time
}

type Post struct {
	ID             string
	Author         AuthorSummary
	Body           string
	Media          *ImageMedia
	CreatedAt      time.Time
	UpdatedAt      time.Time
	LikeCount      int
	CommentCount   int
	ViewerHasLiked bool
	ViewerHasSaved bool
	ViewerOwns     bool
}

type Comment struct {
	ID         string
	PostID     string
	Author     AuthorSummary
	Body       string
	CreatedAt  time.Time
	UpdatedAt  time.Time
	ViewerOwns bool
}

type Cursor struct {
	CreatedAt time.Time `json:"createdAt"`
	ID        string    `json:"id"`
}

func NewAuthorSummary(author profile.Profile) AuthorSummary {
	var avatar *string
	if value, ok := author.AvatarURL(); ok {
		avatar = &value
	}

	return AuthorSummary{
		ID:          author.ID(),
		DisplayName: author.DisplayName(),
		AvatarURL:   avatar,
	}
}

func EncodeCursor(cursor Cursor) (string, error) {
	if _, err := ValidateID("encode_cursor", cursor.ID); err != nil || cursor.CreatedAt.IsZero() {
		return "", feedError("encode_cursor", ErrorCodeInvalidCursor)
	}

	payload, err := json.Marshal(cursor)
	if err != nil {
		return "", feedError("encode_cursor", ErrorCodeInvalidCursor)
	}
	return base64.RawURLEncoding.EncodeToString(payload), nil
}

func DecodeCursor(value string) (*Cursor, error) {
	if strings.TrimSpace(value) == "" {
		return nil, nil
	}
	if len(value) > 512 {
		return nil, feedError("decode_cursor", ErrorCodeInvalidCursor)
	}

	payload, err := base64.RawURLEncoding.DecodeString(value)
	if err != nil {
		return nil, feedError("decode_cursor", ErrorCodeInvalidCursor)
	}

	var cursor Cursor
	if json.Unmarshal(payload, &cursor) != nil || cursor.CreatedAt.IsZero() {
		return nil, feedError("decode_cursor", ErrorCodeInvalidCursor)
	}
	validID, err := ValidateID("decode_cursor", cursor.ID)
	if err != nil {
		return nil, feedError("decode_cursor", ErrorCodeInvalidCursor)
	}

	cursor.ID = validID
	cursor.CreatedAt = cursor.CreatedAt.UTC()
	return &cursor, nil
}

func ValidateID(operation string, value string) (string, error) {
	trimmed := strings.TrimSpace(value)
	parsed, err := uuid.Parse(trimmed)
	if err != nil || parsed.String() != strings.ToLower(trimmed) {
		return "", feedError(operation, ErrorCodeInvalidID)
	}
	return parsed.String(), nil
}

func ValidateBody(operation string, body string, maxRunes int) (string, error) {
	trimmed := strings.TrimSpace(body)
	if trimmed == "" || !utf8.ValidString(trimmed) || len([]rune(trimmed)) > maxRunes || containsControlExceptWhitespace(trimmed) {
		return "", feedError(operation, ErrorCodeInvalidBody)
	}
	return trimmed, nil
}

func ValidateLimit(limit int) (int, error) {
	if limit == 0 {
		return DefaultLimit, nil
	}
	if limit < 1 || limit > MaxLimit {
		return 0, feedError("limit", ErrorCodeInvalidLimit)
	}
	return limit, nil
}

func containsControlExceptWhitespace(value string) bool {
	for _, char := range value {
		if unicode.IsControl(char) && char != '\n' && char != '\r' && char != '\t' {
			return true
		}
	}
	return false
}
