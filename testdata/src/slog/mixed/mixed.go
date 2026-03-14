package mixed

import "log/slog"

func check(logger *slog.Logger, password string, token string) {
	logger.Info("Request started")                  // want `log message must start with a lowercase letter`
	slog.Info("ошибка авторизации")                 // want `log message must contain only English text`
	slog.Warn("request failed...")                  // want `log message must not contain special symbols or emoji`
	logger.Info("password: " + password)            // want `log message may contain sensitive data`
	slog.Info("user authenticated", "token", token) // want `structured log field may contain sensitive data`
	slog.Info("request finished", "request_id", "value")
}
