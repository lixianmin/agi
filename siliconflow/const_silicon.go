package siliconflow

import "errors"

/********************************************************************
created:    2024-08-22
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

const (
	maxBufferSize = 512 * 1024
)

var ErrRequestIsNil = errors.New("request is nil")
