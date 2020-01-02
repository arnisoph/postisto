package log

import (
	"fmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var log struct {
	level string
	json  bool

	logger *zap.SugaredLogger
}

func init() {
	if err := InitWithConfig("info", false); err != nil {
		panic(fmt.Sprintf("Failed to initialize logger: %q\n\nThat should not happen. Please call a doctor.", err.Error()))
	}
}

func InitWithConfig(logLevel string, logJson bool) error {
	if log.logger != nil {
		Debug("Logger already initialized. Will re-initialize now..")
	}

	cfg := zap.NewProductionConfig()

	switch logLevel {
	case "trace":
		cfg.Level = zap.NewAtomicLevelAt(zapcore.DebugLevel)
	case "debug":
		cfg.Level = zap.NewAtomicLevelAt(zapcore.DebugLevel)
	case "info":
		cfg.Level = zap.NewAtomicLevelAt(zapcore.InfoLevel)
		cfg.Development = false
	case "error":
		cfg.Level = zap.NewAtomicLevelAt(zapcore.ErrorLevel)
		cfg.Development = false
	default:
		return fmt.Errorf("unknown log level %q", logLevel)
	}

	if logJson {
		cfg.Encoding = "json"
	} else {
		cfg.Encoding = "console"
	}

	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	cfg.DisableStacktrace = true

	rawLogger, err := cfg.Build()

	if err != nil {
		return err
	}
	defer rawLogger.Sync()

	log.level = logLevel
	log.json = logJson
	log.logger = rawLogger.WithOptions(zap.AddCallerSkip(1)).Sugar() // pkg variable
	Debug("logging successfully initialized")

	return err
}

func GetLogLevel() string {
	return log.level
}

func Panic(msg string, err error) {
	log.logger.With("err", err).Panic(msg)
}

func Panicw(msg string, err error, context ...interface{}) {
	context = append(context, "err")
	context = append(context, err)
	log.logger.With(context...).Panic(msg)
}

func Fatal(msg string, err error) {
	log.logger.With("err", err).Fatal(msg)
}

func Fatalw(msg string, err error, context ...interface{}) {
	context = append(context, "err")
	context = append(context, err)
	log.logger.With(context...).Fatal(msg)
}

func Error(msg string, err error) {
	log.logger.With("err", err).Error(msg)
}

func Errorw(msg string, err error, context ...interface{}) {
	context = append(context, "err")
	context = append(context, err)
	log.logger.With(context...).Error(msg)
}

func Info(msg string) {
	log.logger.Info(msg)
}

func Infow(msg string, context ...interface{}) {
	log.logger.With(context...).Info(msg)
}

func Debug(msg string) {
	log.logger.Debug(msg)
}

func Debugw(msg string, context ...interface{}) {
	log.logger.With(context...).Debug(msg)
}
