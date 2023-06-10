/* This part of the code is reserved for functionality */
package controller

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"headscale-panel/log"
	"time"
)

type Template struct {
	Content string `json:"content"`
}

type INoticeController interface {
	Controller(c *gin.Context)
}

type NoticeController struct {
	notice chan string
}

func NewNoticeController() INoticeController {
	return &NoticeController{notice: make(chan string)}
}

func (m *NoticeController) Controller(c *gin.Context) {
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")

	go func() {
		defer func() {
			if err := recover(); err != nil {
				log.Log.Error(err)
			}
		}()
		c.SSEvent("message", <-m.notice)
	}()

	for i := 0; i < 10; i++ {
		data, err := json.Marshal(Template{Content: "test"})
		if err != nil {
			fmt.Errorf("%s", err.Error())
			continue
		}
		m.notice <- fmt.Sprintf("%s", string(data))
		time.Sleep(12 * time.Second)
	}
}
