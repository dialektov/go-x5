package order

import "testing"

func TestCreateOrder(t *testing.T) {
	order, err := CreateOrder(1)
	if err != nil {
		t.Fatalf("CreateOrder вернула ошибку: %v", err)
	}
	if order.ID != 1 {
		t.Errorf("ID = %d, ожидался 1", order.ID)
	}
	if order.Status != StatusNew {
		t.Errorf("статус = %q, ожидался %q", order.Status, StatusNew)
	}
	if total := CalculateTotal(order); total != 0 {
		t.Errorf("сумма пустого заказа = %d, ожидался 0", total)
	}
}

func TestCreateOrderWithInvalidID(t *testing.T) {
	invalidIDs := []int{0, -1}

	for _, id := range invalidIDs {
		order, err := CreateOrder(id)
		if err == nil {
			t.Errorf("CreateOrder(%d) не вернула ошибку", id)
		}
		if order != nil {
			t.Errorf("CreateOrder(%d) вернула заказ при ошибке", id)
		}
	}
}

func TestAddItemsAndCalculateTotal(t *testing.T) {
	order, err := CreateOrder(1)
	if err != nil {
		t.Fatalf("CreateOrder вернула ошибку: %v", err)
	}

	items := []Item{
		{Name: "Книга", Price: 12000},
		{Name: "Ручка", Price: 1500},
	}

	for _, item := range items {
		if err := AddItem(order, item); err != nil {
			t.Fatalf("AddItem вернула ошибку: %v", err)
		}
	}

	if total := CalculateTotal(order); total != 13500 {
		t.Errorf("сумма заказа = %d, ожидалась 13500", total)
	}
}

func TestAddItemWithInvalidData(t *testing.T) {
	order, err := CreateOrder(1)
	if err != nil {
		t.Fatalf("CreateOrder вернула ошибку: %v", err)
	}

	tests := []struct {
		name  string
		order *Order
		item  Item
	}{
		{
			name:  "nil вместо заказа",
			order: nil,
			item:  Item{Name: "Книга", Price: 12000},
		},
		{
			name:  "пустое название",
			order: order,
			item:  Item{Name: "", Price: 12000},
		},
		{
			name:  "отрицательная цена",
			order: order,
			item:  Item{Name: "Книга", Price: -1},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if err := AddItem(test.order, test.item); err == nil {
				t.Error("AddItem не вернула ошибку")
			}
		})
	}

	if len(order.Items) != 0 {
		t.Errorf("после ошибок в заказе появилось %d позиций", len(order.Items))
	}
}

func TestChangingOrderCopyDoesNotChangeOriginal(t *testing.T) {
	original, err := CreateOrder(1)
	if err != nil {
		t.Fatalf("CreateOrder вернула ошибку: %v", err)
	}

	copyOfOrder := *original
	copyOfOrder.ID = 2
	copyOfOrder.Status = StatusPaid

	if original.ID != 1 {
		t.Errorf("ID исходного заказа изменился на %d", original.ID)
	}
	if original.Status != StatusNew {
		t.Errorf("статус исходного заказа изменился на %q", original.Status)
	}
}

func TestStatusNames(t *testing.T) {
	if StatusNames == nil {
		t.Fatal("StatusNames не создана через make")
	}

	if StatusNames[StatusNew] != "Новый" {
		t.Errorf("название статуса %q = %q", StatusNew, StatusNames[StatusNew])
	}

	StatusNames["cancelled"] = "Отменён"
	if StatusNames["cancelled"] != "Отменён" {
		t.Error("не удалось записать значение в StatusNames")
	}
}
