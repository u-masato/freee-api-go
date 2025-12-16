package accounting

import (
	"context"
	"fmt"

	"github.com/muno/freee-api-go/internal/gen"
)

// ListTagsOptions contains optional parameters for listing tags.
type ListTagsOptions struct {
	// StartUpdateDate filters by update date start (yyyy-mm-dd)
	StartUpdateDate *string

	// EndUpdateDate filters by update date end (yyyy-mm-dd)
	EndUpdateDate *string

	// Offset for pagination (default: 0)
	Offset *int64

	// Limit for pagination (default: 50, min: 1, max: 3000)
	Limit *int64
}

// ListTagsResult contains the result of listing tags.
type ListTagsResult struct {
	// Tags is the list of tags
	Tags []gen.Tag

	// Count is the number of tags returned in this response
	Count int
}

// List retrieves a list of tags for the specified company.
//
// This method returns all tags matching the optional filter criteria.
// Use ListTagsOptions to filter by update date range or pagination.
//
// Example:
//
//	opts := &accounting.ListTagsOptions{
//	    Limit: int64Ptr(100),
//	}
//	result, err := tagsService.List(ctx, companyID, opts)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	for _, tag := range result.Tags {
//	    fmt.Printf("Tag ID: %d, Name: %s\n", tag.Id, tag.Name)
//	}
func (s *TagsService) List(ctx context.Context, companyID int64, opts *ListTagsOptions) (*ListTagsResult, error) {
	// Build parameters
	params := &gen.GetTagsParams{
		CompanyId: companyID,
	}

	if opts != nil {
		params.StartUpdateDate = opts.StartUpdateDate
		params.EndUpdateDate = opts.EndUpdateDate
		params.Offset = opts.Offset
		params.Limit = opts.Limit
	}

	// Call the generated client
	resp, err := s.genClient.GetTagsWithResponse(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to list tags: %w", err)
	}

	// Handle error responses
	if resp.JSON200 == nil {
		return nil, fmt.Errorf("unexpected response status: %s", resp.Status())
	}

	// Return the result
	return &ListTagsResult{
		Tags:  resp.JSON200.Tags,
		Count: len(resp.JSON200.Tags),
	}, nil
}

// Get retrieves a single tag by ID.
//
// Example:
//
//	tag, err := tagsService.Get(ctx, companyID, tagID)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("Tag: %+v\n", tag)
func (s *TagsService) Get(ctx context.Context, companyID int64, tagID int64) (*gen.TagResponse, error) {
	// Build parameters
	params := &gen.GetTagParams{
		CompanyId: companyID,
	}

	// Call the generated client
	resp, err := s.genClient.GetTagWithResponse(ctx, tagID, params)
	if err != nil {
		return nil, fmt.Errorf("failed to get tag: %w", err)
	}

	// Handle error responses
	if resp.JSON200 == nil {
		return nil, fmt.Errorf("unexpected response status: %s", resp.Status())
	}

	return resp.JSON200, nil
}

// Create creates a new tag.
//
// The params parameter should contain all required fields for creating a tag,
// including company ID and name.
//
// Example:
//
//	params := gen.TagParams{
//	    CompanyId: companyID,
//	    Name:      "新規タグ",
//	}
//	tag, err := tagsService.Create(ctx, params)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("Created tag ID: %d\n", tag.Tag.Id)
func (s *TagsService) Create(ctx context.Context, params gen.TagParams) (*gen.TagResponse, error) {
	// Call the generated client
	resp, err := s.genClient.CreateTagWithResponse(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to create tag: %w", err)
	}

	// Handle error responses
	if resp.JSON201 == nil {
		return nil, fmt.Errorf("unexpected response status: %s", resp.Status())
	}

	return resp.JSON201, nil
}

// Update updates an existing tag.
//
// The params parameter should contain the fields to update.
//
// Example:
//
//	params := gen.TagParams{
//	    CompanyId: companyID,
//	    Name:      "更新後タグ",
//	}
//	tag, err := tagsService.Update(ctx, tagID, params)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("Updated tag ID: %d\n", tag.Tag.Id)
func (s *TagsService) Update(ctx context.Context, tagID int64, params gen.TagParams) (*gen.TagResponse, error) {
	// Call the generated client
	resp, err := s.genClient.UpdateTagWithResponse(ctx, tagID, params)
	if err != nil {
		return nil, fmt.Errorf("failed to update tag: %w", err)
	}

	// Handle error responses
	if resp.JSON200 == nil {
		return nil, fmt.Errorf("unexpected response status: %s", resp.Status())
	}

	return resp.JSON200, nil
}

// Delete deletes a tag by ID.
//
// Example:
//
//	err := tagsService.Delete(ctx, companyID, tagID)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println("Tag deleted successfully")
func (s *TagsService) Delete(ctx context.Context, companyID int64, tagID int64) error {
	// Build parameters
	params := &gen.DestroyTagParams{
		CompanyId: companyID,
	}

	// Call the generated client
	resp, err := s.genClient.DestroyTagWithResponse(ctx, tagID, params)
	if err != nil {
		return fmt.Errorf("failed to delete tag: %w", err)
	}

	// Check for error responses
	if resp.StatusCode() >= 400 {
		return fmt.Errorf("failed to delete tag: %s", resp.Status())
	}

	return nil
}

// ListIter returns an iterator for paginated tag results.
//
// The iterator transparently handles pagination, automatically fetching
// new pages as needed. This is more convenient than manually managing
// offset/limit parameters.
//
// Example:
//
//	opts := &accounting.ListTagsOptions{
//	    Limit: int64Ptr(100),
//	}
//	iter := tagsService.ListIter(ctx, companyID, opts)
//	for iter.Next() {
//	    tag := iter.Value()
//	    fmt.Printf("Tag ID: %d, Name: %s\n", tag.Id, tag.Name)
//	}
//	if err := iter.Err(); err != nil {
//	    log.Fatal(err)
//	}
func (s *TagsService) ListIter(ctx context.Context, companyID int64, opts *ListTagsOptions) Iterator[gen.Tag] {
	// Determine page size (limit)
	limit := int64(50) // Default for tags API
	if opts != nil && opts.Limit != nil {
		limit = *opts.Limit
	}

	// Create a fetcher function that captures the service and options
	fetcher := func(ctx context.Context, offset, limit int64) ([]gen.Tag, int64, error) {
		// Create a copy of options with updated offset/limit
		fetchOpts := &ListTagsOptions{}
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

		// Tags API doesn't return total_count, so we use -1 to indicate unknown
		// The pager will continue fetching until an empty page is returned
		totalCount := int64(-1)
		if result.Count < int(limit) {
			// This is the last page
			totalCount = offset + int64(result.Count)
		}

		return result.Tags, totalCount, nil
	}

	return NewPager(ctx, fetcher, limit)
}
