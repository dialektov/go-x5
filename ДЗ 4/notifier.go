package notify

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

func Dispatch(all []Notifier, msg string) error {
	for _, n := range all {
		if err := n.Send(msg); err != nil {
			return err
		}
	}
	return nil
}

func Channel(n Notifier) string {
	switch v := n.(type) {
	case EmailNotifier:
		return v.Channel
	case PushNotifier:
		return v.Channel
	default:
		return "unknown"
	}
}

var (
	_ Notifier = EmailNotifier{}
	_ Notifier = PushNotifier{}
)
