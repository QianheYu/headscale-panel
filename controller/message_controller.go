/* This part of the code is reserved for functionality */
package controller

import (
	"github.com/gin-gonic/gin"
	"headscale-panel/repository"
	"headscale-panel/response"
	"headscale-panel/vo"
)

type IMessageController interface {
	ListMessages(c *gin.Context)
	DeleteMessage(c *gin.Context)
	HaveReadMessage(c *gin.Context)
}

type MessageController struct {
	repo repository.IMessageRepository
}

func NewMessageController() IMessageController {
	return &MessageController{repo: repository.NewMessageRepository()}
}

func (m *MessageController) ListMessages(c *gin.Context) {
	req := &vo.ListMessages{}
	if err := c.ShouldBindJSON(req); err != nil {
		response.Fail(c, nil, "")
		return
	}

	list, total, err := m.repo.ListMessages(req)
	if err != nil {
		response.Fail(c, nil, err.Error())
		return
	}
	response.Success(c, map[string]interface{}{"list": list, "total": total}, "")
}

func (m *MessageController) DeleteMessage(c *gin.Context) {
	req := &vo.DeleteMessage{}
	if err := c.ShouldBindJSON(req); err != nil {
		response.Fail(c, nil, "")
		return
	}

	if err := m.repo.DeleteMessage(req.Ids); err != nil {
		response.Fail(c, nil, err.Error())
		return
	}
	response.Success(c, nil, "")
}

func (m *MessageController) HaveReadMessage(c *gin.Context) {
	req := &vo.HaveReadMessage{}
	if err := c.ShouldBindJSON(req); err != nil {
		response.Fail(c, nil, "")
		return
	}

	if err := m.repo.UpdateMessage(req.Ids, req.HaveRead); err != nil {
		response.Fail(c, nil, err.Error())
		return
	}
	response.Success(c, nil, "")
}
