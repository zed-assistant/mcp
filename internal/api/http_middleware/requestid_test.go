//nolint:fatcontext,containedctx
package httpmiddleware_test

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	httpmiddleware "github.com/zed-assistant/mcp/internal/api/http_middleware"
	"github.com/zed-assistant/mcp/internal/appctx"
)

func testDiscardLogger() *slog.Logger {
	return slog.New(slog.DiscardHandler)
}

func TestRequestIdMiddlewareIgnoresProvidedHeader(t *testing.T) {
	t.Parallel()

	requestID := uuid.NewString()
	var capturedCtx context.Context

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedCtx = r.Context()
		w.WriteHeader(http.StatusOK)
	})

	middleware := httpmiddleware.RequestIdMiddleware(testDiscardLogger())
	wrappedHandler := middleware(handler)

	req := httptest.NewRequestWithContext(context.Background(), "GET", "/test", http.NoBody)
	req.Header.Set("X-Request-Id", requestID)
	w := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(w, req)

	responseID := w.Header().Get("X-Request-Id")
	assert.NotEmpty(t, responseID)
	assert.NotEqual(t, requestID, responseID)

	_, err := uuid.Parse(responseID)
	require.NoError(t, err)

	contextID := appctx.GetRequestId(capturedCtx)
	assert.Equal(t, responseID, contextID)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRequestIdMiddlewareGeneratesUUID(t *testing.T) {
	t.Parallel()

	var capturedCtx context.Context

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedCtx = r.Context()
		w.WriteHeader(http.StatusOK)
	})

	middleware := httpmiddleware.RequestIdMiddleware(testDiscardLogger())
	wrappedHandler := middleware(handler)

	req := httptest.NewRequestWithContext(context.Background(), "GET", "/test", http.NoBody)
	w := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(w, req)

	responseID := w.Header().Get("X-Request-Id")
	assert.NotEmpty(t, responseID)

	parsedID, err := uuid.Parse(responseID)
	require.NoError(t, err)
	assert.NotEqual(t, uuid.UUID{}, parsedID)

	contextID := appctx.GetRequestId(capturedCtx)
	assert.Equal(t, responseID, contextID)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRequestIdMiddlewareAlwaysSucceedsRegardlessOfHeader(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		requestID string
	}{
		{
			name:      "invalid UUID format",
			requestID: "not-a-uuid",
		},
		{
			name:      "empty string generates new UUID",
			requestID: "",
		},
		{
			name:      "malformed UUID",
			requestID: "12345678-1234-1234-1234",
		},
		{
			name:      "random gibberish",
			requestID: "this-is-not-valid-at-all!@#$",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			handlerCalled := false

			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				handlerCalled = true
				w.WriteHeader(http.StatusOK)
			})

			middleware := httpmiddleware.RequestIdMiddleware(testDiscardLogger())
			wrappedHandler := middleware(handler)

			req := httptest.NewRequestWithContext(context.Background(), "GET", "/test", http.NoBody)
			if tt.requestID != "" {
				req.Header.Set("X-Request-Id", tt.requestID)
			}
			w := httptest.NewRecorder()

			wrappedHandler.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)
			assert.True(t, handlerCalled)

			responseID := w.Header().Get("X-Request-Id")
			assert.NotEmpty(t, responseID)
			_, err := uuid.Parse(responseID)
			require.NoError(t, err)
		})
	}
}

func TestGetRequestId(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		ctx       context.Context
		wantEmpty bool
	}{
		{
			name:      "returns empty string when not in context",
			ctx:       context.Background(),
			wantEmpty: true,
		},
		{
			name:      "returns empty string with nil context",
			ctx:       nil,
			wantEmpty: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := appctx.GetRequestId(tt.ctx)
			assert.Empty(t, got)
		})
	}
}

func TestGetRequestIdViaMiddleware(t *testing.T) {
	t.Parallel()

	requestID := uuid.NewString()
	var capturedCtx context.Context

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedCtx = r.Context()
		w.WriteHeader(http.StatusOK)
	})

	middleware := httpmiddleware.RequestIdMiddleware(testDiscardLogger())
	wrappedHandler := middleware(handler)

	req := httptest.NewRequestWithContext(context.Background(), "GET", "/test", http.NoBody)
	req.Header.Set("X-Request-Id", requestID)
	w := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(w, req)

	responseID := w.Header().Get("X-Request-Id")

	got := appctx.GetRequestId(capturedCtx)
	assert.Equal(t, responseID, got)
	assert.NotEqual(t, requestID, got)
}

func TestRequestIdMiddlewareMultipleRequests(t *testing.T) {
	t.Parallel()

	var ids [2]string
	var ctxs [2]context.Context

	for i := range 2 {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctxs[i] = r.Context()
			w.WriteHeader(http.StatusOK)
		})

		middleware := httpmiddleware.RequestIdMiddleware(testDiscardLogger())
		wrappedHandler := middleware(handler)

		req := httptest.NewRequestWithContext(context.Background(), "GET", "/test", http.NoBody)
		w := httptest.NewRecorder()

		wrappedHandler.ServeHTTP(w, req)

		ids[i] = w.Header().Get("X-Request-Id")
	}

	assert.NotEmpty(t, ids[0])
	assert.NotEmpty(t, ids[1])
	assert.NotEqual(t, ids[0], ids[1])
}

func TestRequestIdMiddlewareGeneratesFreshIdForEachProvidedUUID(t *testing.T) {
	t.Parallel()

	providedUUIDs := []string{
		uuid.NewString(),
		uuid.NewString(),
		uuid.NewString(),
	}

	for _, providedID := range providedUUIDs {
		t.Run("provided_uuid_"+providedID[:8], func(t *testing.T) {
			t.Parallel()

			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			middleware := httpmiddleware.RequestIdMiddleware(testDiscardLogger())
			wrappedHandler := middleware(handler)

			req := httptest.NewRequestWithContext(context.Background(), "GET", "/test", http.NoBody)
			req.Header.Set("X-Request-Id", providedID)
			w := httptest.NewRecorder()

			wrappedHandler.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)

			responseID := w.Header().Get("X-Request-Id")
			assert.NotEqual(t, providedID, responseID)
			_, err := uuid.Parse(responseID)
			require.NoError(t, err)
		})
	}
}

func TestRequestIdMiddlewareResponseHeader(t *testing.T) {
	t.Parallel()

	requestID := uuid.NewString()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := httpmiddleware.RequestIdMiddleware(testDiscardLogger())
	wrappedHandler := middleware(handler)

	req := httptest.NewRequestWithContext(context.Background(), "GET", "/test", http.NoBody)
	req.Header.Set("X-Request-Id", requestID)
	w := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(w, req)

	responseID := w.Header().Get("X-Request-Id")
	assert.NotEmpty(t, responseID)
	assert.NotEqual(t, requestID, responseID)
	assert.Equal(t, http.StatusOK, w.Code)
}
