package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.praktikum-services.ru/Stasyan/momo-store/cmd/api/app"
	"gitlab.praktikum-services.ru/Stasyan/momo-store/cmd/api/dependencies"
)

func newTestRouter(t *testing.T) http.Handler {
	t.Helper()

	store, err := dependencies.NewFakeDumplingsStore()
	require.NoError(t, err)

	instance, err := app.NewInstance(store)
	require.NoError(t, err)

	router, err := newRouter(instance)
	require.NoError(t, err)

	return router
}

func serve(router http.Handler, method, target string) *httptest.ResponseRecorder {
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(method, target, nil))

	return recorder
}

func TestRouterRoutes(t *testing.T) {
	router := newTestRouter(t)

	tests := []struct {
		name     string
		method   string
		target   string
		wantCode int
		wantJSON bool
	}{
		{name: "health", method: http.MethodGet, target: "/health", wantCode: http.StatusOK},
		{name: "products", method: http.MethodGet, target: "/products", wantCode: http.StatusOK, wantJSON: true},
		{name: "products with trailing slash", method: http.MethodGet, target: "/products/", wantCode: http.StatusOK, wantJSON: true},
		{name: "categories", method: http.MethodGet, target: "/categories", wantCode: http.StatusOK, wantJSON: true},
		{name: "whoami", method: http.MethodGet, target: "/auth/whoami", wantCode: http.StatusOK, wantJSON: true},
		{name: "metrics", method: http.MethodGet, target: "/metrics", wantCode: http.StatusOK},
		{name: "method not allowed", method: http.MethodPost, target: "/products", wantCode: http.StatusMethodNotAllowed},
		{name: "unknown route", method: http.MethodGet, target: "/unknown", wantCode: http.StatusNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := serve(router, tt.method, tt.target)

			assert.Equal(t, tt.wantCode, recorder.Code)
			if tt.wantJSON {
				assert.Equal(t, "application/json", recorder.Header().Get("Content-Type"))
			}
		})
	}
}

func TestRouterCreateOrder(t *testing.T) {
	router := newTestRouter(t)

	recorder := serve(router, http.MethodPost, "/orders")

	require.Equal(t, http.StatusOK, recorder.Code)
	assert.JSONEq(t, `{"id":1}`, recorder.Body.String())
}

func TestRouterMetricsExposeCounters(t *testing.T) {
	router := newTestRouter(t)

	order := serve(router, http.MethodPost, "/orders")
	require.Equal(t, http.StatusOK, order.Code)

	metrics := serve(router, http.MethodGet, "/metrics")

	require.Equal(t, http.StatusOK, metrics.Code)
	body := metrics.Body.String()
	assert.Contains(t, body, "orders_count 1")
	assert.Contains(t, body, "requests_count 1")
}

func TestRouterCORS(t *testing.T) {
	router := newTestRouter(t)

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/products", nil)
	request.Header.Set("Origin", "https://momo.example")
	router.ServeHTTP(recorder, request)

	require.Equal(t, http.StatusOK, recorder.Code)
	assert.NotEmpty(t, recorder.Header().Get("Access-Control-Allow-Origin"))
	assert.Contains(t, recorder.Header().Get("Vary"), "Origin")
}
