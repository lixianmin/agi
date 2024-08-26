package deepseek

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"time"

	"github.com/lixianmin/agi/chat"
	"github.com/lixianmin/agi/ifs"
	"github.com/lixianmin/got/convert"
)

/*
*******************************************************************
created:    2024-08-23
author:     lixianmin

Copyright (C) - All Rights Reserved
********************************************************************
*/

type (
	DeepSeekClient struct {
		client        *http.Client
		authorization string
	}

	ChatRequest struct {
		chat.Request

		FrequencyPenalty float32  `json:"frequency_penalty,omitempty"`
		MaxTokens        int32    `json:"max_tokens,omitempty"`
		Stop             []string `json:"stop,omitempty"`
		Temperature      float32  `json:"temperature,omitempty"`
		TopP             float32  `json:"top_p,omitempty"`

		ResponseFormat string `json:"response_format,omitempty"`
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

// NewDeepSeekClient 线程安全+无状态
func NewDeepSeekClient(secretKey string) *DeepSeekClient {
	var client = &http.Client{}
	return &DeepSeekClient{
		client:        client,
		authorization: "Bearer " + secretKey,
	}
}

func (my *DeepSeekClient) Chat(ctx context.Context, request *ChatRequest) (*ChatCompletionChunk, error) {
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

func (my *DeepSeekClient) sendChatRequest(ctx context.Context, request *ChatRequest) (*http.Response, error) {
	const requestUrl = "https://api.deepseek.com/chat/completions"
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
