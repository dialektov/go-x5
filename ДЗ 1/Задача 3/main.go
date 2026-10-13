package main

import "fmt"

func main() {
	var amount int

	fmt.Print("Введите сумму заказа в копейках: ")
	fmt.Scan(&amount)

	if amount < 0 {
		fmt.Println("Ошибка: сумма заказа не может быть отрицательной")
		return
	}

	switch {
	case amount < 10000:
		fmt.Println("small")
	case amount < 50000:
		fmt.Println("regular")
	default:
		fmt.Println("large")
	}
}
