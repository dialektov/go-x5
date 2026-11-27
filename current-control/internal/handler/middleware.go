package handler

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"net/http"
	"runtime/debug"
	"strconv"
	"sync/atomic"
	"time"
)

type contextKey string

const requestIDKey contextKey = "request_id"

var fallbackRequestCounter atomic.Uint64

type statusRecorder struct {
	http.ResponseWriter
	status      int
	wroteHeader bool
}

func (r *statusRecorder) WriteHeader(status int) {
	if !r.wroteHeader {
		r.status = status
		r.wroteHeader = true
	}
	r.ResponseWriter.WriteHeader(status)
}

func (r *statusRecorder) Write(body []byte) (int, error) {
	if !r.wroteHeader {
		r.status = http.StatusOK
		r.wroteHeader = true
	}

	return r.ResponseWriter.Write(body)
}

func requestLoggingMiddleware(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			startedAt := time.Now()
			requestID := newRequestID()
			recorder := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
			w.Header().Set("X-Request-ID", requestID)

			ctx := context.WithValue(r.Context(), requestIDKey, requestID)
			next.ServeHTTP(recorder, r.WithContext(ctx))

			logger.Info("http request",
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.Int("status", recorder.status),
				slog.Duration("duration", time.Since(startedAt)),
				slog.String("request_id", requestID),
			)
		})
	}
}

func recoveryMiddleware(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				value := recover()
				if value == nil {
					return
				}

				requestID, _ := r.Context().Value(requestIDKey).(string)
				logger.Error("panic recovered",
					slog.Any("panic", value),
					slog.String("method", r.Method),
					slog.String("path", r.URL.Path),
					slog.String("request_id", requestID),
					slog.String("stack", string(debug.Stack())),
				)

				if recorder, ok := w.(*statusRecorder); ok && recorder.wroteHeader {
					return
				}
				writeJSON(w, http.StatusInternalServerError, ErrorResponse{
					Error: "internal server error",
					Code:  "INTERNAL_ERROR",
				})
			}()

			next.ServeHTTP(w, r)
		})
	}
}

func newRequestID() string {
	var buf [16]byte
	if _, err := rand.Read(buf[:]); err != nil {
		return strconv.FormatUint(fallbackRequestCounter.Add(1), 16)
	}

	return hex.EncodeToString(buf[:])
}
