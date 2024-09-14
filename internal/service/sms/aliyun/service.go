package aliyun

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/dysmsapi"
	"go.uber.org/zap"
	"strings"
)

type Service struct {
	signName string
	client   *dysmsapi.Client
}

func NewService(signName string, client *dysmsapi.Client) *Service {
	return &Service{
		signName: signName,
		client:   client,
	}
}

func (s Service) Send(ctx context.Context, tplId string, args []string, numbers ...string) error {
	//创建请求
	req := dysmsapi.CreateSendSmsRequest()
	req.SignName = s.signName
	req.TemplateCode = tplId
	//请求协议
	req.Scheme = "https"
	//接收短信的手机号码
	req.PhoneNumbers = strings.Join(numbers, ",")

	//传入的是map[string]string类型
	argsMap := make(map[string]interface{})
	for _, arg := range args {
		argsMap["code"] = arg
	}
	bCode, err := json.Marshal(argsMap)
	if err != nil {
		return err
	}
	req.TemplateParam = string(bCode)
	var resp *dysmsapi.SendSmsResponse
	resp, err = s.client.SendSms(req)
	zap.L().Debug("发送短信",
		zap.Any("req", req),
		zap.Any("resp", resp),
		zap.Error(err))
	if err != nil {
		return err
	}
	if resp.Code != "OK" {
		return fmt.Errorf("发送失败, code:%s, 原因:%s", resp.Code,
			resp.Message)
	}
	return err
}
