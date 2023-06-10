/* This part of the code is reserved for functionality */
package repository

import (
	"headscale-panel/common"
	"headscale-panel/model"
	"headscale-panel/vo"
)

type IMessageRepository interface {
	ListMessages(req *vo.ListMessages) ([]*model.Message, int64, error)
	CreateMessage(msg *model.Message) error
	DeleteMessage(id []uint) error
	UpdateMessage(id []uint, haveRead bool) error
}

type messageRepository struct {
	pip chan *model.Message
}

func NewMessageRepository() IMessageRepository {
	return &messageRepository{pip: make(chan *model.Message)}
}

func (m *messageRepository) ListMessages(req *vo.ListMessages) ([]*model.Message, int64, error) {
	data := make([]*model.Message, 10)
	db := common.DB.Model(data).Order("created_at DESC")
	if req.Type != 0 {
		db = db.Where("type = ?", req.Type)
	}

	var total int64
	err := db.Count(&total).Error
	if err != nil {
		return nil, total, err
	}

	if req.PageNum > 0 && req.PageSize > 0 {
		err = db.Offset(int((req.PageNum - 1) * req.PageSize)).Limit(int(req.PageSize)).Find(&data).Error
	} else {
		err = db.Find(&data).Error
	}

	return data, total, err
}

func (m *messageRepository) CreateMessage(msg *model.Message) error {
	if err := common.DB.Create(msg).Error; err != nil {
		return err
	}
	return nil
}

func (m *messageRepository) DeleteMessage(id []uint) error {
	msg := &model.Message{}
	return common.DB.Model(msg).Where("id in (?)", id).Delete(msg).Error
}

func (m *messageRepository) UpdateMessage(id []uint, haveRead bool) error {
	return common.DB.Model(&model.Message{}).Where("id in (?)", id).Update("have_read", haveRead).Error
}
