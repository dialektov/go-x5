package service

import (
	"context"
	"fmt"
	"math"
	"sort"
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
	items, total, err := normalizeItems(input.Items)
	if err != nil {
		return order.Order{}, err
	}

	createdAt := time.Now().UTC()
	value := order.Order{
		Items:      items,
		TotalPrice: total,
		Status:     order.StatusNew,
		CreatedAt:  createdAt,
		History: []order.StatusChange{{
			Status:    order.StatusNew,
			ChangedAt: createdAt,
		}},
	}

	created, err := s.repository.Create(ctx, value)
	if err != nil {
		return order.Order{}, fmt.Errorf("create order: %w", err)
	}

	return created, nil
}

func (s *OrderService) GetByID(ctx context.Context, id int64) (order.Order, error) {
	if id <= 0 {
		return order.Order{}, fmt.Errorf("%w: order id must be positive", order.ErrInvalidOrder)
	}

	found, err := s.repository.GetByID(ctx, id)
	if err != nil {
		return order.Order{}, fmt.Errorf("get order %d: %w", id, err)
	}

	return found, nil
}

func (s *OrderService) List(ctx context.Context) ([]order.Order, error) {
	items, err := s.repository.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("list orders: %w", err)
	}
	if items == nil {
		return []order.Order{}, nil
	}

	return items, nil
}

func (s *OrderService) UpdateStatus(ctx context.Context, id int64, status order.Status) (order.Order, error) {
	if id <= 0 {
		return order.Order{}, fmt.Errorf("%w: order id must be positive", order.ErrInvalidOrder)
	}
	if !status.IsValid() {
		return order.Order{}, fmt.Errorf("%w: unknown status %q", order.ErrInvalidStatus, status)
	}

	current, err := s.repository.GetByID(ctx, id)
	if err != nil {
		return order.Order{}, fmt.Errorf("get order %d: %w", id, err)
	}
	if !allowedTransition(current.Status, status) {
		return order.Order{}, fmt.Errorf("%w: %s -> %s", order.ErrInvalidStatusTransition, current.Status, status)
	}

	expectedStatus := current.Status
	current.Status = status
	current.History = append(current.History, order.StatusChange{
		Status:    status,
		ChangedAt: time.Now().UTC(),
	})

	updated, err := s.repository.Update(ctx, current, expectedStatus)
	if err != nil {
		return order.Order{}, fmt.Errorf("update order %d: %w", id, err)
	}

	return updated, nil
}

func (s *OrderService) History(ctx context.Context, id int64) ([]order.StatusChange, error) {
	if id <= 0 {
		return nil, fmt.Errorf("%w: order id must be positive", order.ErrInvalidOrder)
	}

	current, err := s.repository.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get order %d history: %w", id, err)
	}

	history := make([]order.StatusChange, len(current.History))
	copy(history, current.History)
	sort.SliceStable(history, func(i, j int) bool {
		return history[i].ChangedAt.Before(history[j].ChangedAt)
	})

	return history, nil
}

func normalizeItems(items []order.Item) ([]order.Item, int64, error) {
	if len(items) == 0 {
		return nil, 0, fmt.Errorf("%w: order must contain at least one item", order.ErrInvalidOrder)
	}

	normalized := make([]order.Item, len(items))
	var total int64
	for index, item := range items {
		item.Name = strings.TrimSpace(item.Name)
		if item.Name == "" {
			return nil, 0, fmt.Errorf("%w: item name cannot be empty", order.ErrInvalidOrder)
		}
		if item.Quantity <= 0 {
			return nil, 0, fmt.Errorf("%w: item quantity must be positive", order.ErrInvalidOrder)
		}
		if item.Price <= 0 {
			return nil, 0, fmt.Errorf("%w: item price must be positive", order.ErrInvalidOrder)
		}

		quantity := int64(item.Quantity)
		if item.Price > math.MaxInt64/quantity {
			return nil, 0, fmt.Errorf("%w: item total price overflow", order.ErrInvalidOrder)
		}
		lineTotal := item.Price * quantity
		if total > math.MaxInt64-lineTotal {
			return nil, 0, fmt.Errorf("%w: order total price overflow", order.ErrInvalidOrder)
		}

		normalized[index] = item
		total += lineTotal
	}

	return normalized, total, nil
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
