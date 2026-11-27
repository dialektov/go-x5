package memory

import (
	"context"
	"testing"

	"example.com/order-service/internal/order"
	"github.com/stretchr/testify/require"
)

func TestUpdateRejectsStaleStatus(t *testing.T) {
	repository := NewOrderRepository()
	created, err := repository.Create(context.Background(), order.Order{
		Items:  []order.Item{{Name: "Keyboard", Quantity: 1, Price: 100}},
		Status: order.StatusNew,
	})
	require.NoError(t, err)

	created.Status = order.StatusProcessing
	updated, err := repository.Update(context.Background(), created, order.StatusNew)
	require.NoError(t, err)

	updated.Status = order.StatusCancelled
	_, err = repository.Update(context.Background(), updated, order.StatusNew)
	require.ErrorIs(t, err, order.ErrInvalidStatusTransition)
}
