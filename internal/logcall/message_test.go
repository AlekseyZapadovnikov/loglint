package logcall

import "testing"

func TestExtractMessage_StaticLiteral(t *testing.T) {
	t.Parallel()

	checked := mustCheckSource(t, `
package p
func f() {
	_ = "request started"
}
`)

	msg := extractMessage(checked.info, firstAssignRHS(t, checked.file))
	if !msg.IsStatic() {
		t.Fatal("expected static message")
	}

	text, ok := msg.StaticText()
	if !ok {
		t.Fatal("expected static text")
	}
	if text != "request started" {
		t.Fatalf("unexpected static text: got %q, want %q", text, "request started")
	}
}

func TestExtractMessage_StaticConstantConcatenation(t *testing.T) {
	t.Parallel()

	checked := mustCheckSource(t, `
package p
func f() {
	_ = "request " + "started"
}
`)

	msg := extractMessage(checked.info, firstAssignRHS(t, checked.file))
	if !msg.IsStatic() {
		t.Fatal("expected static message for constant concatenation")
	}

	text, ok := msg.StaticText()
	if !ok {
		t.Fatal("expected static text")
	}
	if text != "request started" {
		t.Fatalf("unexpected static text: got %q, want %q", text, "request started")
	}
}

func TestExtractMessage_DynamicConcatenation(t *testing.T) {
	t.Parallel()

	checked := mustCheckSource(t, `
package p
func f(password string) {
	_ = "password: " + password + " leaked"
}
`)

	msg := extractMessage(checked.info, firstAssignRHS(t, checked.file))
	if !msg.HasDynamicParts() {
		t.Fatal("expected dynamic parts")
	}

	if prefix := msg.StaticPrefix(); prefix != "password: " {
		t.Fatalf("unexpected static prefix: got %q, want %q", prefix, "password: ")
	}

	if _, ok := msg.StaticText(); ok {
		t.Fatal("expected non-static message")
	}
}

func TestExtractMessage_MergesAdjacentLiterals(t *testing.T) {
	t.Parallel()

	checked := mustCheckSource(t, `
package p
func f(token string) {
	_ = "pass" + "word: " + token
}
`)

	msg := extractMessage(checked.info, firstAssignRHS(t, checked.file))
	if prefix := msg.StaticPrefix(); prefix != "password: " {
		t.Fatalf("unexpected static prefix: got %q, want %q", prefix, "password: ")
	}
}
