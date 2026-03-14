package ok

import "log/slog"

func check(logger *slog.Logger, userID int) {
	slog.Info("request started")
	logger.Info("token validated")
	slog.Info("request", "user_id", userID)
	slog.Info("password changed successfully")
	slog.Info("request_id=123")
}
