package handler

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strconv"

	"example.com/order-service/internal/order"
	"github.com/go-chi/chi/v5"
	httpSwagger "github.com/swaggo/http-swagger"
)

type OrderService interface {
	Create(ctx context.Context, input order.CreateOrderInput) (order.Order, error)
	GetByID(ctx context.Context, id int64) (order.Order, error)
	List(ctx context.Context) ([]order.Order, error)
	UpdateStatus(ctx context.Context, id int64, status order.Status) (order.Order, error)
	History(ctx context.Context, id int64) ([]order.StatusChange, error)
}

type Handler struct {
	service OrderService
	logger  *slog.Logger
}

type ErrorResponse struct {
	Error string `json:"error"`
	Code  string `json:"code"`
}

type healthResponse struct {
	Status string `json:"status" example:"ok"`
}

func NewRouter(service OrderService, logger *slog.Logger) http.Handler {
	h := &Handler{service: service, logger: logger}
	router := chi.NewRouter()
	router.Use(requestLoggingMiddleware(logger))
	router.Use(recoveryMiddleware(logger))

	router.Get("/health", h.health)
	router.Post("/api/v1/orders", h.createOrder)
	router.Get("/api/v1/orders", h.listOrders)
	router.Get("/api/v1/orders/{id}", h.getOrder)
	router.Patch("/api/v1/orders/{id}/status", h.updateStatus)
	router.Get("/api/v1/orders/{id}/history", h.orderHistory)
	router.Get("/swagger/*", httpSwagger.WrapHandler)

	return router
}

// health godoc
// @Summary Проверка доступности сервиса
// @Tags health
// @Produce json
// @Success 200 {object} healthResponse
// @Router /health [get]
func (h *Handler) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, healthResponse{Status: "ok"})
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func (h *Handler) writeServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, order.ErrOrderNotFound):
		writeJSON(w, http.StatusNotFound, ErrorResponse{Error: "order not found", Code: "ORDER_NOT_FOUND"})
	case errors.Is(err, order.ErrInvalidStatusTransition):
		writeJSON(w, http.StatusConflict, ErrorResponse{Error: "status transition is not allowed", Code: "INVALID_STATUS_TRANSITION"})
	case errors.Is(err, order.ErrInvalidOrder):
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid order", Code: "INVALID_ORDER"})
	case errors.Is(err, order.ErrInvalidStatus):
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid status", Code: "INVALID_STATUS"})
	case errors.Is(err, errInvalidID):
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid order id", Code: "INVALID_ID"})
	default:
		h.logger.Error("request failed", slog.String("error", err.Error()))
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "internal server error", Code: "INTERNAL_ERROR"})
	}
}

var errInvalidID = errors.New("invalid order id")

func orderID(r *http.Request) (int64, error) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil || id <= 0 {
		return 0, errInvalidID
	}

	return id, nil
}

func decodeJSON(r *http.Request, destination any) error {
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(destination); err != nil {
		return err
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return errors.New("unexpected trailing data in request body")
	}

	return nil
}
