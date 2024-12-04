package main

import (
	"context"
	"iptv/cmd/iptv/cmds"
	"iptv/internal/pkg/logging"
	"iptv/internal/pkg/util"
	"path"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func init() {
	currPath, err := util.GetCurrentAbPathByExecutable()
	if err != nil {
		panic(err)
	}
	logFile := path.Join(currPath, "iptv.log")

	// 初始化日志
	err = logging.InitLogger(&logging.LogConfig{
		Level:      zapcore.InfoLevel,
		FileName:   logFile,
		MaxSize:    100,
		MaxBackups: 10,
		IsStdout:   true,
	})
	if err != nil {
		panic(err)
	}
}

func main() {
	// L()：获取全局logger
	logger := zap.L()
	defer logger.Sync()

	cobra.CheckErr(cmds.NewRootCLI().ExecuteContext(context.Background()))
}
