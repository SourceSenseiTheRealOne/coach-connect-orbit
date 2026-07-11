package feed

import (
	"encoding/base64"
	"errors"
	"testing"
	"time"
)

func TestCursorRoundTripAndRejectsMalformedValues(t *testing.T) {
	t.Parallel()

	timestamp := time.Date(2026, 7, 11, 3, 0, 0, 123, time.UTC)
	const cursorID = "00000000-0000-4000-8000-000000000001"
	cursor, err := EncodeCursor(Cursor{CreatedAt: timestamp, ID: cursorID})
	if err != nil {
		t.Fatalf("EncodeCursor() error = %v", err)
	}

	decoded, err := DecodeCursor(cursor)
	if err != nil {
		t.Fatalf("DecodeCursor() error = %v", err)
	}
	if decoded == nil || !decoded.CreatedAt.Equal(timestamp) || decoded.ID != cursorID {
		t.Fatalf("decoded cursor = %#v", decoded)
	}

	empty, err := DecodeCursor("")
	if err != nil || empty != nil {
		t.Fatalf("empty cursor = %#v, %v", empty, err)
	}

	if _, err := DecodeCursor("not-base64"); !errors.Is(err, ErrInvalidCursor) {
		t.Fatalf("expected invalid cursor, got %v", err)
	}
	forged := base64.RawURLEncoding.EncodeToString([]byte(`{"createdAt":"2026-07-11T03:00:00Z","id":"not-a-uuid"}`))
	if _, err := DecodeCursor(forged); !errors.Is(err, ErrInvalidCursor) {
		t.Fatalf("expected invalid cursor id, got %v", err)
	}
	if _, err := ValidateID("post", "not-a-uuid"); !errors.Is(err, ErrInvalidID) {
		t.Fatalf("expected invalid id, got %v", err)
	}
}

func TestValidateBodyTrimsAndRejectsUnsafeText(t *testing.T) {
	t.Parallel()

	body, err := ValidateBody("create", " football insight ", MaxPostBodyRunes)
	if err != nil {
		t.Fatalf("ValidateBody() error = %v", err)
	}
	if body != "football insight" {
		t.Fatalf("expected trimmed body, got %q", body)
	}

	if _, err := ValidateBody("create", " ", MaxPostBodyRunes); !errors.Is(err, ErrInvalidBody) {
		t.Fatalf("expected invalid body for blank text, got %v", err)
	}
}
