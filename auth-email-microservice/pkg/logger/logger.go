package logger

import (
	"os"

	"github.com/Chuuch/ecom-microservices/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger methods interface
type Logger interface {
	InitLogger()
	Debug(args ...interface{})
	Debugf(template string, args ...interface{})
	Info(args ...interface{})
	Infof(template string, args ...interface{})
	Warn(args ...interface{})
	Warnf(template string, args ...interface{})
	Error(args ...interface{})
	Errorf(template string, args ...interface{})
	Fatal(args ...interface{})
	Fatalf(template string, args ...interface{})
	DPanic(args ...interface{})
	DPanicf(template string, args ...interface{})
}

// Logger

type apiLogger struct {
	cfg           *config.Config
	sugaredLogger *zap.SugaredLogger
}

// App logger constructor
func NewApiLogger(cfg *config.Config) *apiLogger {
	return &apiLogger{cfg: cfg}
}

// For mapping config logger to app logger levels
var loggerLevelMap = map[string]zapcore.Level{
	"debug":  zapcore.DebugLevel,
	"info":   zapcore.InfoLevel,
	"warn":   zapcore.WarnLevel,
	"error":  zapcore.ErrorLevel,
	"fatal":  zapcore.FatalLevel,
	"dpanic": zapcore.DPanicLevel,
	"panic":  zapcore.PanicLevel,
}

func (l *apiLogger) getLoggerLevel(cfg *config.Config) zapcore.Level {
	level, exist := loggerLevelMap[cfg.Logger.Level]
	if !exist {
		return zapcore.DebugLevel
	}
	return level
}

// Init logger
func (l *apiLogger) InitLogger() {
	logLevel := l.getLoggerLevel(l.cfg)

	var logWriter zapcore.WriteSyncer
	logWriter = zapcore.AddSync(os.Stderr)

	var encoderCfg zapcore.EncoderConfig
	if l.cfg.Server.Mode == "Development" {
		encoderCfg = zap.NewDevelopmentEncoderConfig()
	} else {
		encoderCfg = zap.NewProductionEncoderConfig()
	}

	var encoder zapcore.Encoder
	encoderCfg.LevelKey = "LEVEL"
	encoderCfg.CallerKey = "CALLER"
	encoderCfg.TimeKey = "TIME"
	encoderCfg.NameKey = "NAME"
	encoderCfg.MessageKey = "MESSAGE"

	if l.cfg.Logger.Encoding == "console" {
		encoder = zapcore.NewConsoleEncoder(encoderCfg)
	} else {
		encoder = zapcore.NewJSONEncoder(encoderCfg)
	}

	encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder
	core := zapcore.NewCore(encoder, logWriter, zap.NewAtomicLevelAt(logLevel))
	logger := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))

	l.sugaredLogger = logger.Sugar()

	if err := l.sugaredLogger.Sync(); err != nil {
		l.sugaredLogger.Error(err)
	}
}

func (l *apiLogger) Fatal(args ...interface{}) {
	l.sugaredLogger.Fatal(args...)
}

func (l *apiLogger) Fatalf(template string, args ...interface{}) {
	l.sugaredLogger.Fatalf(template, args...)
}

func (l *apiLogger) DPanic(args ...interface{}) {
	l.sugaredLogger.DPanic(args...)
}

func (l *apiLogger) DPanicf(template string, args ...interface{}) {
	l.sugaredLogger.DPanicf(template, args...)
}

func (l *apiLogger) Debug(args ...interface{}) {
	l.sugaredLogger.Debug(args...)
}

func (l *apiLogger) Debugf(template string, args ...interface{}) {
	l.sugaredLogger.Debugf(template, args...)
}

func (l *apiLogger) Info(args ...interface{}) {
	l.sugaredLogger.Info(args...)
}

func (l *apiLogger) Infof(template string, args ...interface{}) {
	l.sugaredLogger.Infof(template, args...)
}

func (l *apiLogger) Error(args ...interface{}) {
	l.sugaredLogger.Error(args...)
}

func (l *apiLogger) Errorf(template string, args ...interface{}) {
	l.sugaredLogger.Errorf(template, args...)
}

func (l *apiLogger) Warn(args ...interface{}) {
	l.sugaredLogger.Warn(args...)
}

func (l *apiLogger) Warnf(template string, args ...interface{}) {
	l.sugaredLogger.Warnf(template, args...)
}
