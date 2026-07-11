package realtime

import (
	"errors"
	"testing"
)

func TestErrorExposesStableCodeAndSupportsErrorsIs(t *testing.T) {
	t.Parallel()

	err := operationError("open", ErrorCodeRegistryClosed)
	if err.Error() != "realtime open: registry_closed" {
		t.Fatalf("unexpected safe error message %q", err.Error())
	}
	if !errors.Is(err, ErrRegistryClosed) {
		t.Fatal("expected registry-closed errors to match by stable code")
	}
	if errors.Is(err, operationError("open", ErrorCodeConnectionLimit)) {
		t.Fatal("different error codes must not match")
	}
}

func TestErrorWithoutOperationUsesStableCode(t *testing.T) {
	t.Parallel()

	if ErrRegistryClosed.Error() != "realtime: registry_closed" {
		t.Fatalf("unexpected sentinel error message %q", ErrRegistryClosed.Error())
	}
}
