package rules_test

import (
	"go/ast"
	"testing"

	"github.com/AlekseyZapadovnikov/loglint/internal/logcall"
	"github.com/AlekseyZapadovnikov/loglint/internal/rules"
)

const (
	lowercaseViolationMessage      = "log message must start with a lowercase letter"
	englishViolationMessage        = "log message must contain only English text"
	symbolsViolationMessage        = "log message must not contain special symbols or emoji"
	sensitiveMessageViolation      = "log message may contain sensitive data"
	sensitiveFieldViolationMessage = "structured log field may contain sensitive data"
)

func recordWithStaticMessage(text string) logcall.Record {
	expr := &ast.BasicLit{}

	return logcall.Record{
		Message: logcall.NewStaticMessage(expr, text),
	}
}

func recordWithDynamicMessage(prefix string) logcall.Record {
	expr := &ast.BinaryExpr{}

	return logcall.Record{
		Message: logcall.NewMessageFromParts(expr, []logcall.MessagePart{
			{
				Kind: logcall.MessagePartLiteral,
				Expr: &ast.BasicLit{},
				Text: prefix,
			},
			{
				Kind: logcall.MessagePartDynamic,
				Expr: ast.NewIdent("value"),
			},
		}),
	}
}

func recordWithField(key string) logcall.Record {
	return logcall.Record{
		Fields: []logcall.Field{
			{
				Key:       key,
				KeyExpr:   &ast.BasicLit{},
				ValueExpr: ast.NewIdent("value"),
				StaticKey: true,
			},
		},
	}
}

func recordWithStaticMessageAndField(message, key string) logcall.Record {
	record := recordWithStaticMessage(message)
	record.Fields = append(record.Fields, recordWithField(key).Fields...)

	return record
}

func stringSlicePtr(values ...string) *[]string {
	copied := append([]string(nil), values...)
	return &copied
}

func requireViolationCount(t *testing.T, got []rules.Violation, want int) {
	t.Helper()

	if len(got) != want {
		t.Fatalf("unexpected violations count: got %d, want %d", len(got), want)
	}
}

func requireSingleViolation(t *testing.T, got []rules.Violation, wantRule rules.ID, wantMessage string) {
	t.Helper()

	requireViolationCount(t, got, 1)

	if got[0].Rule != wantRule {
		t.Fatalf("unexpected violation rule: got %q, want %q", got[0].Rule, wantRule)
	}

	if got[0].Message != wantMessage {
		t.Fatalf("unexpected violation message: got %q, want %q", got[0].Message, wantMessage)
	}

	if got[0].Expr == nil {
		t.Fatal("expected violation expr to be set")
	}
}

func requireViolationRules(t *testing.T, got []rules.Violation, want ...rules.ID) {
	t.Helper()

	requireViolationCount(t, got, len(want))

	for i := range want {
		if got[i].Rule != want[i] {
			t.Fatalf("unexpected violation rule at index %d: got %q, want %q", i, got[i].Rule, want[i])
		}

		if got[i].Expr == nil {
			t.Fatalf("expected violation expr to be set at index %d", i)
		}
	}
}
