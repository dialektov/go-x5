package service

import (
	"context"
	"strings"
	"time"

	"example.com/order-service/internal/order"
)

type OrderService struct {
	repository order.Repository
}

func NewOrderService(repository order.Repository) *OrderService {
	return &OrderService{repository: repository}
}

func (s *OrderService) Create(ctx context.Context, input order.CreateOrderInput) (order.Order, error) {
	items, err := normalizeItems(input.Items)
	if err != nil {
		return order.Order{}, err
	}

	createdAt := time.Now().UTC()
	value := order.Order{
		Items:      items,
		TotalPrice: totalPrice(items),
		Status:     order.StatusNew,
		CreatedAt:  createdAt,
		History: []order.StatusChange{{
			Status:    order.StatusNew,
			ChangedAt: createdAt,
		}},
	}

	return s.repository.Create(ctx, value)
}

func (s *OrderService) GetByID(ctx context.Context, id int64) (order.Order, error) {
	return s.repository.GetByID(ctx, id)
}

func (s *OrderService) List(ctx context.Context) ([]order.Order, error) {
	items, err := s.repository.List(ctx)
	if err != nil {
		return nil, err
	}
	if items == nil {
		return []order.Order{}, nil
	}

	return items, nil
}

func (s *OrderService) UpdateStatus(ctx context.Context, id int64, status order.Status) (order.Order, error) {
	if !knownStatus(status) {
		return order.Order{}, order.ErrInvalidStatus
	}

	current, err := s.repository.GetByID(ctx, id)
	if err != nil {
		return order.Order{}, err
	}
	if !allowedTransition(current.Status, status) {
		return order.Order{}, order.ErrInvalidStatusTransition
	}

	current.Status = status
	current.History = append(current.History, order.StatusChange{
		Status:    status,
		ChangedAt: time.Now().UTC(),
	})

	return s.repository.Update(ctx, current)
}

func (s *OrderService) History(ctx context.Context, id int64) ([]order.StatusChange, error) {
	current, err := s.repository.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	history := make([]order.StatusChange, len(current.History))
	copy(history, current.History)

	return history, nil
}

func normalizeItems(items []order.Item) ([]order.Item, error) {
	if len(items) == 0 {
		return nil, order.ErrInvalidOrder
	}

	normalized := make([]order.Item, len(items))
	for index, item := range items {
		item.Name = strings.TrimSpace(item.Name)
		if item.Name == "" || item.Quantity <= 0 || item.Price <= 0 {
			return nil, order.ErrInvalidOrder
		}
		normalized[index] = item
	}

	return normalized, nil
}

func totalPrice(items []order.Item) int64 {
	var total int64
	for _, item := range items {
		total += item.Price * int64(item.Quantity)
	}

	return total
}

func knownStatus(status order.Status) bool {
	switch status {
	case order.StatusNew, order.StatusProcessing, order.StatusCompleted, order.StatusCancelled:
		return true
	default:
		return false
	}
}

func allowedTransition(from, to order.Status) bool {
	switch from {
	case order.StatusNew:
		return to == order.StatusProcessing || to == order.StatusCancelled
	case order.StatusProcessing:
		return to == order.StatusCompleted || to == order.StatusCancelled
	default:
		return false
	}
}
