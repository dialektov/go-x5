package handler

import "github.com/dialektov/go-x5/hw6/internal/model"

type Sender interface {
	Send(channel, message string) (model.Notification, error)
}

type Handler struct {
	notifications Sender
}

type SendRequest struct {
	Channel string
	Message string
}

func New(notifications Sender) *Handler {
	return &Handler{notifications: notifications}
}

func (h *Handler) Send(request SendRequest) (model.Notification, error) {
	return h.notifications.Send(request.Channel, request.Message)
}
