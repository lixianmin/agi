package chat

import "sync"

/********************************************************************
created:    2024-06-01
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

type (
	Thread struct {
		userRole    string
		botRole     string
		temperature float32
		topP        float32
		topK        int32

		messages []Message
		cloned   []Message
		m        sync.Mutex
	}
)

func NewThread(opts ...ThreadOption) *Thread {
	// 默认值
	var options = threadOptions{
		prompt:      "You are an English expert, you can help me to improve my English skills. The following are chats between you and me.",
		userRole:    "user",
		botRole:     "assistant",
		temperature: 0.7,
		topK:        50,
		topP:        0.7,

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
		messages:    make([]Message, 1, options.historySize+1), // index=0 is system prompt
		cloned:      make([]Message, 1, options.historySize+1),
	}

	thread.messages[0] = Message{
		Role:    "system",
		Content: options.prompt,
	}

	return thread
}

func (my *Thread) SetPrompt(prompt string) {
	if prompt != "" {
		my.m.Lock()
		my.messages[0].Content = prompt
		my.m.Unlock()
	}
}

func (my *Thread) AddUserMessage(content string) {
	if content != "" {
		var message = Message{Role: my.userRole, Content: content}
		my.addMessage(message)
	}
}

func (my *Thread) AddBotMessage(content string) {
	if content != "" {
		var message = Message{Role: my.botRole, Content: content}
		my.addMessage(message)
	}
}

func (my *Thread) addMessage(message Message) {
	my.m.Lock()
	{
		var count = len(my.messages)
		if count == cap(my.messages) {
			copy(my.messages[1:], my.messages[2:])
			my.messages[count-1] = message
		} else {
			my.messages = append(my.messages, message)
		}
	}
	my.m.Unlock()
}

func (my *Thread) GetMessages() []Message {
	my.m.Lock()
	{
		var deltaSize = len(my.messages) - len(my.cloned)
		for i := 0; i < deltaSize; i++ {
			my.cloned = append(my.cloned, Message{})
		}

		copy(my.cloned, my.messages)
	}
	my.m.Unlock()

	return my.cloned
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
