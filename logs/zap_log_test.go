package logs

import (
	"go.uber.org/zap"
	"testing"
)

func TestZapLog(t *testing.T) {
	config := zap.NewDevelopmentConfig()
	config.OutputPaths = []string{"stderr", "log.txt"}
	build, _ := config.Build()
	Logger = build
	Logger.Info("helloworld", zap.String("user", "user"))
}
