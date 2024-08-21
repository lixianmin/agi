package siliconflow

// import (
// 	"context"
// 	"net/http"
// 	"strings"

// 	"github.com/lixianmin/fiction-spider/app/ai"
// 	"github.com/ollama/ollama/api"
// )

// /********************************************************************
// created:    2024-06-30
// author:     lixianmin

// Copyright (C) - All Rights Reserved
// *********************************************************************/

// type (
// 	SiliconService struct {
// 		chatUrl string
// 		client  *SiliconClient
// 	}
// )

// func NewSiliconService(secretKey string) *SiliconService {

// 	var client = NewSiliconClient(http.DefaultClient, secretKey)
// 	var service = &SiliconService{
// 		client: client,
// 	}

// 	return service
// }

// func (my *SiliconService) StreamChat(ctx context.Context, modelName string, thread *ai.Thread, fn func(line string)) (string, error) {
// 	if modelName == "" || thread == nil {
// 		return "", nil
// 	}

// 	var request = &ChatRequest{
// 		Model:       modelName,
// 		Messages:    thread.GetMessages(),
// 		Stream:      true,
// 		Temperature: thread.GetTemperature(),
// 		TopK:        thread.GetTopK(),
// 		TopP:        thread.GetTopP(),
// 	}

// 	var sb strings.Builder
// 	var answer string

// 	// the callback of client.Chat is a synchronized function
// 	var err = my.client.StreamChat(ctx, request, func(response api.ChatResponse) error {
// 		var message = response.Message
// 		var content = message.Content
// 		sb.WriteString(content)

// 		if response.Done {
// 			answer = sb.String()

// 			//var newMessage = api.Message{Role: message.Role, Content: answer}
// 			thread.AddAnswer(answer)
// 		}

// 		if fn != nil {
// 			fn(content)
// 		}

// 		return nil
// 	})

// 	return answer, err
// }

// func (my *SiliconService) TranscribeAudio(ctx context.Context, modelName string, filePath string) (string, error) {
// 	return my.client.TranscribeAudio(ctx, modelName, filePath)
// }
