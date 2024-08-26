package deepseek

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"testing"
	"time"

	"github.com/joho/godotenv"
	"github.com/lixianmin/agi/chat"
)

/********************************************************************
created:    2024-08-24
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

func getSecretKey() string {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	var sk = os.Getenv("DEEPSEEK_SECRET_KEY")
	return sk
}

func TestChat(t *testing.T) {
	var sk = getSecretKey()
	var client = NewDeepSeekClient(sk)

	const modelName = "deepseek-chat"
	var chatThread = chat.NewThread()
	chatThread.SetPrompt("你是一个脑残人士, 无论我说什么, 你都回答`是的`. 现在开始: ")
	chatThread.AddUserMessage("今天天气怎么样?")
	chatThread.AddBotMessage("是的")
	chatThread.AddUserMessage("你觉得我帅嘛?")

	var request = &ChatRequest{
		Request: chat.Request{
			Model:    modelName,
			Messages: chatThread.CloneMessages(),
		},
		Temperature: chatThread.GetTemperature(),
		TopP:        chatThread.GetTopP(),
	}

	var req, _ = json.Marshal(request)
	println(string(req))

	var ctx, cancel = context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	var response, err = client.Chat(ctx, request)
	if err != nil {
		log.Fatalf("chat error: %v", err)
		return
	}

	var result, _ = json.Marshal(response)
	println(string(result))
}
