package accounting

import (
	"context"
	"errors"
	"testing"
)

func TestPager_SinglePage(t *testing.T) {
	ctx := context.Background()

	// Create a fetcher that returns 5 items on the first page
	items := []int{1, 2, 3, 4, 5}
	fetcher := func(ctx context.Context, offset, limit int64) ([]int, int64, error) {
		if offset != 0 {
			return []int{}, 5, nil
		}
		return items, 5, nil
	}

	iter := NewPager(ctx, fetcher, 10)

	// Iterate through all items
	var result []int
	for iter.Next() {
		result = append(result, iter.Value())
	}

	// Check error
	if err := iter.Err(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify results
	if len(result) != len(items) {
		t.Fatalf("expected %d items, got %d", len(items), len(result))
	}
	for i, v := range result {
		if v != items[i] {
			t.Errorf("item %d: expected %d, got %d", i, items[i], v)
		}
	}
}

func TestPager_MultiplePages(t *testing.T) {
	ctx := context.Background()

	// Create test data across 3 pages
	allItems := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	pageSize := 3

	fetcher := func(ctx context.Context, offset, limit int64) ([]int, int64, error) {
		start := int(offset)
		end := start + int(limit)
		if start >= len(allItems) {
			return []int{}, int64(len(allItems)), nil
		}
		if end > len(allItems) {
			end = len(allItems)
		}
		return allItems[start:end], int64(len(allItems)), nil
	}

	iter := NewPager(ctx, fetcher, int64(pageSize))

	// Iterate through all items
	var result []int
	for iter.Next() {
		result = append(result, iter.Value())
	}

	// Check error
	if err := iter.Err(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify results
	if len(result) != len(allItems) {
		t.Fatalf("expected %d items, got %d", len(allItems), len(result))
	}
	for i, v := range result {
		if v != allItems[i] {
			t.Errorf("item %d: expected %d, got %d", i, allItems[i], v)
		}
	}
}

func TestPager_EmptyResult(t *testing.T) {
	ctx := context.Background()

	fetcher := func(ctx context.Context, offset, limit int64) ([]int, int64, error) {
		return []int{}, 0, nil
	}

	iter := NewPager(ctx, fetcher, 10)

	// Should return false immediately
	if iter.Next() {
		t.Error("expected Next() to return false for empty result")
	}

	// Check no error
	if err := iter.Err(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestPager_FetchError(t *testing.T) {
	ctx := context.Background()

	expectedErr := errors.New("fetch failed")
	fetcher := func(ctx context.Context, offset, limit int64) ([]int, int64, error) {
		return nil, 0, expectedErr
	}

	iter := NewPager(ctx, fetcher, 10)

	// First Next() should trigger fetch and return false
	if iter.Next() {
		t.Error("expected Next() to return false on fetch error")
	}

	// Error should be available
	err := iter.Err()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, expectedErr) {
		t.Errorf("expected error to wrap %v, got %v", expectedErr, err)
	}
}

func TestPager_ErrorOnSecondPage(t *testing.T) {
	ctx := context.Background()

	expectedErr := errors.New("second page fetch failed")
	fetchCount := 0

	fetcher := func(ctx context.Context, offset, limit int64) ([]int, int64, error) {
		fetchCount++
		if fetchCount == 1 {
			// First page succeeds
			return []int{1, 2, 3}, 10, nil
		}
		// Second page fails
		return nil, 0, expectedErr
	}

	iter := NewPager(ctx, fetcher, 3)

	// First 3 items should succeed
	count := 0
	for iter.Next() {
		count++
		if count > 3 {
			break
		}
	}

	// Should have gotten 3 items
	if count != 3 {
		t.Errorf("expected to get 3 items, got %d", count)
	}

	// Error should be available
	err := iter.Err()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, expectedErr) {
		t.Errorf("expected error to wrap %v, got %v", expectedErr, err)
	}
}

func TestPager_LimitValidation(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name          string
		inputLimit    int64
		expectedLimit int64
	}{
		{
			name:          "zero limit defaults to 20",
			inputLimit:    0,
			expectedLimit: 20,
		},
		{
			name:          "negative limit defaults to 20",
			inputLimit:    -1,
			expectedLimit: 20,
		},
		{
			name:          "limit above 100 capped at 100",
			inputLimit:    200,
			expectedLimit: 100,
		},
		{
			name:          "valid limit used as-is",
			inputLimit:    50,
			expectedLimit: 50,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actualLimit := int64(0)
			testFetcher := func(ctx context.Context, offset, limit int64) ([]int, int64, error) {
				actualLimit = limit
				return []int{1}, 1, nil
			}

			iter := NewPager(ctx, testFetcher, tt.inputLimit)
			iter.Next()

			if actualLimit != tt.expectedLimit {
				t.Errorf("expected limit %d, got %d", tt.expectedLimit, actualLimit)
			}
		})
	}
}

