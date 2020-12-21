package log

import (
	"strings"

	"github.com/kelseyhightower/envconfig"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	Logger *zap.Logger

	statLogLevelCount = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "log_level_total",
			Help: "Number of log statements, differentiated by log level.",
		},
		[]string{"level"})
)

type Config struct {
	LogLevel string `split_words:"true" default:"info"`
}

func Level(cfgLevel string) zap.AtomicLevel {
	var atomicLevel zap.AtomicLevel
	switch cfgLevel {
	case "debug":
		atomicLevel = zap.NewAtomicLevelAt(zap.DebugLevel)
	case "info":
		atomicLevel = zap.NewAtomicLevelAt(zap.InfoLevel)
	case "error", "err":
		atomicLevel = zap.NewAtomicLevelAt(zap.ErrorLevel)
	case "fatal":
		atomicLevel = zap.NewAtomicLevelAt(zap.FatalLevel)
	case "panic":
		atomicLevel = zap.NewAtomicLevelAt(zap.PanicLevel)
	default:
		atomicLevel = zap.NewAtomicLevelAt(zap.InfoLevel)
	}
	return atomicLevel
}

func newLogger(ll string) *zap.Logger {
	cfg := zap.Config{
		Level:            Level(strings.ToLower(ll)),
		Encoding:         "json",
		DisableCaller:    true,
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
		EncoderConfig: zapcore.EncoderConfig{
			MessageKey: "message",

			LevelKey:    "level",
			EncodeLevel: zapcore.CapitalLevelEncoder,

			TimeKey:    "time",
			EncodeTime: zapcore.RFC3339TimeEncoder,

			CallerKey:    "caller",
			EncodeCaller: zapcore.ShortCallerEncoder,
		},
	}

	logger, err := cfg.Build(zap.WrapCore(func(core zapcore.Core) zapcore.Core {
		return zapcore.RegisterHooks(core, func(e zapcore.Entry) error {
			statLogLevelCount.WithLabelValues(e.Level.String()).Inc()
			return nil
		})
	}))

	if err != nil {
		panic(err)
	}

	return logger
}

func init() {
	var c Config
	configErr := envconfig.Process("", &c)
	Logger = newLogger(c.LogLevel)
	if configErr != nil {
		Logger.Error("Logger Config failed to Load", zap.Error(configErr))
	}
	Logger.Info("Logger Initialized", zap.String("level", c.LogLevel))
}
