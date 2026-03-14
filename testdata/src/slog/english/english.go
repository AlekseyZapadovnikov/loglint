package english

import "log/slog"

func check() {
	slog.Info("запуск сервера") // want `log message must contain only English text`
	slog.Info("сервер запущен") // want `log message must contain only English text`
	slog.Info("server started")
}
