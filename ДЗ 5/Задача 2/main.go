package main

import "fmt"

func unstableOperation() {
	panic("неожиданная ошибка драйвера")
}

func serviceWithoutRecover() {
	unstableOperation()
}

func recoverInsideService() (err error) {
	defer func() {
		if value := recover(); value != nil {
			err = fmt.Errorf("паника перехвачена в сервисе: %v", value)
		}
	}()

	unstableOperation()
	return nil
}

func recoverAtProgramBoundary() {
	defer func() {
		if value := recover(); value != nil {
			fmt.Println("паника перехвачена на верхнем уровне:", value)
		}
	}()

	serviceWithoutRecover()
}

func main() {
	err := recoverInsideService()
	fmt.Println("результат перехвата в сервисе:", err)

	recoverAtProgramBoundary()
	fmt.Println("программа продолжила работу")
}
