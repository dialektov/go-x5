package service_test

import (
	"context"
	"math"
	"testing"
	"time"

	"example.com/order-service/internal/order"
	"example.com/order-service/internal/service"
	"github.com/stretchr/testify/require"
)

func TestCreateOrder(t *testing.T) {
	orderService := service.NewOrderService(newFakeRepository())

	created, err := orderService.Create(context.Background(), order.CreateOrderInput{
		Items: []order.Item{
			{Name: " Keyboard ", Quantity: 1, Price: 500000},
			{Name: "Mouse", Quantity: 2, Price: 150000},
		},
	})

	require.NoError(t, err)
	require.Equal(t, int64(1), created.ID)
	require.Equal(t, int64(800000), created.TotalPrice)
	require.Equal(t, order.StatusNew, created.Status)
	require.False(t, created.CreatedAt.IsZero())
	require.Equal(t, []order.Item{
		{Name: "Keyboard", Quantity: 1, Price: 500000},
		{Name: "Mouse", Quantity: 2, Price: 150000},
	}, created.Items)
	require.Equal(t, []order.Status{order.StatusNew}, statuses(created.History))
	require.Equal(t, created.CreatedAt, created.History[0].ChangedAt)
}

func TestCreateOrderValidation(t *testing.T) {
	tests := []struct {
		name  string
		items []order.Item
	}{
		{name: "nil items", items: nil},
		{name: "no items", items: []order.Item{}},
		{name: "empty name", items: []order.Item{{Name: " ", Quantity: 1, Price: 100}}},
		{name: "zero quantity", items: []order.Item{{Name: "Keyboard", Quantity: 0, Price: 100}}},
		{name: "negative quantity", items: []order.Item{{Name: "Keyboard", Quantity: -1, Price: 100}}},
		{name: "zero price", items: []order.Item{{Name: "Keyboard", Quantity: 1, Price: 0}}},
		{name: "negative price", items: []order.Item{{Name: "Keyboard", Quantity: 1, Price: -5}}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			orderService := service.NewOrderService(newFakeRepository())
			_, err := orderService.Create(context.Background(), order.CreateOrderInput{Items: test.items})
			require.ErrorIs(t, err, order.ErrInvalidOrder)
		})
	}
}

func TestGetAndListOrders(t *testing.T) {
	repository := newFakeRepository()
	orderService := service.NewOrderService(repository)
	first := mustCreate(t, orderService, "Keyboard", 1, 100)
	second := mustCreate(t, orderService, "Mouse", 1, 200)

	found, err := orderService.GetByID(context.Background(), first.ID)
	require.NoError(t, err)
	require.Equal(t, first.ID, found.ID)

	_, err = orderService.GetByID(context.Background(), 99)
	require.ErrorIs(t, err, order.ErrOrderNotFound)

	orders, err := orderService.List(context.Background())
	require.NoError(t, err)
	require.Equal(t, []int64{first.ID, second.ID}, []int64{orders[0].ID, orders[1].ID})

	repository.listErr = order.ErrNotImplemented
	_, err = orderService.List(context.Background())
	require.ErrorIs(t, err, order.ErrNotImplemented)
}

func TestStatusTransitions(t *testing.T) {
	tests := []struct {
		name    string
		steps   []order.Status
		wantErr error
	}{
		{name: "new to processing", steps: []order.Status{order.StatusProcessing}},
		{name: "new to cancelled", steps: []order.Status{order.StatusCancelled}},
		{name: "processing to completed", steps: []order.Status{order.StatusProcessing, order.StatusCompleted}},
		{name: "processing to cancelled", steps: []order.Status{order.StatusProcessing, order.StatusCancelled}},
		{name: "new to completed", steps: []order.Status{order.StatusCompleted}, wantErr: order.ErrInvalidStatusTransition},
		{name: "completed stays final", steps: []order.Status{order.StatusProcessing, order.StatusCompleted, order.StatusCancelled}, wantErr: order.ErrInvalidStatusTransition},
		{name: "cancelled stays final", steps: []order.Status{order.StatusCancelled, order.StatusProcessing}, wantErr: order.ErrInvalidStatusTransition},
		{name: "unknown status", steps: []order.Status{"shipped"}, wantErr: order.ErrInvalidStatus},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			orderService := service.NewOrderService(newFakeRepository())
			created := mustCreate(t, orderService, "Keyboard", 1, 100)

			var err error
			current := created
			for _, status := range test.steps {
				current, err = orderService.UpdateStatus(context.Background(), created.ID, status)
				if err != nil {
					break
				}
			}

			if test.wantErr == nil {
				require.NoError(t, err)
				require.Equal(t, test.steps[len(test.steps)-1], current.Status)
				return
			}
			require.ErrorIs(t, err, test.wantErr)
		})
	}
}

func TestStatusHistoryIsChronological(t *testing.T) {
	orderService := service.NewOrderService(newFakeRepository())
	created := mustCreate(t, orderService, "Keyboard", 1, 100)
	_, err := orderService.UpdateStatus(context.Background(), created.ID, order.StatusProcessing)
	require.NoError(t, err)
	_, err = orderService.UpdateStatus(context.Background(), created.ID, order.StatusCompleted)
	require.NoError(t, err)

	history, err := orderService.History(context.Background(), created.ID)
	require.NoError(t, err)
	require.Equal(t, []order.Status{order.StatusNew, order.StatusProcessing, order.StatusCompleted}, statuses(history))
	for position := 1; position < len(history); position++ {
		require.False(t, history[position].ChangedAt.Before(history[position-1].ChangedAt))
	}

	_, err = orderService.History(context.Background(), 99)
	require.ErrorIs(t, err, order.ErrOrderNotFound)
}

