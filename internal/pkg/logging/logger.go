package logging

import (
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

type LogConfig struct {
	Level        zapcore.Level `json:"level"`          // Level 最低日志等级，DEBUG<INFO<WARN<ERROR<FATAL 例如：info-->收集info等级以上的日志
	FileName     string        `json:"file_name"`      // FileName 日志文件位置
	MaxSize      int           `json:"max_size"`       // MaxSize 进行切割之前，日志文件的最大大小(MB为单位)，默认为100MB
	MaxAge       int           `json:"max_age"`        // MaxAge 是根据文件名中编码的时间戳保留旧日志文件的最大天数。
	MaxBackups   int           `json:"max_backups"`    // MaxBackups 是要保留的旧日志文件的最大数量。默认是保留所有旧的日志文件（尽管 MaxAge 可能仍会导致它们被删除。）
	IsStdout     bool          `json:"is_stdout"`      // IsStdout 是否输出到控制台
	IsStackTrace bool          `json:"is_stack_trace"` // IsStackTrace 是否输出堆栈信息
}

// InitLogger 初始化Logger
func InitLogger(lCfg *LogConfig) (err error) {
	writeSyncer := getLogWriter(lCfg.FileName, lCfg.MaxSize, lCfg.MaxBackups, lCfg.MaxAge, lCfg.IsStdout)
	encoder := getEncoder()

	core := zapcore.NewCore(encoder, writeSyncer, lCfg.Level)
	var logger *zap.Logger
	if lCfg.IsStackTrace {
		logger = zap.New(core, zap.AddCaller(), zap.AddStacktrace(zap.ErrorLevel))
	} else {
		logger = zap.New(core, zap.AddCaller())
	}
	zap.ReplaceGlobals(logger)
	return
}

// getEncoder 负责设置 encoding 的日志格式
func getEncoder() zapcore.Encoder {
	encodeConfig := zap.NewProductionEncoderConfig()
	encodeConfig.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(t.Format("2006-01-02 15:04:05.000"))
	}
	encodeConfig.TimeKey = "time"
	encodeConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	encodeConfig.EncodeCaller = zapcore.ShortCallerEncoder
	return zapcore.NewJSONEncoder(encodeConfig)
}

// getLogWriter 负责日志写入的位置
func getLogWriter(filename string, maxsize, maxBackup, maxAge int, isStdout bool) zapcore.WriteSyncer {
	lumberJackLogger := &lumberjack.Logger{
		Filename:   filename,  // 文件位置
		MaxSize:    maxsize,   // 进行切割之前,日志文件的最大大小(MB为单位)
		MaxAge:     maxAge,    // 保留旧文件的最大天数
		MaxBackups: maxBackup, // 保留旧文件的最大个数
		Compress:   true,      // 是否压缩/归档旧文件
	}
	if isStdout {
		return zapcore.NewMultiWriteSyncer(zapcore.AddSync(lumberJackLogger), zapcore.AddSync(os.Stdout))
	} else {
		return zapcore.AddSync(lumberJackLogger)
	}
}
