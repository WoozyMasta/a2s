package ping

import (
	"context"
	"errors"
	"testing"
)

func TestFormatPingError(t *testing.T) {
	if got := formatPingError(context.DeadlineExceeded); got != "timeout" {
		t.Fatalf("formatPingError() = %q, want timeout", got)
	}

	cause := errors.New("connection refused")
	if got := formatPingError(cause); got != cause.Error() {
		t.Fatalf("formatPingError() = %q, want %q", got, cause.Error())
	}
}
