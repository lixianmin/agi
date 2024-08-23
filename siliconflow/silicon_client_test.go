package siliconflow

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
created:    2020-09-22
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

func getSecretKey() string {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	var sk = os.Getenv("SLICONFLOW_SECRET_KEY")
	return sk
}

func TestChat(t *testing.T) {
	var sk = getSecretKey()
	var client = NewSiliconClient(sk)

	const modelName = "Qwen/Qwen2-7B-Instruct"
	var chatThread = chat.NewThread()
	chatThread.SetPrompt("你是一个脑残人士, 无论我说什么, 你都回答`是的`. 现在开始: ")
	chatThread.AddUserMessage("今天天气怎么样?")
	chatThread.AddBotMessage("是的")
	chatThread.AddUserMessage("你觉得我帅嘛?")

	var request = &ChatRequest{
		Model:       modelName,
		Messages:    chatThread.CloneMessages(),
		Temperature: chatThread.GetTemperature(),
		TopK:        chatThread.GetTopK(),
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

func TestTranscribeAudio(t *testing.T) {
	var sk = getSecretKey()
	var client = NewSiliconClient(sk)

	const modelName = "iic/SenseVoiceSmall"

	var ctx, cancel = context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	var bts, _ = os.ReadFile("C:\\Users\\user\\Downloads\\temp\\record_out.wav")
	var result, err = client.TranscribeAudio(ctx, modelName, bts)
	if err != nil {
		log.Fatalf("transcribe error: %v", err)
		return
	}

	println(result)
}
