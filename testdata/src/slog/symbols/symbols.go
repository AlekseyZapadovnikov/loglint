package symbols

import "log/slog"

func check() {
	slog.Info("request started!")    // want `log message must not contain special symbols or emoji`
	slog.Info("request started...")  // want `log message must not contain special symbols or emoji`
	slog.Info("request started 🤣🤣🤣") // want `log message must not contain special symbols or emoji`
	slog.Info("request failed??")    // want `log message must not contain special symbols or emoji`
}
