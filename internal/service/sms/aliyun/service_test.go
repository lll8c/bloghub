package aliyun

/*
// 这个需要手动跑，也就是你需要在本地搞好这些环境变量
func TestSender(t *testing.T) {

	client, err := dysmsapi.NewClientWithAccessKey("cn-hunan",
		"123",
		"")
	if err != nil {
		t.Fatal(err)
	}

	s := NewService("小微书", client)

	tplId := "SMS_472665076"
	params := []string{"hhh"}
	numbers := []string{"18873288626"}

	err = s.Send(context.Background(), tplId, params, numbers...)
	if err != nil {
		fmt.Println(err)
	}
}*/
