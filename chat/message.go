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
)
