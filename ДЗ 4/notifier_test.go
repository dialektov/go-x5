package notify

import (
	"errors"
	"testing"
)

type callStub struct {
	calls []string
	err   error
}

func (s *callStub) Send(msg string) error {
	s.calls = append(s.calls, msg)
	return s.err
}

func TestDispatchCallsEveryNotifier(t *testing.T) {
	email := &callStub{}
	push := &callStub{}

	err := Dispatch([]Notifier{email, push}, "заказ готов")
	if err != nil {
		t.Fatalf("Dispatch вернула ошибку: %v", err)
	}

	if len(email.calls) != 1 || email.calls[0] != "заказ готов" {
		t.Fatalf("email вызван с %v", email.calls)
	}
	if len(push.calls) != 1 || push.calls[0] != "заказ готов" {
		t.Fatalf("push вызван с %v", push.calls)
	}
}

func TestDispatchStopsAfterError(t *testing.T) {
	email := &callStub{err: ErrEmptyMessage}
	push := &callStub{}

	err := Dispatch([]Notifier{email, push}, "заказ готов")
	if !errors.Is(err, ErrEmptyMessage) {
		t.Fatalf("ошибка = %v", err)
	}
	if len(push.calls) != 0 {
		t.Fatalf("push вызван после ошибки: %v", push.calls)
	}
}

func TestEmailAndPush(t *testing.T) {
	email := EmailNotifier{BaseNotifier: BaseNotifier{Channel: "email"}}
	push := PushNotifier{BaseNotifier: BaseNotifier{Channel: "push"}}

	if err := Dispatch([]Notifier{email, push}, "заказ готов"); err != nil {
		t.Fatalf("Dispatch вернула ошибку: %v", err)
	}
	if err := email.Send(""); !errors.Is(err, ErrEmptyMessage) {
		t.Fatalf("пустой email: %v", err)
	}
	if err := push.Send(""); !errors.Is(err, ErrEmptyMessage) {
		t.Fatalf("пустой push: %v", err)
	}
	if got := Channel(email); got != "email" {
		t.Fatalf("канал email = %q", got)
	}
	if got := Channel(push); got != "push" {
		t.Fatalf("канал push = %q", got)
	}
	if got := Channel(&callStub{}); got != "unknown" {
		t.Fatalf("неизвестный канал = %q", got)
	}
}
