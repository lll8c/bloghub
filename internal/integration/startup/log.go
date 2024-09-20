package startup

import "geektime/webook/pkg/logger"

func InitLogger() logger.LoggerV1 {
	return logger.NewNopLogger()
}
