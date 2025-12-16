package accounting

import (
	"context"
	"fmt"

	"github.com/u-masato/freee-api-go/internal/gen"
)

// ListAccountItemsOptions contains optional parameters for listing account items.
type ListAccountItemsOptions struct {
	// BaseDate specifies the base date for tax code calculation (yyyy-mm-dd)
	BaseDate *string

	// StartUpdateDate filters by update date start (yyyy-mm-dd)
	StartUpdateDate *string

	// EndUpdateDate filters by update date end (yyyy-mm-dd)
	EndUpdateDate *string
}

// ListAccountItemsResult contains the result of listing account items.
type ListAccountItemsResult struct {
	// AccountItems is the response containing account items
	AccountItems *gen.AccountItemsResponse

	// Count is the number of account items returned in this response
	Count int
}

// List retrieves a list of account items for the specified company.
//
// This method returns all account items matching the optional filter criteria.
// Use ListAccountItemsOptions to filter by base date or update date range.
//
// Example:
//
//	opts := &accounting.ListAccountItemsOptions{
//	    BaseDate: stringPtr("2024-01-01"),
//	}
//	result, err := accountItemsService.List(ctx, companyID, opts)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	for _, item := range result.AccountItems.AccountItems {
//	    fmt.Printf("Account Item ID: %d, Name: %s\n", item.Id, item.Name)
//	}
func (s *AccountItemsService) List(ctx context.Context, companyID int64, opts *ListAccountItemsOptions) (*ListAccountItemsResult, error) {
	// Build parameters
	params := &gen.GetAccountItemsParams{
		CompanyId: companyID,
	}

	if opts != nil {
		params.BaseDate = opts.BaseDate
		params.StartUpdateDate = opts.StartUpdateDate
		params.EndUpdateDate = opts.EndUpdateDate
	}

	// Call the generated client
	resp, err := s.genClient.GetAccountItemsWithResponse(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to list account items: %w", err)
	}

	// Handle error responses
	if resp.JSON200 == nil {
		return nil, fmt.Errorf("unexpected response status: %s", resp.Status())
	}

	// Return the result
	return &ListAccountItemsResult{
		AccountItems: resp.JSON200,
		Count:        len(resp.JSON200.AccountItems),
	}, nil
}

// Get retrieves a single account item by ID.
//
// Example:
//
//	item, err := accountItemsService.Get(ctx, companyID, accountItemID)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("Account Item: %+v\n", item)
func (s *AccountItemsService) Get(ctx context.Context, companyID int64, accountItemID int64) (*gen.AccountItemResponse, error) {
	// Build parameters
	params := &gen.GetAccountItemParams{
		CompanyId: companyID,
	}

	// Call the generated client
	resp, err := s.genClient.GetAccountItemWithResponse(ctx, accountItemID, params)
	if err != nil {
		return nil, fmt.Errorf("failed to get account item: %w", err)
	}

	// Handle error responses
	if resp.JSON200 == nil {
		return nil, fmt.Errorf("unexpected response status: %s", resp.Status())
	}

	return resp.JSON200, nil
}

// Create creates a new account item.
//
// The params parameter should contain all required fields for creating an account item,
// including company ID and name.
//
// Example:
//
//	params := gen.AccountItemCreateParams{
//	    CompanyId: companyID,
//	    AccountItem: gen.AccountItemCreateParamsAccountItem{
//	        Name: "新規勘定科目",
//	        ...
//	    },
//	}
//	item, err := accountItemsService.Create(ctx, params)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("Created account item ID: %d\n", item.AccountItem.Id)
func (s *AccountItemsService) Create(ctx context.Context, params gen.AccountItemCreateParams) (*gen.AccountItemResponse, error) {
	// Call the generated client
	resp, err := s.genClient.CreateAccountItemWithResponse(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to create account item: %w", err)
	}

	// Handle error responses
	if resp.JSON201 == nil {
		return nil, fmt.Errorf("unexpected response status: %s", resp.Status())
	}

	return resp.JSON201, nil
}

// Update updates an existing account item.
//
// The params parameter should contain the fields to update.
//
// Example:
//
//	params := gen.AccountItemUpdateParams{
//	    CompanyId: companyID,
//	    AccountItem: gen.AccountItemUpdateParamsAccountItem{
//	        Name: stringPtr("更新後勘定科目"),
//	        ...
//	    },
//	}
//	item, err := accountItemsService.Update(ctx, accountItemID, params)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("Updated account item ID: %d\n", item.AccountItem.Id)
func (s *AccountItemsService) Update(ctx context.Context, accountItemID int64, params gen.AccountItemUpdateParams) (*gen.AccountItemResponse, error) {
	// Call the generated client
	resp, err := s.genClient.UpdateAccountItemWithResponse(ctx, accountItemID, params)
	if err != nil {
		return nil, fmt.Errorf("failed to update account item: %w", err)
	}

	// Handle error responses
	if resp.JSON200 == nil {
		return nil, fmt.Errorf("unexpected response status: %s", resp.Status())
	}

	return resp.JSON200, nil
}

// Delete deletes an account item by ID.
//
// Example:
//
//	err := accountItemsService.Delete(ctx, companyID, accountItemID)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println("Account item deleted successfully")
func (s *AccountItemsService) Delete(ctx context.Context, companyID int64, accountItemID int64) error {
	// Build parameters
	params := &gen.DestroyAccountItemParams{
		CompanyId: companyID,
	}

	// Call the generated client
	resp, err := s.genClient.DestroyAccountItemWithResponse(ctx, accountItemID, params)
	if err != nil {
		return fmt.Errorf("failed to delete account item: %w", err)
	}

	// Check for error responses
	if resp.StatusCode() >= 400 {
		return fmt.Errorf("failed to delete account item: %s", resp.Status())
	}

	return nil
}
