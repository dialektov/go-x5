package main

import (
	"errors"
	"fmt"
)

var ErrInvalidOrder = errors.New("некорректный заказ")

type OrderError struct {
	Code    string
	OrderID int
	Err     error
}

func (e *OrderError) Error() string {
	return fmt.Sprintf("заказ %d, код %q: %v", e.OrderID, e.Code, e.Err)
}

func (e *OrderError) Unwrap() error {
	return e.Err
}

func (e *OrderError) Is(target error) bool {
	targetOrderError, ok := target.(*OrderError)
	if !ok {
		return false
	}
	return e.Code == targetOrderError.Code
}

func createOrder(id int) error {
	if id <= 0 {
		return &OrderError{
			Code:    "invalid_id",
			OrderID: id,
			Err:     ErrInvalidOrder,
		}
	}
	return nil
}

func main() {
	err := createOrder(-1)

	fmt.Println("ошибка:", err)
	fmt.Println("найдена исходная ошибка через Unwrap:", errors.Is(err, ErrInvalidOrder))

	sameCode := &OrderError{Code: "invalid_id"}
	fmt.Println("совпадает код через Is:", errors.Is(err, sameCode))

	var orderError *OrderError
	if errors.As(err, &orderError) {
		fmt.Println("ID из кастомной ошибки:", orderError.OrderID)
	}
}
