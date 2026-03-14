// Package zap provides minimal declarations for analysistest.
package zap

type Logger struct{}

type SugaredLogger struct{}

type Field struct{}

func (l *Logger) Debug(msg string, fields ...Field)  {}
func (l *Logger) Info(msg string, fields ...Field)   {}
func (l *Logger) Warn(msg string, fields ...Field)   {}
func (l *Logger) Error(msg string, fields ...Field)  {}
func (l *Logger) DPanic(msg string, fields ...Field) {}
func (l *Logger) Fatal(msg string, fields ...Field)  {}

func (s *SugaredLogger) Debugw(msg string, keysAndValues ...any) {}
func (s *SugaredLogger) Infow(msg string, keysAndValues ...any)  {}
func (s *SugaredLogger) Warnw(msg string, keysAndValues ...any)  {}
func (s *SugaredLogger) Errorw(msg string, keysAndValues ...any) {}
func (s *SugaredLogger) Info(msg string, args ...any)            {}
func (s *SugaredLogger) Infof(template string, args ...any)      {}

func String(key string, val string) Field { return Field{} }
func Int(key string, val int) Field       { return Field{} }
