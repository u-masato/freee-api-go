package accounting

import (
	"context"
	"fmt"

	"github.com/muno/freee-api-go/internal/gen"
)

// ListItemsOptions contains optional parameters for listing items.
type ListItemsOptions struct {
	// StartUpdateDate filters by update date start (yyyy-mm-dd)
	StartUpdateDate *string

	// EndUpdateDate filters by update date end (yyyy-mm-dd)
	EndUpdateDate *string

	// Offset for pagination (default: 0)
	Offset *int64

	// Limit for pagination (default: 50, min: 1, max: 3000)
	Limit *int64
}

// ListItemsResult contains the result of listing items.
type ListItemsResult struct {
	// Items is the list of items
	Items []gen.Item

	// Count is the number of items returned in this response
	Count int
}

// List retrieves a list of items for the specified company.
//
// This method returns all items matching the optional filter criteria.
// Use ListItemsOptions to filter by update date range or pagination.
//
// Example:
//
//	opts := &accounting.ListItemsOptions{
//	    Limit: int64Ptr(100),
//	}
//	result, err := itemsService.List(ctx, companyID, opts)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	for _, item := range result.Items {
//	    fmt.Printf("Item ID: %d, Name: %s\n", item.Id, item.Name)
//	}
func (s *ItemsService) List(ctx context.Context, companyID int64, opts *ListItemsOptions) (*ListItemsResult, error) {
	// Build parameters
	params := &gen.GetItemsParams{
		CompanyId: companyID,
	}

	if opts != nil {
		params.StartUpdateDate = opts.StartUpdateDate
		params.EndUpdateDate = opts.EndUpdateDate
		params.Offset = opts.Offset
		params.Limit = opts.Limit
	}

	// Call the generated client
	resp, err := s.genClient.GetItemsWithResponse(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to list items: %w", err)
	}

	// Handle error responses
	if resp.JSON200 == nil {
		return nil, fmt.Errorf("unexpected response status: %s", resp.Status())
	}

	// Return the result
	return &ListItemsResult{
		Items: resp.JSON200.Items,
		Count: len(resp.JSON200.Items),
	}, nil
}

// Get retrieves a single item by ID.
//
// Example:
//
//	item, err := itemsService.Get(ctx, companyID, itemID)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("Item: %+v\n", item)
func (s *ItemsService) Get(ctx context.Context, companyID int64, itemID int64) (*gen.ItemResponse, error) {
	// Build parameters
	params := &gen.GetItemParams{
		CompanyId: companyID,
	}

	// Call the generated client
	resp, err := s.genClient.GetItemWithResponse(ctx, itemID, params)
	if err != nil {
		return nil, fmt.Errorf("failed to get item: %w", err)
	}

	// Handle error responses
	if resp.JSON200 == nil {
		return nil, fmt.Errorf("unexpected response status: %s", resp.Status())
	}

	return resp.JSON200, nil
}

// Create creates a new item.
//
// The params parameter should contain all required fields for creating an item,
// including company ID and name.
//
// Example:
//
//	params := gen.ItemParams{
//	    CompanyId: companyID,
//	    Name:      "新規品目",
//	}
//	item, err := itemsService.Create(ctx, params)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("Created item ID: %d\n", item.Item.Id)
func (s *ItemsService) Create(ctx context.Context, params gen.ItemParams) (*gen.ItemResponse, error) {
	// Call the generated client
	resp, err := s.genClient.CreateItemWithResponse(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to create item: %w", err)
	}

	// Handle error responses
	if resp.JSON201 == nil {
		return nil, fmt.Errorf("unexpected response status: %s", resp.Status())
	}

	return resp.JSON201, nil
}

// Update updates an existing item.
//
// The params parameter should contain the fields to update.
//
// Example:
//
//	params := gen.ItemParams{
//	    CompanyId: companyID,
//	    Name:      "更新後品目",
//	}
//	item, err := itemsService.Update(ctx, itemID, params)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("Updated item ID: %d\n", item.Item.Id)
func (s *ItemsService) Update(ctx context.Context, itemID int64, params gen.ItemParams) (*gen.ItemResponse, error) {
	// Call the generated client
	resp, err := s.genClient.UpdateItemWithResponse(ctx, itemID, params)
	if err != nil {
		return nil, fmt.Errorf("failed to update item: %w", err)
	}

	// Handle error responses
	if resp.JSON200 == nil {
		return nil, fmt.Errorf("unexpected response status: %s", resp.Status())
	}

	return resp.JSON200, nil
}

// Delete deletes an item by ID.
//
// Example:
//
//	err := itemsService.Delete(ctx, companyID, itemID)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println("Item deleted successfully")
func (s *ItemsService) Delete(ctx context.Context, companyID int64, itemID int64) error {
	// Build parameters
	params := &gen.DestroyItemParams{
		CompanyId: companyID,
	}

	// Call the generated client
	resp, err := s.genClient.DestroyItemWithResponse(ctx, itemID, params)
	if err != nil {
		return fmt.Errorf("failed to delete item: %w", err)
	}

	// Check for error responses
	if resp.StatusCode() >= 400 {
		return fmt.Errorf("failed to delete item: %s", resp.Status())
	}

	return nil
}

// ListIter returns an iterator for paginated item results.
//
// The iterator transparently handles pagination, automatically fetching
// new pages as needed. This is more convenient than manually managing
// offset/limit parameters.
//
// Example:
//
//	opts := &accounting.ListItemsOptions{
//	    Limit: int64Ptr(100),
//	}
//	iter := itemsService.ListIter(ctx, companyID, opts)
//	for iter.Next() {
//	    item := iter.Value()
//	    fmt.Printf("Item ID: %d, Name: %s\n", item.Id, item.Name)
//	}
//	if err := iter.Err(); err != nil {
//	    log.Fatal(err)
//	}
func (s *ItemsService) ListIter(ctx context.Context, companyID int64, opts *ListItemsOptions) Iterator[gen.Item] {
	// Determine page size (limit)
	limit := int64(50) // Default for items API
	if opts != nil && opts.Limit != nil {
		limit = *opts.Limit
	}

	// Create a fetcher function that captures the service and options
	fetcher := func(ctx context.Context, offset, limit int64) ([]gen.Item, int64, error) {
		// Create a copy of options with updated offset/limit
		fetchOpts := &ListItemsOptions{}
		if opts != nil {
			*fetchOpts = *opts
		}
		fetchOpts.Offset = &offset
		fetchOpts.Limit = &limit

		// Fetch the page
		result, err := s.List(ctx, companyID, fetchOpts)
		if err != nil {
			return nil, 0, err
		}

		// Items API doesn't return total_count, so we use -1 to indicate unknown
		// The pager will continue fetching until an empty page is returned
		totalCount := int64(-1)
		if result.Count < int(limit) {
			// This is the last page
			totalCount = offset + int64(result.Count)
		}

		return result.Items, totalCount, nil
	}

	return NewPager(ctx, fetcher, limit)
}
