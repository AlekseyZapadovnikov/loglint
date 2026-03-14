package lowercase

import "log/slog"

func check() {
	log := slog.Logger{}

	log.Info("Starting server on port 8080") // want `log message must start with a lowercase letter`
	slog.Info("Server started")              // want `log message must start with a lowercase letter`
	slog.Info("123 Starting server")         // want `log message must start with a lowercase letter`
}
