package fake

import (
	"context"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.praktikum-services.ru/Stasyan/momo-store/internal/store/dumplings"
)

func TestStoreListProducts(t *testing.T) {
	store := NewStore()

	products, err := store.ListProducts(context.Background())
	require.NoError(t, err)
	assert.Empty(t, products)

	packs := []dumplings.Product{
		{ID: 1, Name: "Пельмени"},
		{ID: 2, Name: "Хинкали"},
	}
	store.SetAvailablePacks(packs...)

	products, err = store.ListProducts(context.Background())
	require.NoError(t, err)
	assert.Equal(t, packs, products)
}

func TestStoreCreateOrderIncrementsID(t *testing.T) {
	store := NewStore()
	ctx := context.Background()

	for want := int64(1); want <= 5; want++ {
		id, err := store.CreateOrder(ctx)
		require.NoError(t, err)
		assert.Equal(t, want, id)
	}
}

func TestStoreCreateOrderIsConcurrencySafe(t *testing.T) {
	const (
		goroutines     = 8
		ordersPerGoroutine = 25
		totalOrders    = goroutines * ordersPerGoroutine
	)

	store := NewStore()
	ctx := context.Background()
	ids := make(chan int64, totalOrders)

	var wg sync.WaitGroup
	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()

			for j := 0; j < ordersPerGoroutine; j++ {
				id, err := store.CreateOrder(ctx)
				if err != nil {
					t.Errorf("CreateOrder returned error: %v", err)
					return
				}
				ids <- id
			}
		}()
	}
	wg.Wait()
	close(ids)

	seen := make(map[int64]struct{}, totalOrders)
	for id := range ids {
		assert.NotContains(t, seen, id)
		seen[id] = struct{}{}
	}
	assert.Len(t, seen, totalOrders)
}
