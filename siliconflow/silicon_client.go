package siliconflow

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"regexp"
	"time"

	"github.com/lixianmin/agi/chat"
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
		Model       string          `json:"model"`
		Messages    []*chat.Message `json:"messages"`
		Stream      bool            `json:"stream,omitempty"`
		Temperature float32         `json:"temperature,omitempty"`
		TopK        int32           `json:"top_k,omitempty"`
		TopP        float32         `json:"top_p,omitempty"`
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
		return nil, ErrRequestIsNil
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
		return ErrRequestIsNil
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

func (my *SiliconClient) TranscribeAudio(ctx context.Context, modelName string, filePath string) (string, error) {
	const requestUrl = "https://api.siliconflow.cn/v1/audio/transcriptions"

	var fin, err1 = os.Open(filePath)
	if err1 != nil {
		return "", err1
	}
	defer fin.Close()

	var requestBody bytes.Buffer
	var writer = multipart.NewWriter(&requestBody)

	// Add form fields
	_ = writer.WriteField("model", modelName)

	// filename是什么不重要，重要的是扩展名
	part, err2 := writer.CreateFormFile("file", "abc.mp3")
	if err2 != nil {
		return "", err2
	}

	var _, err3 = io.Copy(part, fin)
	if err3 != nil {
		return "", err3
	}

	_ = writer.Close()

	req, err4 := http.NewRequestWithContext(ctx, "POST", requestUrl, &requestBody)
	if err4 != nil {
		return "", err4
	}

	// Set headers
	req.Header.Set("accept", "application/json")
	req.Header.Set("authorization", my.authorization)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err5 := my.client.Do(req)
	if err5 != nil {
		return "", err5
	}
	defer resp.Body.Close()

	body, err6 := io.ReadAll(resp.Body)
	if err6 != nil {
		return "", err6
	}

	type Output struct {
		Text string `json:"text"`
	}

	var output Output
	convert.FromJson(body, &output)
	return output.Text, nil
}
