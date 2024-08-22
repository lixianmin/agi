package chat

/********************************************************************
created:    2024-07-07
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

type threadOptions struct {
	prompt      string
	userRole    string
	botRole     string
	temperature float32
	topK        int32
	topP        float32

	historySize int
}

type ThreadOption func(*threadOptions)

func WithPrompt(prompt string) ThreadOption {
	return func(options *threadOptions) {
		if options.prompt != "" {
			options.prompt = prompt
		}
	}
}

// 这个设置好像没有意义, 因此第一次请求的时候server也不会知道bot的role是什么, 然后就返回了assistant. 使用siliconflow的测试是这样的结果
// func WithUserRole(userRole string) ThreadOption {
// 	return func(options *threadOptions) {
// 		if options.userRole != "" {
// 			options.userRole = userRole
// 		}
// 	}
// }

// func WithBotRole(botRole string) ThreadOption {
// 	return func(options *threadOptions) {
// 		if options.botRole != "" {
// 			options.botRole = botRole
// 		}
// 	}
// }

func WithHistorySize(size int) ThreadOption {
	return func(options *threadOptions) {
		if size > 0 {
			options.historySize = size / 2 * 2
		}
	}
}

func WithTemperature(temperature float32) ThreadOption {
	return func(options *threadOptions) {
		if temperature > 0 {
			options.temperature = min(temperature, 1)
		}
	}
}

func WithTopK(v int32) ThreadOption {
	return func(options *threadOptions) {
		if v > 0 {
			options.topK = v
		}
	}
}

func WithTopP(v float32) ThreadOption {
	return func(options *threadOptions) {
		if v > 0 {
			options.topP = min(v, 1)
		}
		//options.topP = mathx.Clamp(v, 0, 1)
	}
}
