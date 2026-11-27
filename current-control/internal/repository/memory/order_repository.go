package memory

import (
	"context"
	"sync"

	"example.com/order-service/internal/order"
)

type OrderRepository struct {
	mu     sync.RWMutex
	items  []order.Order
	nextID int64
}

func NewOrderRepository() *OrderRepository {
	return &OrderRepository{
		items:  make([]order.Order, 0),
		nextID: 1,
	}
}

func (r *OrderRepository) Create(_ context.Context, value order.Order) (order.Order, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	value.ID = r.nextID
	r.nextID++
	stored := cloneOrder(value)
	r.items = append(r.items, stored)

	return cloneOrder(stored), nil
}

func (r *OrderRepository) GetByID(_ context.Context, id int64) (order.Order, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, item := range r.items {
		if item.ID == id {
			return cloneOrder(item), nil
		}
	}

	return order.Order{}, order.ErrOrderNotFound
}

func (r *OrderRepository) List(_ context.Context) ([]order.Order, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]order.Order, 0, len(r.items))
	for _, item := range r.items {
		result = append(result, cloneOrder(item))
	}

	return result, nil
}

func (r *OrderRepository) Update(_ context.Context, value order.Order) (order.Order, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for index, item := range r.items {
		if item.ID == value.ID {
			stored := cloneOrder(value)
			r.items[index] = stored

			return cloneOrder(stored), nil
		}
	}

	return order.Order{}, order.ErrOrderNotFound
}

func cloneOrder(value order.Order) order.Order {
	cloned := value
	if value.Items != nil {
		cloned.Items = append([]order.Item(nil), value.Items...)
	}
	if value.History != nil {
		cloned.History = append([]order.StatusChange(nil), value.History...)
	}

	return cloned
}
