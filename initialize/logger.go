package initialize

import "go.uber.org/zap"

func InitLogger() {
	config := zap.NewDevelopmentConfig()
	config.OutputPaths = []string{"stdout"}
	build, _ := config.Build()
	zap.ReplaceGlobals(build)
}
