package main

import "github.com/labstack/echo/v4"

// Text 向前端返回一个简单的文本消息。
// 为了保持统一性，总是向前端返回 JSON, 因此即使是简单的文本消息也使用 JSON.
type Text struct {
	Text string `json:"text"`
}

func waitingFolder(c echo.Context) error {

}
