// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: Copyright 2025-2026 WoozyMasta
// Source: https://github.com/WoozyMasta/a2s

package ping

import (
	"context"
	"errors"
	"testing"
	"time"
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

func TestFormatMilliseconds(t *testing.T) {
	if got := formatMilliseconds(56*time.Millisecond + 377600*time.Nanosecond); got != "56.3776" {
		t.Fatalf("formatMilliseconds() = %q, want 56.3776", got)
	}
}

func TestSuccessPercentage(t *testing.T) {
	for _, test := range []struct {
		received int
		failed   int
		want     int
	}{
		{received: 4, failed: 0, want: 100},
		{received: 3, failed: 1, want: 75},
		{received: 0, failed: 0, want: 0},
	} {
		if got := successPercentage(test.received, test.failed); got != test.want {
			t.Fatalf("successPercentage(%d, %d) = %d, want %d", test.received, test.failed, got, test.want)
		}
	}
}
