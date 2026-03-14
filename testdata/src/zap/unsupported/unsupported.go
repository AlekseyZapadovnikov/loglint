package unsupported

import "go.uber.org/zap"

func check(logger *zap.Logger, sugar *zap.SugaredLogger) {
	logger.Fatal("Starting server")
	logger.DPanic("request failed!")
	sugar.Info("Request started")
	sugar.Infof("request failed %s", "now")
}
