package handler

import (
	"net/http"

	"example.com/order-service/internal/order"
)

// createOrder godoc
// @Summary Создание заказа
// @Tags orders
// @Accept json
// @Produce json
// @Param request body order.CreateOrderInput true "Позиции заказа"
// @Success 201 {object} order.Order
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/orders [post]
func (h *Handler) createOrder(w http.ResponseWriter, r *http.Request) {
	var input order.CreateOrderInput
	if err := decodeJSON(r, &input); err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid request body", Code: "INVALID_REQUEST"})
		return
	}

	created, err := h.service.Create(r.Context(), input)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, created)
}

// listOrders godoc
// @Summary Получение списка заказов
// @Tags orders
// @Produce json
// @Success 200 {array} order.Order
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/orders [get]
func (h *Handler) listOrders(w http.ResponseWriter, r *http.Request) {
	orders, err := h.service.List(r.Context())
	if err != nil {
		h.writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, orders)
}

// getOrder godoc
// @Summary Получение заказа
// @Tags orders
// @Produce json
// @Param id path int true "Идентификатор заказа"
// @Success 200 {object} order.Order
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/orders/{id} [get]
func (h *Handler) getOrder(w http.ResponseWriter, r *http.Request) {
	id, err := orderID(r)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}

	found, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, found)
}

// updateStatus godoc
// @Summary Изменение статуса заказа
// @Tags orders
// @Accept json
// @Produce json
// @Param id path int true "Идентификатор заказа"
// @Param request body order.UpdateStatusInput true "Новый статус"
// @Success 200 {object} order.Order
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 409 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/orders/{id}/status [patch]
func (h *Handler) updateStatus(w http.ResponseWriter, r *http.Request) {
	id, err := orderID(r)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}

	var input order.UpdateStatusInput
	if err := decodeJSON(r, &input); err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid request body", Code: "INVALID_REQUEST"})
		return
	}

	updated, err := h.service.UpdateStatus(r.Context(), id, input.Status)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, updated)
}

// orderHistory godoc
// @Summary Получение истории статусов
// @Tags orders
// @Produce json
// @Param id path int true "Идентификатор заказа"
// @Success 200 {array} order.StatusChange
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/orders/{id}/history [get]
func (h *Handler) orderHistory(w http.ResponseWriter, r *http.Request) {
	id, err := orderID(r)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}

	history, err := h.service.History(r.Context(), id)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, history)
}
