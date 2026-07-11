package logger

import (
	"context"
	"log/slog"
)

type contextHandler struct {
	inner slog.Handler
}

func NewContextHandler(inner slog.Handler) slog.Handler { // nolint:ireturn
	return &contextHandler{inner: inner}
}

func (h *contextHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.inner.Enabled(ctx, level)
}

func (h *contextHandler) Handle(ctx context.Context, r slog.Record) error {
	return h.inner.Handle(ctx, r)
}

func (h *contextHandler) WithAttrs(attrs []slog.Attr) slog.Handler { // nolint:ireturn
	return &contextHandler{inner: h.inner.WithAttrs(attrs)}
}

func (h *contextHandler) WithGroup(name string) slog.Handler { // nolint:ireturn
	return &contextHandler{inner: h.inner.WithGroup(name)}
}
