package chat

/********************************************************************
created:    2024-08-22
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

type (
	Message struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	}

	ResponseFormat struct {
		Type string `json:"type,omitempty"` // 支持： json_object
	}

	Request struct {
		Model          string          `json:"model"`
		Messages       []*Message      `json:"messages"`
		Stream         bool            `json:"stream,omitempty"`
		ResponseFormat *ResponseFormat `json:"response_format,omitempty"`
	}
)
