package service

import (
	"errors"

	"github.com/dialektov/go-x5/hw6/internal/model"
	"github.com/dialektov/go-x5/hw6/internal/notifier"
)

var ErrUnknownChannel = errors.New("неизвестный канал")

type Repository interface {
	Save(item model.Notification) error
	ByID(id int) (model.Notification, error)
}

type Service struct {
	repo     Repository
	channels map[string]notifier.Notifier
	nextID   int
}

func NewService(repo Repository, channels map[string]notifier.Notifier) *Service {
	return &Service{
		repo:     repo,
		channels: channels,
		nextID:   1,
	}
}

func (s *Service) Send(channel, message string) (model.Notification, error) {
	if message == "" {
		return model.Notification{}, notifier.ErrEmptyMessage
	}

	sender, ok := s.channels[channel]
	if !ok {
		return model.Notification{}, ErrUnknownChannel
	}
	if err := sender.Send(message); err != nil {
		return model.Notification{}, err
	}

	item := model.Notification{
		ID:      s.nextID,
		Channel: channel,
		Message: message,
	}
	if err := s.repo.Save(item); err != nil {
		return model.Notification{}, err
	}
	s.nextID++

	return item, nil
}
