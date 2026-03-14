package logger

import "go.uber.org/zap"

func check(logger *zap.Logger, token string, sessionID string) {
	logger.Info("Starting server")                                // want `log message must start with a lowercase letter`
	logger.Info("запуск сервера")                                 // want `log message must contain only English text`
	logger.Info("request started!")                               // want `log message must not contain special symbols or emoji`
	logger.Info("user authenticated", zap.String("token", token)) // want `structured log field may contain sensitive data`
	logger.Warn("auth", zap.String("session_id", sessionID))      // want `structured log field may contain sensitive data`
}
