package metrics

import (
	"context"
	"log/slog"
	"strings"
)

var sensitiveKeys = map[string]bool{
	"password":      true,
	"passcode":      true,
	"token":         true,
	"access_token":  true,
	"refresh_token": true,
	"secret":        true,
	"jwt":           true,
	"authorization": true,
	"cookie":        true,
}

// RedactingHandler wraps an slog.Handler to redact sensitive keys like passwords and tokens.
type RedactingHandler struct {
	slog.Handler
}

// NewRedactingHandler creates an slog.Handler that strips sensitive data.
func NewRedactingHandler(inner slog.Handler) *RedactingHandler {
	if rh, ok := inner.(*RedactingHandler); ok {
		return rh
	}
	return &RedactingHandler{Handler: inner}
}

// Handle redacts sensitive attributes before passing to the underlying handler.
func (h *RedactingHandler) Handle(ctx context.Context, r slog.Record) error {
	redactedRecord := slog.NewRecord(r.Time, r.Level, r.Message, r.PC)
	r.Attrs(func(a slog.Attr) bool {
		redactedRecord.AddAttrs(sanitizeAttr(a))
		return true
	})
	return h.Handler.Handle(ctx, redactedRecord)
}

// WithAttrs returns a new RedactingHandler with sanitized attributes.
func (h *RedactingHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	sanitized := make([]slog.Attr, len(attrs))
	for i, a := range attrs {
		sanitized[i] = sanitizeAttr(a)
	}
	return &RedactingHandler{Handler: h.Handler.WithAttrs(sanitized)}
}

// WithGroup returns a new RedactingHandler with group name.
func (h *RedactingHandler) WithGroup(name string) slog.Handler {
	return &RedactingHandler{Handler: h.Handler.WithGroup(name)}
}

func sanitizeAttr(a slog.Attr) slog.Attr {
	keyLower := strings.ToLower(a.Key)
	if sensitiveKeys[keyLower] || strings.Contains(keyLower, "password") || strings.Contains(keyLower, "secret") || strings.Contains(keyLower, "token") {
		return slog.String(a.Key, "[REDACTED]")
	}
	if a.Value.Kind() == slog.KindGroup {
		groupAttrs := a.Value.Group()
		sanitized := make([]slog.Attr, len(groupAttrs))
		for i, child := range groupAttrs {
			sanitized[i] = sanitizeAttr(child)
		}
		return slog.Group(a.Key, anySlice(sanitized)...)
	}
	return a
}

func anySlice(attrs []slog.Attr) []any {
	out := make([]any, len(attrs))
	for i, a := range attrs {
		out[i] = a
	}
	return out
}
