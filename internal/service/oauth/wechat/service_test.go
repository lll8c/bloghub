package wechat

import (
	"context"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

func Test_service_VerifyCode(t *testing.T) {
	appID, ok := os.LookupEnv("WECHAT_APP_ID")
	if !ok {
		panic("找不到环境变量 WECHAT_APP_ID")
	}
	appSecret, ok := os.LookupEnv("WECHAT_APP_SECRET")
	if !ok {
		panic("找不到环境变量 WECHAT_APP_SECRET")
	}
	svc := NewService(appID, appSecret)
	res, err := svc.VerifyCode(context.Background(), "", "state")
	require.NoError(t, err)
	t.Log(res)
}
