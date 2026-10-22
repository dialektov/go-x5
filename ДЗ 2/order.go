package order

import "errors"

const (
	StatusNew       = "new"
	StatusPaid      = "paid"
	StatusCompleted = "completed"
)

var (
	ErrInvalidID    = errors.New("ID заказа должен быть положительным")
	ErrNilOrder     = errors.New("заказ не задан")
	ErrEmptyName    = errors.New("название позиции не может быть пустым")
	ErrNegativeCost = errors.New("цена не может быть отрицательной")
)

var StatusNames = createStatusNames()

type Item struct {
	Name  string
	Price int
}

type Order struct {
	ID     int
	Status string
	Items  []Item
}

func createStatusNames() map[string]string {
	statusNames := make(map[string]string)
	statusNames[StatusNew] = "Новый"
	statusNames[StatusPaid] = "Оплачен"
	statusNames[StatusCompleted] = "Завершён"
	return statusNames
}

func CreateOrder(id int) (*Order, error) {
	if id <= 0 {
		return nil, ErrInvalidID
	}

	return &Order{
		ID:     id,
		Status: StatusNew,
		Items:  make([]Item, 0),
	}, nil
}

func AddItem(order *Order, item Item) error {
	if order == nil {
		return ErrNilOrder
	}
	if item.Name == "" {
		return ErrEmptyName
	}
	if item.Price < 0 {
		return ErrNegativeCost
	}

	order.Items = append(order.Items, item)
	return nil
}

func CalculateTotal(order *Order) int {
	if order == nil {
		return 0
	}

	total := 0
	for _, item := range order.Items {
		total += item.Price
	}
	return total
}
