package logs

import (
	"go.uber.org/zap"
	"rollcall_xmu/initialize"
	"testing"
)

func TestZapLog(t *testing.T) {
	initialize.InitLogger()
	zap.L().Info("hellow world")
}

type Node struct {
	val  int
	next *Node
}

func TestTmp(t *testing.T) {
	Logger.Info(
		"test",
		zap.String("key", "value"),
		zap.Int("num", 1234),
	)
}