func TestRepositoryErrors(t *testing.T) {
	repository := newFakeRepository()
	repository.createErr = order.ErrNotImplemented
	orderService := service.NewOrderService(repository)

	_, err := orderService.Create(context.Background(), order.CreateOrderInput{
		Items: []order.Item{{Name: "Keyboard", Quantity: 1, Price: 100}},
	})
	require.ErrorIs(t, err, order.ErrNotImplemented)

	repository.createErr = nil
	created := mustCreate(t, orderService, "Keyboard", 1, 100)
	repository.updateErr = order.ErrNotImplemented
	_, err = orderService.UpdateStatus(context.Background(), created.ID, order.StatusProcessing)
	require.ErrorIs(t, err, order.ErrNotImplemented)
}

func TestListReturnsEmptySlice(t *testing.T) {
	repository := newFakeRepository()
	repository.emptyList = true
	orders, err := service.NewOrderService(repository).List(context.Background())
	require.NoError(t, err)
	require.NotNil(t, orders)
	require.Empty(t, orders)
}

type fakeRepository struct {
	orders    map[int64]order.Order
	sequence  []int64
	nextID    int64
	createErr error
	updateErr error
	listErr   error
	emptyList bool
}

func newFakeRepository() *fakeRepository {
	return &fakeRepository{
		orders: make(map[int64]order.Order),
		nextID: 1,
	}
}

func (r *fakeRepository) Create(_ context.Context, value order.Order) (order.Order, error) {
	if r.createErr != nil {
		return order.Order{}, r.createErr
	}
	value.ID = r.nextID
	r.nextID++
	stored := cloneOrder(value)
	r.orders[stored.ID] = stored
	r.sequence = append(r.sequence, stored.ID)

	return cloneOrder(stored), nil
}

func (r *fakeRepository) GetByID(_ context.Context, id int64) (order.Order, error) {
	value, ok := r.orders[id]
	if !ok {
		return order.Order{}, order.ErrOrderNotFound
	}

	return cloneOrder(value), nil
}

func (r *fakeRepository) List(_ context.Context) ([]order.Order, error) {
	if r.listErr != nil {
		return nil, r.listErr
	}
	if r.emptyList {
		return nil, nil
	}

	result := make([]order.Order, 0, len(r.sequence))
	for _, id := range r.sequence {
		result = append(result, cloneOrder(r.orders[id]))
	}

	return result, nil
}

func (r *fakeRepository) Update(_ context.Context, value order.Order, expectedStatus order.Status) (order.Order, error) {
	if r.updateErr != nil {
		return order.Order{}, r.updateErr
	}
	current, ok := r.orders[value.ID]
	if !ok {
		return order.Order{}, order.ErrOrderNotFound
	}
	if current.Status != expectedStatus {
		return order.Order{}, order.ErrInvalidStatusTransition
	}
	stored := cloneOrder(value)
	r.orders[stored.ID] = stored

	return cloneOrder(stored), nil
}

func cloneOrder(value order.Order) order.Order {
	cloned := value
	cloned.Items = append([]order.Item(nil), value.Items...)
	cloned.History = append([]order.StatusChange(nil), value.History...)

	return cloned
}

func mustCreate(t *testing.T, orderService *service.OrderService, name string, quantity int, price int64) order.Order {
	t.Helper()
	created, err := orderService.Create(context.Background(), order.CreateOrderInput{
		Items: []order.Item{{Name: name, Quantity: quantity, Price: price}},
	})
	require.NoError(t, err)

	return created
}

func TestCreateRejectsPriceOverflow(t *testing.T) {
	orderService := service.NewOrderService(newFakeRepository())
	_, err := orderService.Create(context.Background(), order.CreateOrderInput{
		Items: []order.Item{{Name: "Keyboard", Quantity: 2, Price: math.MaxInt64}},
	})
	require.ErrorIs(t, err, order.ErrInvalidOrder)
}

func TestNonPositiveID(t *testing.T) {
	orderService := service.NewOrderService(newFakeRepository())

	_, err := orderService.GetByID(context.Background(), 0)
	require.ErrorIs(t, err, order.ErrInvalidOrder)

	_, err = orderService.UpdateStatus(context.Background(), -1, order.StatusProcessing)
	require.ErrorIs(t, err, order.ErrInvalidOrder)

	_, err = orderService.History(context.Background(), 0)
	require.ErrorIs(t, err, order.ErrInvalidOrder)
}

func TestHistorySortsByTime(t *testing.T) {
	repository := newFakeRepository()
	repository.nextID = 2
	repository.sequence = []int64{1}
	repository.orders[1] = order.Order{
		ID:     1,
		Status: order.StatusProcessing,
		History: []order.StatusChange{
			{Status: order.StatusProcessing, ChangedAt: time.Date(2026, 11, 27, 12, 5, 0, 0, time.UTC)},
			{Status: order.StatusNew, ChangedAt: time.Date(2026, 11, 27, 12, 0, 0, 0, time.UTC)},
		},
	}

	history, err := service.NewOrderService(repository).History(context.Background(), 1)
	require.NoError(t, err)
	require.Equal(t, []order.Status{order.StatusNew, order.StatusProcessing}, statuses(history))
}

func statuses(history []order.StatusChange) []order.Status {
	result := make([]order.Status, len(history))
	for index, change := range history {
		result[index] = change.Status
	}

	return result
}
