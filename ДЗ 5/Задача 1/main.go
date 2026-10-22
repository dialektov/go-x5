package main

import (
	"errors"
	"fmt"
	"strings"
)

type MultiError struct {
	errs []error
}

func (m *MultiError) Add(err error) {
	m.errs = append(m.errs, err)
}

func (m *MultiError) Error() string {
	messages := make([]string, 0, len(m.errs))
	for _, err := range m.errs {
		messages = append(messages, err.Error())
	}
	return strings.Join(messages, "; ")
}

type Customer struct {
	Age  int
	Name string
}

// ValidateWrong возвращает интерфейс error, внутри которого лежит nil-указатель.
func (c Customer) ValidateWrong() error {
	var validationErrors *MultiError

	if c.Age < 0 {
		validationErrors = &MultiError{}
		validationErrors.Add(errors.New("возраст не может быть отрицательным"))
	}
	if c.Name == "" {
		if validationErrors == nil {
			validationErrors = &MultiError{}
		}
		validationErrors.Add(errors.New("имя не может быть пустым"))
	}

	return validationErrors
}

// ValidateCorrect не помещает nil-указатель внутрь интерфейса error.
func (c Customer) ValidateCorrect() error {
	var validationErrors *MultiError

	if c.Age < 0 {
		validationErrors = &MultiError{}
		validationErrors.Add(errors.New("возраст не может быть отрицательным"))
	}
	if c.Name == "" {
		if validationErrors == nil {
			validationErrors = &MultiError{}
		}
		validationErrors.Add(errors.New("имя не может быть пустым"))
	}

	if validationErrors != nil {
		return validationErrors
	}
	return nil
}

func printErrorSafely(err error) {
	defer func() {
		if value := recover(); value != nil {
			fmt.Println("перехвачена паника:", value)
		}
	}()

	fmt.Println("текст ошибки:", err.Error())
}

func main() {
	customer := Customer{Age: 33, Name: "Иван"}

	wrongErr := customer.ValidateWrong()
	fmt.Println("ValidateWrong вернула nil:", wrongErr == nil)
	printErrorSafely(wrongErr)

	correctErr := customer.ValidateCorrect()
	fmt.Println("ValidateCorrect вернула nil:", correctErr == nil)
}
