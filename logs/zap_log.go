package logs

import (
	"go.uber.org/zap"
)

var Logger *zap.Logger

func init() {
	config := zap.NewDevelopmentConfig()
	config.OutputPaths = []string{"stderr", "logs/log.txt"}
	build, _ := config.Build()
	zap.ReplaceGlobals(build)
}
