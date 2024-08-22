package siliconflow

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"time"

	"github.com/lixianmin/agi/chat"
	"github.com/lixianmin/got/convert"
)

/********************************************************************
created:    2024-06-30
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

type SiliconClient struct {
	baseUrl       *url.URL
	client        *http.Client
	authorization string
}

type ChatRequest struct {
	Model       string         `json:"model"`
	Messages    []chat.Message `json:"messages"`
	Stream      bool           `json:"stream,omitempty"`
	Temperature float32        `json:"temperature,omitempty"`
	TopK        int32          `json:"top_k,omitempty"`
	TopP        float32        `json:"top_p,omitempty"`
}

type ChatResponse struct {
	Model      string       `json:"model"`
	CreatedAt  time.Time    `json:"created_at"`
	Message    chat.Message `json:"message"`
	DoneReason string       `json:"done_reason,omitempty"`

	Done bool `json:"done"`
}

type ChatCompletionChunk struct {
	ID                string          `json:"id"`
	Object            string          `json:"object"`
	Created           int64           `json:"created"`
	Model             string          `json:"model"`
	SystemFingerprint string          `json:"system_fingerprint"`
	Choices           []ChunkedChoice `json:"choices"`
}

type ChunkedChoice struct {
	Index        int          `json:"index"`
	Delta        chat.Message `json:"delta"`
	FinishReason string       `json:"finish_reason"`
}

type ChatResponseFunc func(ChatResponse) error

const maxBufferSize = 512 * 1024

func NewSiliconClient(secretKey string) *SiliconClient {
	const base = "https://api.siliconflow.cn"
	var baseUrl, err = url.Parse(base)
	if err != nil {
		panic(err)
	}

	var client = http.DefaultClient
	return &SiliconClient{
		baseUrl:       baseUrl,
		client:        client,
		authorization: "Bearer " + secretKey,
	}
}

func (my *SiliconClient) StreamChat(ctx context.Context, request *ChatRequest, fn ChatResponseFunc) error {
	if fn == nil {
		return errors.New("fn is nil")
	}

	var data = request
	const path = "/v1/chat/completions"

	var buf *bytes.Buffer
	if data != nil {
		bts, err := convert.ToJsonE(data)
		if err != nil {
			return err
		}

		buf = bytes.NewBuffer(bts)
	}

	var requestURL = my.baseUrl.JoinPath(path)
	var request1, err1 = http.NewRequestWithContext(ctx, http.MethodPost, requestURL.String(), buf)
	if err1 != nil {
		return err1
	}

	request1.Header.Set("Content-Type", "application/json")
	request1.Header.Set("authorization", my.authorization)

	var response, err2 = my.client.Do(request1)
	if err2 != nil {
		return err2
	}
	defer response.Body.Close()

	var scanner = bufio.NewScanner(response.Body)
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
		var err4 = json.Unmarshal(convert.Bytes(cleanLine), &chunk)
		if err4 != nil {
			return err4
		}

		if len(chunk.Choices) > 0 {
			var choice = chunk.Choices[0]
			var chatResponse = ChatResponse{
				Model:     chunk.Model,
				CreatedAt: time.Unix(chunk.Created, 0),
				Message: chat.Message{
					Role:    choice.Delta.Role,
					Content: choice.Delta.Content,
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

func (my *SiliconClient) TranscribeAudio(ctx context.Context, modelName string, filePath string) (string, error) {
	const path = "/v1/audio/transcriptions"
	var requestUrl = my.baseUrl.JoinPath(path)

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

	req, err4 := http.NewRequestWithContext(ctx, "POST", requestUrl.String(), &requestBody)
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
