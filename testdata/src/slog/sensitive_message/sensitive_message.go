package sensitivemessage

import "log/slog"

func check(password string, token string) {
	slog.Info("password: " + password)  // want `log message may contain sensitive data`
	slog.Info("password" + password)    // want `log message may contain sensitive data`
	slog.Info("password " + password)   // want `log message may contain sensitive data`
	slog.Info("password = " + password) // want `log message may contain sensitive data`
	slog.Info("token: " + token)        // want `log message may contain sensitive data`
}
