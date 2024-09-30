package siliconflow

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"regexp"
	"time"

	"github.com/lixianmin/agi/chat"
	"github.com/lixianmin/agi/ifs"
	"github.com/lixianmin/got/convert"
)

/*
*******************************************************************
created:    2024-06-30
author:     lixianmin

Copyright (C) - All Rights Reserved
********************************************************************
*/
type (
	SiliconClient struct {
		client        *http.Client
		authorization string
	}

	ChatRequest struct {
		chat.Request

		FrequencyPenalty float32  `json:"frequency_penalty,omitempty"`
		MaxTokens        int32    `json:"max_tokens,omitempty"`
		Stop             []string `json:"stop,omitempty"`
		Temperature      float32  `json:"temperature,omitempty"`
		TopK             int32    `json:"top_k,omitempty"`
		TopP             float32  `json:"top_p,omitempty"`
	}

	ChatResponse struct {
		Model      string       `json:"model"`
		CreatedAt  time.Time    `json:"created_at"`
		Message    chat.Message `json:"message"`
		DoneReason string       `json:"done_reason,omitempty"`

		Done bool `json:"done"`
	}

	ChatCompletionChunk struct {
		ID                string          `json:"id"`
		Object            string          `json:"object"`
		Created           int64           `json:"created"`
		Model             string          `json:"model"`
		SystemFingerprint string          `json:"system_fingerprint"`
		Choices           []ChunkedChoice `json:"choices"`
	}

	ChunkedChoice struct {
		Index        int          `json:"index"`
		Message      chat.Message `json:"message"`
		FinishReason string       `json:"finish_reason"`
	}

	ChatResponseFunc func(ChatResponse) error
)

// NewSiliconClient 线程安全+无状态
func NewSiliconClient(secretKey string) *SiliconClient {
	var client = &http.Client{}
	return &SiliconClient{
		client:        client,
		authorization: "Bearer " + secretKey,
	}
}

func (my *SiliconClient) Chat(ctx context.Context, request *ChatRequest) (*ChatCompletionChunk, error) {
	if request == nil {
		return nil, ifs.ErrRequestIsNil
	}

	request.Stream = false
	var response1, err1 = my.sendChatRequest(ctx, request)
	if err1 != nil {
		return nil, err1
	}
	defer response1.Body.Close()

	var bts, err2 = io.ReadAll(response1.Body)
	if err2 != nil {
		return nil, err2
	}

	var response ChatCompletionChunk
	if err3 := convert.FromJsonE(bts, &response); err3 != nil {
		return nil, err3
	}

	return &response, nil
}

func (my *SiliconClient) StreamChat(ctx context.Context, request *ChatRequest, fn ChatResponseFunc) error {
	if request == nil {
		return ifs.ErrRequestIsNil
	}

	if fn == nil {
		return errors.New("fn is nil")
	}

	request.Stream = true
	var response1, err1 = my.sendChatRequest(ctx, request)
	if err1 != nil {
		return err1
	}
	defer response1.Body.Close()

	var scanner = bufio.NewScanner(response1.Body)
	// increase the buffer size to avoid running out of space
	var scanBuf = make([]byte, 0, maxBufferSize)
	scanner.Buffer(scanBuf, maxBufferSize)

	var regex = regexp.MustCompile(`^data: `)
	for scanner.Scan() {
		var tokens = scanner.Bytes()
		var line = convert.String(tokens)
		if line == "" {
			continue
		}

		if line == "data: [DONE]" {
			var chatResponse = ChatResponse{
				Done: true,
			}

			if err3 := fn(chatResponse); err3 != nil {
				return err3
			}
			return nil
		}

		var cleanLine = regex.ReplaceAllString(line, "")

		var chunk ChatCompletionChunk
		convert.FromJsonS(cleanLine, &chunk)

		if len(chunk.Choices) > 0 {
			var choice = chunk.Choices[0]
			var chatResponse = ChatResponse{
				Model:     chunk.Model,
				CreatedAt: time.Unix(chunk.Created, 0),
				Message: chat.Message{
					Role:    choice.Message.Role,
					Content: choice.Message.Content,
				},
				DoneReason: choice.FinishReason,
				Done:       choice.FinishReason == "stop",
			}

			if err5 := fn(chatResponse); err5 != nil {
				return err5
			}

			if chatResponse.Done {
				return nil
			}
		}
	}

	return nil
}

func (my *SiliconClient) sendChatRequest(ctx context.Context, request *ChatRequest) (*http.Response, error) {
	const requestUrl = "https://api.siliconflow.cn/v1/chat/completions"
	var bts1, err1 = convert.ToJsonE(request)
	if err1 != nil {
		return nil, err1
	}

	var requestBody = bytes.NewBuffer(bts1)
	var request2, err2 = http.NewRequestWithContext(ctx, http.MethodPost, requestUrl, requestBody)
	if err2 != nil {
		return nil, err2
	}

	var header = request2.Header
	header.Set("accept", "application/json")
	header.Set("Content-Type", "application/json")
	header.Set("authorization", my.authorization)

	var response3, err3 = my.client.Do(request2)
	return response3, err3
}

func (my *SiliconClient) TranscribeAudio(ctx context.Context, modelName string, audioData []byte) (string, error) {
	if modelName == "" || len(audioData) == 0 {
		return "", errors.New("invalid parameters")
	}

	const requestUrl = "https://api.siliconflow.cn/v1/audio/transcriptions"

	var requestBody bytes.Buffer
	var writer = multipart.NewWriter(&requestBody)

	// Add form fields
	_ = writer.WriteField("model", modelName)

	// 虽然文件扩展名是.mp3, 但上传.wav也是可以的, 至少iic/SenseVoiceSmall这个模型是可以的
	var part1, err1 = writer.CreateFormFile("file", "audio.mp3")
	if err1 != nil {
		return "", err1
	}

	// Write the byte array to the form file
	var _, err2 = part1.Write(audioData)
	if err2 != nil {
		return "", err2
	}

	_ = writer.Close()

	var request3, err3 = http.NewRequestWithContext(ctx, http.MethodPost, requestUrl, &requestBody)
	if err3 != nil {
		return "", err3
	}

	// Set headers
	var header = request3.Header
	header.Set("accept", "application/json")
	header.Set("authorization", my.authorization)
	header.Set("Content-Type", writer.FormDataContentType())

	var response4, err4 = my.client.Do(request3)
	if err4 != nil {
		return "", err4
	}
	defer response4.Body.Close()

	var body, err5 = io.ReadAll(response4.Body)
	if err5 != nil {
		return "", err5
	}

	type Output struct {
		Text string `json:"text"`
	}

	var output Output
	convert.FromJson(body, &output)
	return output.Text, nil
}
