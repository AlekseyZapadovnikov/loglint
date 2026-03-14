package logcall

import (
	"go/ast"
	"strings"
)

// LoggerKind identifies the logger family that produced the record.
type LoggerKind string

const (
	LoggerUnknown    LoggerKind = ""
	LoggerSlog       LoggerKind = "slog"
	LoggerZap        LoggerKind = "zap"
	LoggerZapSugared LoggerKind = "zap_sugared"
)

// Level is a normalized log level.
type Level string

const (
	LevelUnknown Level = ""
	LevelDebug   Level = "debug"
	LevelInfo    Level = "info"
	LevelWarn    Level = "warn"
	LevelError   Level = "error"
)

// MessagePartKind describes one piece of a log message.
type MessagePartKind uint8

const (
	MessagePartUnknown MessagePartKind = iota
	MessagePartLiteral
	MessagePartDynamic
)

// MessagePart is one fragment of the original message expression.
// For example, for `"token: " + token` the first part is literal,
// the second part is dynamic.
type MessagePart struct {
	Kind MessagePartKind
	Expr ast.Expr
	Text string
}

// IsStatic reports whether the part is known at analysis time.
func (p MessagePart) IsStatic() bool {
	return p.Kind == MessagePartLiteral
}

// Message represents the extracted message argument of a log call.
type Message struct {
	// Expr is the original AST expression that produced the message.
	Expr ast.Expr

	// Parts is a normalized view of the message.
	// It allows us to handle both plain string literals and concatenations.
	Parts []MessagePart
}

// IsStatic reports whether the entire message is statically known.
func (m Message) IsStatic() bool {
	if m.Expr == nil || len(m.Parts) == 0 {
		return false
	}

	for _, part := range m.Parts {
		if !part.IsStatic() {
			return false
		}
	}

	return true
}

// HasDynamicParts reports whether the message contains runtime values.
func (m Message) HasDynamicParts() bool {
	for _, part := range m.Parts {
		if !part.IsStatic() {
			return true
		}
	}

	return false
}

// StaticText returns the full message text if it is completely static.
func (m Message) StaticText() (string, bool) {
	if !m.IsStatic() {
		return "", false
	}

	var b strings.Builder
	for _, part := range m.Parts {
		b.WriteString(part.Text)
	}

	return b.String(), true
}

// StaticPrefix returns the leading static part of the message.
// This is useful for patterns like `"password: " + password`.
func (m Message) StaticPrefix() string {
	var b strings.Builder

	for _, part := range m.Parts {
		if !part.IsStatic() {
			break
		}

		b.WriteString(part.Text)
	}

	return b.String()
}

// Field represents one structured logging field.
//
// Examples:
//
//	slog.Info("msg", "user_id", id)
//	zap.String("token", token)
type Field struct {
	Key       string
	KeyExpr   ast.Expr
	ValueExpr ast.Expr
	StaticKey bool
}

// HasStaticKey reports whether the field key is known at analysis time.
func (f Field) HasStaticKey() bool {
	return f.StaticKey && f.Key != ""
}

// Record is the normalized representation of a log call.
// Rules should work with Record instead of raw AST.
type Record struct {
	Kind   LoggerKind
	Level  Level
	Method string

	Call     *ast.CallExpr
	Selector *ast.SelectorExpr

	Message Message
	Fields  []Field
}

// HasMessage reports whether the record contains a message expression.
func (r Record) HasMessage() bool {
	return r.Message.Expr != nil
}

// NewStaticMessage creates a message from a plain static string.
func NewStaticMessage(expr ast.Expr, text string) Message {
	return Message{
		Expr: expr,
		Parts: []MessagePart{
			{
				Kind: MessagePartLiteral,
				Expr: expr,
				Text: text,
			},
		},
	}
}

// NewDynamicMessage creates a message that cannot be resolved statically.
func NewDynamicMessage(expr ast.Expr) Message {
	return Message{
		Expr: expr,
		Parts: []MessagePart{
			{
				Kind: MessagePartDynamic,
				Expr: expr,
			},
		},
	}
}

// NewMessageFromParts creates a message from pre-parsed parts.
// The parts slice is copied to keep the model immutable from the caller's view.
func NewMessageFromParts(expr ast.Expr, parts []MessagePart) Message {
	cloned := make([]MessagePart, len(parts))
	copy(cloned, parts)

	return Message{
		Expr:  expr,
		Parts: cloned,
	}
}
