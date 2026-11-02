package repository

import (
	"errors"

	"github.com/dialektov/go-x5/hw6/internal/model"
)

var ErrNotFound = errors.New("уведомление не найдено")

type Memory struct {
	items map[int]model.Notification
}

func NewMemory() *Memory {
	return &Memory{items: make(map[int]model.Notification)}
}

func (m *Memory) Save(item model.Notification) error {
	m.items[item.ID] = item

	return nil
}

func (m *Memory) ByID(id int) (model.Notification, error) {
	item, ok := m.items[id]
	if !ok {
		return model.Notification{}, ErrNotFound
	}

	return item, nil
}
