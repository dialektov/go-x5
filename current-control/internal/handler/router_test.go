package handler

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"example.com/order-service/internal/repository/memory"
	"example.com/order-service/internal/service"
	"github.com/stretchr/testify/require"
)

func TestOrderAPI(t *testing.T) {
	server := httptest.NewServer(testRouter())
	t.Cleanup(server.Close)

	response := request(t, server, http.MethodGet, "/health", nil)
	require.Equal(t, http.StatusOK, response.Code)
	require.Contains(t, response.Body.String(), `"status":"ok"`)

	response = request(t, server, http.MethodPost, "/api/v1/orders", map[string]any{
		"items": []map[string]any{
			{"name": "Keyboard", "quantity": 1, "price": 500000},
			{"name": "Mouse", "quantity": 2, "price": 150000},
		},
	})
	require.Equal(t, http.StatusCreated, response.Code)
	created := decodeOrder(t, response)
	require.Equal(t, float64(800000), created["total_price"])
	require.Equal(t, "new", created["status"])

	response = request(t, server, http.MethodGet, "/api/v1/orders", nil)
	require.Equal(t, http.StatusOK, response.Code)

	response = request(t, server, http.MethodGet, "/api/v1/orders/1", nil)
	require.Equal(t, http.StatusOK, response.Code)

	response = request(t, server, http.MethodPatch, "/api/v1/orders/1/status", map[string]any{"status": "processing"})
	require.Equal(t, http.StatusOK, response.Code)
	require.Equal(t, "processing", decodeOrder(t, response)["status"])

	response = request(t, server, http.MethodGet, "/api/v1/orders/1/history", nil)
	require.Equal(t, http.StatusOK, response.Code)
	var history []map[string]any
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &history))
	require.Equal(t, []string{"new", "processing"}, []string{history[0]["status"].(string), history[1]["status"].(string)})

	response = request(t, server, http.MethodPost, "/api/v1/orders", []byte("{"))
	require.Equal(t, http.StatusBadRequest, response.Code)
	require.Contains(t, response.Body.String(), `"code":"INVALID_REQUEST"`)

	response = request(t, server, http.MethodGet, "/api/v1/orders/99", nil)
	require.Equal(t, http.StatusNotFound, response.Code)
	require.Contains(t, response.Body.String(), `"code":"ORDER_NOT_FOUND"`)

	response = request(t, server, http.MethodPatch, "/api/v1/orders/1/status", map[string]any{"status": "new"})
	require.Equal(t, http.StatusConflict, response.Code)
	require.Contains(t, response.Body.String(), `"code":"INVALID_STATUS_TRANSITION"`)
	require.Contains(t, response.Body.String(), "status transition is not allowed")
}

func TestRecoveryMiddleware(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	protected := recoveryMiddleware(logger)(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		panic("boom")
	}))

	request := httptest.NewRequest(http.MethodGet, "/panic", nil)
	recorder := httptest.NewRecorder()
	protected.ServeHTTP(recorder, request)

	require.Equal(t, http.StatusInternalServerError, recorder.Code)
	require.Contains(t, recorder.Body.String(), `"code":"INTERNAL_ERROR"`)
}

func testRouter() http.Handler {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	return NewRouter(service.NewOrderService(memory.NewOrderRepository()), logger)
}

func request(t *testing.T, server *httptest.Server, method, path string, body any) *httptest.ResponseRecorder {
	t.Helper()

	var reader io.Reader
	switch value := body.(type) {
	case nil:
	case []byte:
		reader = bytes.NewReader(value)
	default:
		encoded, err := json.Marshal(value)
		require.NoError(t, err)
		reader = bytes.NewReader(encoded)
	}

	httpRequest, err := http.NewRequest(method, server.URL+path, reader)
	require.NoError(t, err)
	response, err := http.DefaultClient.Do(httpRequest)
	require.NoError(t, err)
	defer response.Body.Close()

	payload, err := io.ReadAll(response.Body)
	require.NoError(t, err)
	recorder := httptest.NewRecorder()
	recorder.Code = response.StatusCode
	_, err = recorder.Body.Write(payload)
	require.NoError(t, err)

	return recorder
}

func decodeOrder(t *testing.T, response *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	var orderBody map[string]any
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &orderBody))

	return orderBody
}
