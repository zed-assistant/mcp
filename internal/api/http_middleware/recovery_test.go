//nolint:fatcontext,containedctx
package httpmiddleware_test

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	httpmiddleware "github.com/zed-assistant/mcp/internal/api/http_middleware"
)

func discardLogger() *slog.Logger {
	return slog.New(slog.DiscardHandler)
}

func TestRecoveryMiddlewareNormalRequest(t *testing.T) {
	t.Parallel()

	handlerCalled := false

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})

	middleware := httpmiddleware.RecoveryMiddleware(discardLogger())
	wrappedHandler := middleware(handler)

	req := httptest.NewRequestWithContext(context.Background(), "GET", "/test", http.NoBody)
	w := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(w, req)

	assert.True(t, handlerCalled)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `{"status":"ok"}`, w.Body.String())
}

func TestRecoveryMiddlewarePanicRecovered(t *testing.T) {
	t.Parallel()

	panicMessage := "something went wrong"

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic(panicMessage)
	})

	middleware := httpmiddleware.RecoveryMiddleware(discardLogger())
	wrappedHandler := middleware(handler)

	req := httptest.NewRequestWithContext(context.Background(), "GET", "/test", http.NoBody)
	w := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response map[string]string
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)
	assert.Equal(t, "Unexpected server error", response["message"])
}

func TestRecoveryMiddlewarePanicWithHeadersAlreadyWritten(t *testing.T) {
	t.Parallel()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("started response"))
		panic("panic after headers written")
	})

	middleware := httpmiddleware.RecoveryMiddleware(discardLogger())
	wrappedHandler := middleware(handler)

	req := httptest.NewRequestWithContext(context.Background(), "GET", "/test", http.NoBody)
	w := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "started response", w.Body.String())
}

func TestRecoveryMiddlewareErrAbortHandlerRePanics(t *testing.T) {
	t.Parallel()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic(http.ErrAbortHandler)
	})

	middleware := httpmiddleware.RecoveryMiddleware(discardLogger())
	wrappedHandler := middleware(handler)

	req := httptest.NewRequestWithContext(context.Background(), "GET", "/test", http.NoBody)
	w := httptest.NewRecorder()

	assert.Panics(t, func() {
		wrappedHandler.ServeHTTP(w, req)
	})
}

func TestRecoveryMiddlewareResponseContentType(t *testing.T) {
	t.Parallel()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("test panic")
	})

	middleware := httpmiddleware.RecoveryMiddleware(discardLogger())
	wrappedHandler := middleware(handler)

	req := httptest.NewRequestWithContext(context.Background(), "GET", "/test", http.NoBody)
	w := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(w, req)

	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
}

func TestRecoveryMiddlewareLogsStackTrace(t *testing.T) {
	t.Parallel()

	var logBuffer bytes.Buffer
	logHandler := slog.NewJSONHandler(&logBuffer, nil)
	logger := slog.New(logHandler)

	panicMessage := "critical failure"

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic(panicMessage)
	})

	middleware := httpmiddleware.RecoveryMiddleware(logger)
	wrappedHandler := middleware(handler)

	req := httptest.NewRequestWithContext(context.Background(), "GET", "/test", http.NoBody)
	w := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	logOutput := logBuffer.String()
	assert.Contains(t, logOutput, "Panic recovered")
	assert.Contains(t, logOutput, panicMessage)
	assert.Contains(t, logOutput, "stack")
}

func TestRecoveryMiddlewareMultiplePanics(t *testing.T) {
	t.Parallel()

	var handlerCount int

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCount++
		panic("panic in request")
	})

	middleware := httpmiddleware.RecoveryMiddleware(discardLogger())
	wrappedHandler := middleware(handler)

	req1 := httptest.NewRequestWithContext(context.Background(), "GET", "/test", http.NoBody)
	w1 := httptest.NewRecorder()
	wrappedHandler.ServeHTTP(w1, req1)
	assert.Equal(t, http.StatusInternalServerError, w1.Code)

	req2 := httptest.NewRequestWithContext(context.Background(), "GET", "/test", http.NoBody)
	w2 := httptest.NewRecorder()
	wrappedHandler.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusInternalServerError, w2.Code)

	assert.Equal(t, 2, handlerCount)
}
