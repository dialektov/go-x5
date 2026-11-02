package notifier

import (
	"errors"
	"fmt"
)

var ErrEmptyMessage = errors.New("пустое сообщение")

type Notifier interface {
	Send(msg string) error
}

type BaseNotifier struct {
	Channel string
}

func (b BaseNotifier) Log(msg string) {
	fmt.Println(b.Channel, msg)
}

type EmailNotifier struct {
	BaseNotifier
}

type PushNotifier struct {
	BaseNotifier
}

func (n EmailNotifier) Send(msg string) error {
	if msg == "" {
		return ErrEmptyMessage
	}

	n.Log(msg)

	return nil
}

func (n PushNotifier) Send(msg string) error {
	if msg == "" {
		return ErrEmptyMessage
	}

	n.Log(msg)

	return nil
}
