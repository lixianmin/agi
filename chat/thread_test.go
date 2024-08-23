package chat

import (
	"encoding/json"
	"strconv"
	"testing"
)

/********************************************************************
created:    2020-09-22
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

func TestThread(t *testing.T) {
	var thread = NewThread()
	for i := 0; i < 3; i++ {
		thread.AddUserMessage("user " + strconv.Itoa(i))
		thread.AddBotMessage("bot " + strconv.Itoa(i))
	}

	var messages = thread.CloneMessages()
	var text, _ = json.Marshal(messages)
	println(string(text))
}
