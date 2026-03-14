package ok

import "go.uber.org/zap"

func check(logger *zap.Logger, sugar *zap.SugaredLogger, userID int) {
	logger.Info("request started", zap.Int("user_id", userID))
	sugar.Infow("request started", "user_id", userID)
	logger.Warn("token validated")
}
