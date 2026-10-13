package main

import "fmt"

func main() {
	orders := []int{12000, 7500, 31000, 4500}

	if len(orders) == 0 {
		fmt.Println("Срез пуст")
		return
	}

	minimum := orders[0]
	maximum := orders[0]
	sum := 0

	for _, order := range orders {
		if order < minimum {
			minimum = order
		}
		if order > maximum {
			maximum = order
		}
		sum += order
	}

	average := float64(sum) / float64(len(orders))

	fmt.Println("Минимум:", minimum)
	fmt.Println("Максимум:", maximum)
	fmt.Println("Сумма:", sum)
	fmt.Println("Среднее:", average)
}
