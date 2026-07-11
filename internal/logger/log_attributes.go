package logger

import "log/slog"

func LogError(err error) slog.Attr {
	return slog.Any("error", err)
}
