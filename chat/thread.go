package chat

import (
	"sync"
)

/********************************************************************
created:    2024-06-01
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

type (
	Thread struct {
		systemPromptMessage *Message
		userRole            string
		botRole             string
		temperature         float32
		topP                float32
		topK                int32

		messages []Message
		m        sync.Mutex
	}
)

func NewThread(opts ...ThreadOption) *Thread {
	// 默认值
	var options = threadOptions{
		systemPrompt: "",
		userRole:     "user",
		botRole:      "assistant",
		temperature:  0.7,
		topK:         50,
		topP:         0.7,

		historySize: 20,
	}

	// 初始化
	for _, opt := range opts {
		opt(&options)
	}

	var thread = &Thread{
		userRole:    options.userRole,
		botRole:     options.botRole,
		temperature: options.temperature,
		topP:        options.topP,
		topK:        options.topK,
		messages:    make([]Message, 0, options.historySize),
	}

	if options.systemPrompt != "" {
		thread.systemPromptMessage = &Message{
			Role:    "system",
			Content: options.systemPrompt,
		}
	}

	return thread
}

func (my *Thread) AddAnswer(answer string) {
	var message = Message{Role: my.botRole, Content: answer}
	my.addMessage(message)
}

func (my *Thread) addMessage(message Message) {
	var count = len(my.messages)
	if count == cap(my.messages) {
		for i := 0; i < count-1; i++ {
			my.messages[i] = my.messages[i+1]
		}

		my.messages[count-1] = message
	} else {
		my.messages = append(my.messages, message)
	}
}

func (my *Thread) GetMessages() []Message {
	if my.systemPromptMessage != nil {
		return append([]Message{*my.systemPromptMessage}, my.messages...)
	}

	return my.messages
}

func (my *Thread) GetTemperature() float32 {
	return my.temperature
}

func (my *Thread) GetTopP() float32 {
	return my.topP
}

func (my *Thread) GetTopK() int32 {
	return my.topK
}
