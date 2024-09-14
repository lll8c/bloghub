package ioc

import (
	"geektime/webook/internal/service/sms"
	"geektime/webook/internal/service/sms/aliyun"
	"geektime/webook/internal/service/sms/localsms"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/dysmsapi"
	"os"
)

func InitSMSService() sms.Service {
	return localsms.NewService()
	// 如果有需要，就可以用这个ailyun的服务
	//return aliyun.NewService()
	//加了限流的阿里云服务
	//return ratelimit.NewLimitSMSService()

	//可以通过套娃增加功能
}

func InitAliyunSMSService() sms.Service {
	secretId, ok := os.LookupEnv("SMS_SECRET_ID")
	if !ok {
		panic("找不到腾讯 SMS 的 secret id")
	}
	secretKey, ok := os.LookupEnv("SMS_SECRET_KEY")
	if !ok {
		panic("找不到腾讯 SMS 的 secret key")
	}
	codeClient, err := dysmsapi.NewClientWithAccessKey("cn-hunan",
		secretId,
		secretKey)
	if err != nil {
		panic(err)
	}
	return aliyun.NewService("小微书", codeClient)
}
