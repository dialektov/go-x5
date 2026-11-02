package service

import (
	"errors"
	"testing"

	"github.com/dialektov/go-x5/hw6/internal/notifier"
	"github.com/dialektov/go-x5/hw6/internal/repository"
)

func TestSendSavesNotification(t *testing.T) {
	repo := repository.NewMemory()
	notifications := NewService(repo, map[string]notifier.Notifier{
		"email": notifier.EmailNotifier{BaseNotifier: notifier.BaseNotifier{Channel: "email"}},
	})

	item, err := notifications.Send("email", "заказ готов")
	if err != nil {
		t.Fatalf("Send вернула ошибку: %v", err)
	}

	saved, err := repo.ByID(item.ID)
	if err != nil {
		t.Fatalf("ByID вернула ошибку: %v", err)
	}
	if saved.Channel != "email" || saved.Message != "заказ готов" {
		t.Fatalf("сохранено %+v", saved)
	}
}

func TestSendRejectsEmptyMessage(t *testing.T) {
	notifications := NewService(repository.NewMemory(), nil)

	_, err := notifications.Send("email", "")
	if !errors.Is(err, notifier.ErrEmptyMessage) {
		t.Fatalf("ошибка = %v", err)
	}
}

func TestSendRejectsUnknownChannel(t *testing.T) {
	notifications := NewService(repository.NewMemory(), map[string]notifier.Notifier{})

	_, err := notifications.Send("sms", "заказ готов")
	if !errors.Is(err, ErrUnknownChannel) {
		t.Fatalf("ошибка = %v", err)
	}
}
