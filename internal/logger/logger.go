package logger

import (
	"log/slog"
	"os"
	"time"

	"github.com/lmittmann/tint"
	"github.com/zed-assistant/mcp/internal/configuration"
)

func NewLogger(appConfig *configuration.AppConfig) *slog.Logger {
	var jsonFormat bool
	var disableColor bool

	var handler slog.Handler

	if appConfig != nil {
		jsonFormat = appConfig.Logger.JsonFormat
		disableColor = appConfig.Logger.DisableColor
	}

	if jsonFormat {
		options := slog.HandlerOptions{
			AddSource: true,
		}

		handler = slog.NewJSONHandler(os.Stdout, &options)
	} else {
		options := &tint.Options{
			AddSource:  true,
			TimeFormat: time.RFC3339,
			NoColor:    disableColor,
		}

		handler = tint.NewHandler(os.Stdout, options)
	}

	return slog.New(NewContextHandler(handler))
}
