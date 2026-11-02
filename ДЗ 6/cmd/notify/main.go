package main

import (
	"fmt"
	"os"

	"github.com/dialektov/go-x5/hw6/internal/config"
	"github.com/dialektov/go-x5/hw6/internal/handler"
	"github.com/dialektov/go-x5/hw6/internal/notifier"
	"github.com/dialektov/go-x5/hw6/internal/repository"
	"github.com/dialektov/go-x5/hw6/internal/service"
)

func main() {
	repo := repository.NewMemory()
	notifications := service.NewService(repo, map[string]notifier.Notifier{
		"email": notifier.EmailNotifier{BaseNotifier: notifier.BaseNotifier{Channel: "email"}},
		"push":  notifier.PushNotifier{BaseNotifier: notifier.BaseNotifier{Channel: "push"}},
	})
	h := handler.New(notifications)

	message := "заказ готов"
	if len(os.Args) > 1 {
		message = os.Args[1]
	}

	item, err := h.Send(handler.SendRequest{
		Channel: config.Channel(),
		Message: message,
	})
	if err != nil {
		fmt.Println("ошибка:", err)

		return
	}

	fmt.Printf("отправлено: %+v\n", item)
}
