package sugared

import "go.uber.org/zap"

func check(sugar *zap.SugaredLogger, password string, sessionID string) {
	sugar.Infow("password" + password)                         // want `log message may contain sensitive data`
	sugar.Infow("Request started")                             // want `log message must start with a lowercase letter`
	sugar.Infow("user authenticated", "session_id", sessionID) // want `structured log field may contain sensitive data`
	sugar.Warnw("request failed?!")                            // want `log message must not contain special symbols or emoji`
}
