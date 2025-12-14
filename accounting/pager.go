package accounting

import (
	"context"
	"fmt"
)

// Iterator provides a simple interface for iterating over paginated results.
//
// Iterator abstracts away pagination details, allowing users to iterate
// through all results without manually managing offset/limit parameters.
//
// Example:
//
//	iter := dealsService.ListIter(ctx, companyID, opts)
//	for iter.Next() {
//	    deal := iter.Value()
//	    fmt.Printf("Deal ID: %d\n", deal.Id)
//	}
//	if err := iter.Err(); err != nil {
//	    log.Fatal(err)
//	}
type Iterator[T any] interface {
	// Next advances the iterator to the next item.
	// Returns true if there is a next item, false if iteration is complete.
	// After Next returns false, call Err() to check for errors.
	Next() bool

	// Value returns the current item.
	// Only valid after a successful call to Next().
	Value() T

	// Err returns any error that occurred during iteration.
	// Should be called after Next() returns false to distinguish
	// between normal completion and errors.
	Err() error
}

// PageFetcher is a function that fetches a single page of results.
//
// Parameters:
//   - ctx: Context for the request
//   - offset: The offset for pagination (starting from 0)
//   - limit: Maximum number of items to return
//
// Returns:
//   - items: Slice of items for this page
//   - totalCount: Total number of items across all pages
//   - error: Any error that occurred
type PageFetcher[T any] func(ctx context.Context, offset, limit int64) (items []T, totalCount int64, err error)

// pager implements the Iterator interface for paginated API results.
type pager[T any] struct {
	ctx         context.Context
	fetcher     PageFetcher[T]
	limit       int64
	offset      int64
	totalCount  int64
	items       []T
	currentIdx  int
	fetchedOnce bool
	err         error
}

// NewPager creates a new iterator for paginated results.
//
// Parameters:
//   - ctx: Context for API requests
//   - fetcher: Function that fetches a page of results
//   - limit: Number of items to fetch per page (default: 20, max: 100)
//
// Example:
//
//	fetcher := func(ctx context.Context, offset, limit int64) ([]gen.Deal, int64, error) {
//	    opts := &ListDealsOptions{
//	        Offset: &offset,
//	        Limit:  &limit,
//	        Type:   stringPtr("expense"),
//	    }
//	    result, err := dealsService.List(ctx, companyID, opts)
//	    if err != nil {
//	        return nil, 0, err
//	    }
//	    return result.Deals, result.TotalCount, nil
//	}
//	iter := NewPager(ctx, fetcher, 50)
func NewPager[T any](ctx context.Context, fetcher PageFetcher[T], limit int64) Iterator[T] {
	if limit <= 0 {
		limit = 20 // Default limit
	}
	if limit > 100 {
		limit = 100 // Max limit
	}

	return &pager[T]{
		ctx:     ctx,
		fetcher: fetcher,
		limit:   limit,
		offset:  0,
		items:   nil,
	}
}

// Next advances the iterator to the next item.
func (p *pager[T]) Next() bool {
	// If we have an error from a previous fetch, stop iteration
	if p.err != nil {
		return false
	}

	// If we have items in the current page, advance within the page
	if p.items != nil && p.currentIdx < len(p.items)-1 {
		p.currentIdx++
		return true
	}

	// Check if we've already fetched all pages (only if totalCount is known and > 0)
	if p.fetchedOnce && p.totalCount > 0 && p.offset >= p.totalCount {
		return false
	}

	// Fetch the next page
	items, totalCount, err := p.fetcher(p.ctx, p.offset, p.limit)
	if err != nil {
		p.err = fmt.Errorf("failed to fetch page: %w", err)
		return false
	}

	// Update state
	p.fetchedOnce = true
	p.totalCount = totalCount
	p.items = items
	p.currentIdx = 0

	// If no items returned, we're done
	if len(items) == 0 {
		return false
	}

	// Move offset for next page
	p.offset += int64(len(items))

	return true
}

// Value returns the current item.
func (p *pager[T]) Value() T {
	if p.items == nil || p.currentIdx >= len(p.items) {
		var zero T
		return zero
	}
	return p.items[p.currentIdx]
}

// Err returns any error that occurred during iteration.
func (p *pager[T]) Err() error {
	return p.err
}
