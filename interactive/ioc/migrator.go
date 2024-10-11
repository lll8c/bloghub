package ioc

import (
	"geektime/webook/interactive/repository/dao"
	"geektime/webook/pkg/ginx"
	"geektime/webook/pkg/gormx/connpool"
	"geektime/webook/pkg/logger"
	"geektime/webook/pkg/migrator/events"
	"geektime/webook/pkg/migrator/events/fixer"
	"geektime/webook/pkg/migrator/scheduler"
	"github.com/IBM/sarama"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

// InitMigratorWebServer 管理后台的 server
func InitMigratorWebServer(l logger.LoggerV1,
	src SrcDB,
	dst DstDB,
	pool *connpool.DoubleWritePool,
	producer events.Producer) *ginx.Server {

	engine := gin.Default()
	group := engine.Group("/migrator")
	sch := scheduler.NewScheduler[dao.Interactive](l, src, dst, pool, producer)
	sch.RegisterRoutes(group)
	return &ginx.Server{
		Engine: engine,
		Addr:   viper.GetString("migrator.http.addr"),
	}
}

func InitInteractiveProducer(p sarama.SyncProducer) events.Producer {
	return events.NewSaramaProducer("inconsistent_interactive", p)
}

func InitFixerConsumer(client sarama.Client,
	l logger.LoggerV1,
	src SrcDB,
	dst DstDB) *fixer.Consumer[dao.Interactive] {
	res, err := fixer.NewConsumer[dao.Interactive](client, l, "inconsistent_interactive", src, dst)
	if err != nil {
		panic(err)
	}
	return res
}