func TestPager_ValueBeforeNext(t *testing.T) {
	ctx := context.Background()
	fetcher := func(ctx context.Context, offset, limit int64) ([]int, int64, error) {
		return []int{1, 2, 3}, 3, nil
	}

	iter := NewPager(ctx, fetcher, 10)

	// Calling Value() before Next() should return zero value
	val := iter.Value()
	if val != 0 {
		t.Errorf("expected zero value before Next(), got %d", val)
	}
}

func TestPager_ValueAfterCompletion(t *testing.T) {
	ctx := context.Background()
	fetcher := func(ctx context.Context, offset, limit int64) ([]int, int64, error) {
		if offset == 0 {
			return []int{1}, 1, nil
		}
		return []int{}, 1, nil
	}

	iter := NewPager(ctx, fetcher, 10)

	// Exhaust the iterator
	lastValue := 0
	for iter.Next() {
		lastValue = iter.Value()
	}

	// Calling Value() after completion should return the last value
	// This matches the behavior of other Go iterators (like database/sql.Rows)
	val := iter.Value()
	if val != lastValue {
		t.Errorf("expected last value %d after completion, got %d", lastValue, val)
	}
}

func TestPager_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() // Ensure cancel is always called

	fetcher := func(ctx context.Context, offset, limit int64) ([]int, int64, error) {
		// Check if context is cancelled
		select {
		case <-ctx.Done():
			return nil, 0, ctx.Err()
		default:
			return []int{1, 2, 3}, 10, nil
		}
	}

	iter := NewPager(ctx, fetcher, 3)

	// Get first page
	count := 0
	for iter.Next() {
		count++
		if count == 3 {
			// Cancel context before fetching next page
			cancel()
		}
		if count > 3 {
			break
		}
	}

	// Should have stopped after first page
	if count != 3 {
		t.Errorf("expected 3 items before cancellation, got %d", count)
	}

	// Should have context cancellation error
	err := iter.Err()
	if err == nil {
		t.Fatal("expected context cancellation error, got nil")
	}
}

func TestPager_PartialLastPage(t *testing.T) {
	ctx := context.Background()

	// Total of 7 items, with page size 3 (3 + 3 + 1)
	allItems := []int{1, 2, 3, 4, 5, 6, 7}
	pageSize := 3

	fetcher := func(ctx context.Context, offset, limit int64) ([]int, int64, error) {
		start := int(offset)
		end := start + int(limit)
		if start >= len(allItems) {
			return []int{}, int64(len(allItems)), nil
		}
		if end > len(allItems) {
			end = len(allItems)
		}
		return allItems[start:end], int64(len(allItems)), nil
	}

	iter := NewPager(ctx, fetcher, int64(pageSize))

	// Iterate through all items
	var result []int
	for iter.Next() {
		result = append(result, iter.Value())
	}

	// Check error
	if err := iter.Err(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify results
	if len(result) != len(allItems) {
		t.Fatalf("expected %d items, got %d", len(allItems), len(result))
	}
	for i, v := range result {
		if v != allItems[i] {
			t.Errorf("item %d: expected %d, got %d", i, allItems[i], v)
		}
	}
}
