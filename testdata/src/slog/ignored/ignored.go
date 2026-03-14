package ignored

import "log/slog"

func check() {
	slog.Info("request started", slog.String("token", "value"))
	slog.Warn("request started", slog.Attr{})
}
