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

func TestRequestIdMiddlewareWithProvidedHeader(t *testing.T) {
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

	assert.Equal(t, requestID, w.Header().Get("X-Request-Id"))

	contextID := appctx.GetRequestId(capturedCtx)
	assert.Equal(t, requestID, contextID)

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

func TestRequestIdMiddlewareInvalidHeader(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		requestID    string
		wantStatus   int
		wantResponse string
	}{
		{
			name:         "invalid UUID format",
			requestID:    "not-a-uuid",
			wantStatus:   http.StatusBadRequest,
			wantResponse: "Invalid X-Request-Id header",
		},
		{
			name:       "empty string generates new UUID",
			requestID:  "",
			wantStatus: http.StatusOK,
		},
		{
			name:         "malformed UUID",
			requestID:    "12345678-1234-1234-1234",
			wantStatus:   http.StatusBadRequest,
			wantResponse: "Invalid X-Request-Id header",
		},
		{
			name:         "random gibberish",
			requestID:    "this-is-not-valid-at-all!@#$",
			wantStatus:   http.StatusBadRequest,
			wantResponse: "Invalid X-Request-Id header",
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

			assert.Equal(t, tt.wantStatus, w.Code)

			if tt.wantStatus == http.StatusBadRequest {
				assert.False(t, handlerCalled)
				assert.Contains(t, w.Body.String(), tt.wantResponse)
			} else {
				assert.True(t, handlerCalled)
			}
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

	got := appctx.GetRequestId(capturedCtx)
	assert.Equal(t, requestID, got)
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

func TestRequestIdMiddlewareValidUUIDs(t *testing.T) {
	t.Parallel()

	validUUIDs := []string{
		uuid.NewString(),
		uuid.NewString(),
		uuid.NewString(),
	}

	for _, validID := range validUUIDs {
		t.Run("valid_uuid_"+validID[:8], func(t *testing.T) {
			t.Parallel()

			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			middleware := httpmiddleware.RequestIdMiddleware(testDiscardLogger())
			wrappedHandler := middleware(handler)

			req := httptest.NewRequestWithContext(context.Background(), "GET", "/test", http.NoBody)
			req.Header.Set("X-Request-Id", validID)
			w := httptest.NewRecorder()

			wrappedHandler.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)
			assert.Equal(t, validID, w.Header().Get("X-Request-Id"))
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

	assert.Equal(t, requestID, w.Header().Get("X-Request-Id"))
	assert.Equal(t, http.StatusOK, w.Code)
}
