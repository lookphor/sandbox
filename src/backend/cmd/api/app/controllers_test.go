package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.praktikum-services.ru/Stasyan/momo-store/internal/store/dumplings"
	"gitlab.praktikum-services.ru/Stasyan/momo-store/internal/store/dumplings/fake"
)

var errStoreUnavailable = errors.New("store is unavailable")

// failingStore emulates unavailable storage
type failingStore struct{}

func (failingStore) ListProducts(context.Context) ([]dumplings.Product, error) {
	return nil, errStoreUnavailable
}

func (failingStore) CreateOrder(context.Context, ...dumplings.OrderItem) (int64, error) {
	return 0, errStoreUnavailable
}

func newTestInstance(t *testing.T, store dumplings.Store) *Instance {
	t.Helper()

	instance, err := NewInstance(store, nil)
	require.NoError(t, err)

	return instance
}

func newFakeStore(products ...dumplings.Product) *fake.Store {
	store := fake.NewStore()
	store.SetAvailablePacks(products...)

	return store
}

func testProduct(id int64) dumplings.Product {
	return dumplings.Product{
		ID:          id,
		Name:        fmt.Sprintf("Product %d", id),
		Price:       float32(id),
		Description: fmt.Sprintf("Description %d", id),
		Image:       fmt.Sprintf("https://example.com/%d.jpg", id),
	}
}

func doRequest(handler http.HandlerFunc, method, target string) *httptest.ResponseRecorder {
	recorder := httptest.NewRecorder()
	handler(recorder, httptest.NewRequest(method, target, nil))

	return recorder
}

func decodeJSON(t *testing.T, body []byte, dst interface{}) {
	t.Helper()

	require.NoError(t, json.Unmarshal(body, dst))
}

func TestListDumplingsController(t *testing.T) {
	products := []dumplings.Product{testProduct(1), testProduct(2)}

	t.Run("returns all products", func(t *testing.T) {
		instance := newTestInstance(t, newFakeStore(products...))

		recorder := doRequest(instance.ListDumplingsController, http.MethodGet, "/products")

		require.Equal(t, http.StatusOK, recorder.Code)
		assert.Equal(t, "application/json", recorder.Header().Get("Content-Type"))

		var response struct {
			Results []dumplings.Product `json:"results"`
		}
		decodeJSON(t, recorder.Body.Bytes(), &response)
		assert.Equal(t, products, response.Results)
	})

	t.Run("returns 204 on empty catalog", func(t *testing.T) {
		instance := newTestInstance(t, newFakeStore())

		recorder := doRequest(instance.ListDumplingsController, http.MethodGet, "/products")

		assert.Equal(t, http.StatusNoContent, recorder.Code)
		assert.Zero(t, recorder.Body.Len())
		assert.Empty(t, recorder.Header().Get("Content-Type"))
	})

	t.Run("returns 500 when store fails", func(t *testing.T) {
		instance := newTestInstance(t, failingStore{})

		recorder := doRequest(instance.ListDumplingsController, http.MethodGet, "/products")

		assert.Equal(t, http.StatusInternalServerError, recorder.Code)
		assert.Zero(t, recorder.Body.Len())
	})
}

func TestListDumplingsControllerCountsListings(t *testing.T) {
	instance := newTestInstance(t, newFakeStore(testProduct(1), testProduct(2)))

	doRequest(instance.ListDumplingsController, http.MethodGet, "/products")

	assert.Equal(t, 1.0, testutil.ToFloat64(instance.dumplingsListingCounter.WithLabelValues("1")))
	assert.Equal(t, 1.0, testutil.ToFloat64(instance.dumplingsListingCounter.WithLabelValues("2")))
}

func TestCreateOrderController(t *testing.T) {
	t.Run("returns new order id", func(t *testing.T) {
		instance := newTestInstance(t, newFakeStore())

		recorder := doRequest(instance.CreateOrderController, http.MethodPost, "/orders")

		require.Equal(t, http.StatusOK, recorder.Code)
		assert.Equal(t, "application/json", recorder.Header().Get("Content-Type"))

		var response struct {
			ID int64 `json:"id"`
		}
		decodeJSON(t, recorder.Body.Bytes(), &response)
		assert.Equal(t, int64(1), response.ID)
		assert.Equal(t, 1.0, testutil.ToFloat64(instance.ordersCounter))
	})

	t.Run("returns 500 when store fails", func(t *testing.T) {
		instance := newTestInstance(t, failingStore{})

		recorder := doRequest(instance.CreateOrderController, http.MethodPost, "/orders")

		assert.Equal(t, http.StatusInternalServerError, recorder.Code)
		assert.Zero(t, recorder.Body.Len())
		assert.Zero(t, testutil.ToFloat64(instance.ordersCounter))
	})
}

func TestListCategoriesController(t *testing.T) {
	instance := newTestInstance(t, newFakeStore())

	recorder := doRequest(instance.ListCategoriesController, http.MethodGet, "/categories")

	require.Equal(t, http.StatusOK, recorder.Code)
	assert.Equal(t, "application/json", recorder.Header().Get("Content-Type"))

	var categories []struct {
		ID          int64  `json:"id"`
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	decodeJSON(t, recorder.Body.Bytes(), &categories)

	require.Len(t, categories, 3)
	for i, category := range categories {
		assert.Equal(t, int64(i+1), category.ID)
		assert.NotEmpty(t, category.Name)
		assert.NotEmpty(t, category.Description)
	}
}

func TestWhoAmIController(t *testing.T) {
	instance := newTestInstance(t, newFakeStore())

	recorder := doRequest(instance.WhoAmIController, http.MethodGet, "/auth/whoami")

	require.Equal(t, http.StatusOK, recorder.Code)
	assert.Equal(t, "application/json", recorder.Header().Get("Content-Type"))

	var user struct {
		ID        int64  `json:"id"`
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		Email     string `json:"email"`
	}
	decodeJSON(t, recorder.Body.Bytes(), &user)

	assert.NotZero(t, user.ID)
	assert.NotEmpty(t, user.FirstName)
	assert.NotEmpty(t, user.LastName)
	assert.Contains(t, user.Email, "@")
}

func TestHealthcheckController(t *testing.T) {
	instance := newTestInstance(t, newFakeStore())

	recorder := doRequest(instance.HealthcheckController, http.MethodGet, "/health")

	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Zero(t, recorder.Body.Len())
}
