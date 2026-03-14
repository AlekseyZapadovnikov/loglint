package sensitivefield

import "log/slog"

func check(token string, sessionID string) {
	slog.Info("user authenticated", "token", token)          // want `structured log field may contain sensitive data`
	slog.Info("user authenticated", "token ", token)         // want `structured log field may contain sensitive data`
	slog.Info("user authenticated", "token = ", token)       // want `structured log field may contain sensitive data`
	slog.Info("user authenticated", "session_id", sessionID) // want `structured log field may contain sensitive data`
}
