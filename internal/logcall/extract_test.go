package logcall

import (
	"errors"
	"testing"

	"golang.org/x/tools/go/analysis"
)

func TestExtract_InputValidation(t *testing.T) {
	t.Parallel()

	if _, err := Extract(nil, nil); err == nil {
		t.Fatal("expected error for nil pass")
	}

	pass := &analysis.Pass{}
	if _, err := Extract(pass, nil); err == nil {
		t.Fatal("expected error for nil call expression")
	}
}

func TestExtract_NotLogCall(t *testing.T) {
	t.Parallel()

	checked := mustCheckSource(t, `
package p

import "fmt"

func f() {
	fmt.Println("request started")
}
`)

	calls := callExpressions(checked.file)
	if len(calls) != 1 {
		t.Fatalf("unexpected calls count: got %d, want %d", len(calls), 1)
	}

	pass := &analysis.Pass{TypesInfo: checked.info}
	_, err := Extract(pass, calls[0])
	if !errors.Is(err, ErrNotLogCall) {
		t.Fatalf("expected ErrNotLogCall, got %v", err)
	}
}

func TestExtract_SlogPackageAndMethodCalls(t *testing.T) {
	t.Parallel()

	checked := mustCheckSource(t, `
package p

import "log/slog"

func f(logger *slog.Logger, token string, key string, value any) {
	slog.Info("request started")
	logger.Warn("password: " + token, "token", token, key, value)
}
`)

	pass := &analysis.Pass{TypesInfo: checked.info}

	infoCalls := callsByMethod(checked.file, "Info")
	if len(infoCalls) != 1 {
		t.Fatalf("unexpected Info calls count: got %d, want %d", len(infoCalls), 1)
	}

	infoRecord, err := Extract(pass, infoCalls[0])
	if err != nil {
		t.Fatalf("extract Info call: %v", err)
	}

	if infoRecord.Kind != LoggerSlog {
		t.Fatalf("unexpected logger kind: got %q, want %q", infoRecord.Kind, LoggerSlog)
	}
	if infoRecord.Level != LevelInfo {
		t.Fatalf("unexpected level: got %q, want %q", infoRecord.Level, LevelInfo)
	}
	if got, ok := infoRecord.Message.StaticText(); !ok || got != "request started" {
		t.Fatalf("unexpected static message: got %q, ok=%v", got, ok)
	}

	warnCalls := callsByMethod(checked.file, "Warn")
	if len(warnCalls) != 1 {
		t.Fatalf("unexpected Warn calls count: got %d, want %d", len(warnCalls), 1)
	}

	warnRecord, err := Extract(pass, warnCalls[0])
	if err != nil {
		t.Fatalf("extract Warn call: %v", err)
	}

	if warnRecord.Kind != LoggerSlog {
		t.Fatalf("unexpected logger kind: got %q, want %q", warnRecord.Kind, LoggerSlog)
	}
	if warnRecord.Level != LevelWarn {
		t.Fatalf("unexpected level: got %q, want %q", warnRecord.Level, LevelWarn)
	}
	if !warnRecord.Message.HasDynamicParts() {
		t.Fatal("expected dynamic message for Warn call")
	}
	if prefix := warnRecord.Message.StaticPrefix(); prefix != "password: " {
		t.Fatalf("unexpected static prefix: got %q, want %q", prefix, "password: ")
	}

	if len(warnRecord.Fields) != 2 {
		t.Fatalf("unexpected fields count: got %d, want %d", len(warnRecord.Fields), 2)
	}

	if !warnRecord.Fields[0].HasStaticKey() || warnRecord.Fields[0].Key != "token" {
		t.Fatalf("unexpected first field key: %+v", warnRecord.Fields[0])
	}
	if warnRecord.Fields[1].HasStaticKey() {
		t.Fatalf("expected dynamic key for second field, got %+v", warnRecord.Fields[1])
	}
}

func TestMatchesPackagePath(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		actual   string
		expected string
		want     bool
	}{
		{
			name:     "exact match",
			actual:   "log/slog",
			expected: "log/slog",
			want:     true,
		},
		{
			name:     "vendor path match",
			actual:   "example.com/mod/vendor/log/slog",
			expected: "log/slog",
			want:     true,
		},
		{
			name:     "different package",
			actual:   "fmt",
			expected: "log/slog",
			want:     false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := matchesPackagePath(tt.actual, tt.expected)
			if got != tt.want {
				t.Fatalf("matchesPackagePath(%q, %q) = %v, want %v", tt.actual, tt.expected, got, tt.want)
			}
		})
	}
}
