package accounting

import (
	"context"
	"fmt"

	"github.com/muno/freee-api-go/internal/gen"
)

// ListDealsOptions contains optional parameters for listing deals.
type ListDealsOptions struct {
	// PartnerId filters by partner ID (取引先ID)
	PartnerId *int64

	// AccountItemId filters by account item ID (勘定科目ID)
	AccountItemId *int64

	// PartnerCode filters by partner code (取引先コード)
	PartnerCode *string

	// Status filters by settlement status (決済状況)
	// Values: "unsettled" (未決済), "settled" (完了)
	Status *string

	// Type filters by income/expense type (収支区分)
	// Values: "income" (収入), "expense" (支出)
	Type *string

	// StartIssueDate filters by issue date start (発生日：開始日 yyyy-mm-dd)
	StartIssueDate *string

	// EndIssueDate filters by issue date end (発生日：終了日 yyyy-mm-dd)
	EndIssueDate *string

	// StartDueDate filters by due date start (支払期日：開始日 yyyy-mm-dd)
	StartDueDate *string

	// EndDueDate filters by due date end (支払期日：終了日 yyyy-mm-dd)
	EndDueDate *string

	// StartRenewDate filters by renew date start (更新日：開始日 yyyy-mm-dd)
	StartRenewDate *string

	// EndRenewDate filters by renew date end (更新日：終了日 yyyy-mm-dd)
	EndRenewDate *string

	// Offset for pagination (デフォルト: 0)
	Offset *int64

	// Limit for pagination (デフォルト: 20, 最大: 100)
	Limit *int64

	// Accruals controls display of accrual lines (債権債務行の表示)
	// Values: "without" (表示しない), "with" (表示する)
	Accruals *string
}

// GetDealOptions contains optional parameters for getting a deal.
type GetDealOptions struct {
	// Accruals controls display of accrual lines (債権債務行の表示)
	// Values: "without" (表示しない), "with" (表示する)
	Accruals *string
}

// ListDealsResult contains the result of listing deals.
type ListDealsResult struct {
	// Deals is the list of deals
	Deals []gen.Deal

	// TotalCount is the total number of deals matching the query
	TotalCount int64
}

// List retrieves a list of deals for the specified company.
//
// This method returns all deals matching the optional filter criteria.
// Use ListDealsOptions to filter by partner, account item, dates, etc.
//
// Example:
//
//	opts := &accounting.ListDealsOptions{
//	    Type:   stringPtr("expense"),
//	    Limit:  int64Ptr(50),
//	    Offset: int64Ptr(0),
//	}
//	result, err := dealsService.List(ctx, companyID, opts)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	for _, deal := range result.Deals {
//	    fmt.Printf("Deal ID: %d, Amount: %d\n", deal.Id, deal.Amount)
//	}
func (s *DealsService) List(ctx context.Context, companyID int64, opts *ListDealsOptions) (*ListDealsResult, error) {
	// Build parameters
	params := &gen.GetDealsParams{
		CompanyId: companyID,
	}

	if opts != nil {
		params.PartnerId = opts.PartnerId
		params.AccountItemId = opts.AccountItemId
		params.PartnerCode = opts.PartnerCode
		params.Offset = opts.Offset
		params.Limit = opts.Limit

		if opts.Status != nil {
			status := gen.GetDealsParamsStatus(*opts.Status)
			params.Status = &status
		}
		if opts.Type != nil {
			dealType := gen.GetDealsParamsType(*opts.Type)
			params.Type = &dealType
		}
		if opts.Accruals != nil {
			accruals := gen.GetDealsParamsAccruals(*opts.Accruals)
			params.Accruals = &accruals
		}

		params.StartIssueDate = opts.StartIssueDate
		params.EndIssueDate = opts.EndIssueDate
		params.StartDueDate = opts.StartDueDate
		params.EndDueDate = opts.EndDueDate
		params.StartRenewDate = opts.StartRenewDate
		params.EndRenewDate = opts.EndRenewDate
	}

	// Call the generated client
	resp, err := s.genClient.GetDealsWithResponse(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to list deals: %w", err)
	}

	// Handle error responses
	if resp.JSON200 == nil {
		return nil, fmt.Errorf("unexpected response status: %s", resp.Status())
	}

	// Return the result
	return &ListDealsResult{
		Deals:      resp.JSON200.Deals,
		TotalCount: resp.JSON200.Meta.TotalCount,
	}, nil
}

// Get retrieves a single deal by ID.
//
// Example:
//
//	opts := &accounting.GetDealOptions{
//	    Accruals: stringPtr("with"),
//	}
//	deal, err := dealsService.Get(ctx, companyID, dealID, opts)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("Deal: %+v\n", deal)
func (s *DealsService) Get(ctx context.Context, companyID int64, dealID int64, opts *GetDealOptions) (*gen.DealResponse, error) {
	// Build parameters
	params := &gen.GetDealParams{
		CompanyId: companyID,
	}

	if opts != nil && opts.Accruals != nil {
		accruals := gen.GetDealParamsAccruals(*opts.Accruals)
		params.Accruals = &accruals
	}

	// Call the generated client
	resp, err := s.genClient.GetDealWithResponse(ctx, dealID, params)
	if err != nil {
		return nil, fmt.Errorf("failed to get deal: %w", err)
	}

	// Handle error responses
	if resp.JSON200 == nil {
		return nil, fmt.Errorf("unexpected response status: %s", resp.Status())
	}

	return resp.JSON200, nil
}

// Create creates a new deal.
//
// The params parameter should contain all required fields for creating a deal,
// including company ID, issue date, type, and details.
//
// Example:
//
//	params := gen.DealCreateParams{
//	    CompanyId: companyID,
//	    IssueDate: "2024-01-15",
//	    Type:      "expense",
//	    Details: []struct{...}{
//	        {
//	            AccountItemId: 12345,
//	            TaxCode:       108,
//	            Amount:        10000,
//	        },
//	    },
//	}
//	deal, err := dealsService.Create(ctx, params)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("Created deal ID: %d\n", deal.Id)
func (s *DealsService) Create(ctx context.Context, params gen.DealCreateParams) (*gen.DealCreateResponse, error) {
	// Call the generated client
	resp, err := s.genClient.CreateDealWithResponse(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to create deal: %w", err)
	}

	// Handle error responses
	if resp.JSON201 == nil {
		return nil, fmt.Errorf("unexpected response status: %s", resp.Status())
	}

	return resp.JSON201, nil
}

// Update updates an existing deal.
//
// The params parameter should contain the fields to update.
// Only specified fields will be updated.
//
// Example:
//
//	params := gen.DealUpdateParams{
//	    CompanyId: companyID,
//	    IssueDate: "2024-01-20",
//	    Type:      "expense",
//	    Details: []struct{...}{
//	        {
//	            AccountItemId: 12345,
//	            TaxCode:       108,
//	            Amount:        15000,
//	        },
//	    },
//	}
//	deal, err := dealsService.Update(ctx, dealID, params)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("Updated deal ID: %d\n", deal.Deal.Id)
func (s *DealsService) Update(ctx context.Context, dealID int64, params gen.DealUpdateParams) (*gen.DealResponse, error) {
	// Call the generated client
	resp, err := s.genClient.UpdateDealWithResponse(ctx, dealID, params)
	if err != nil {
		return nil, fmt.Errorf("failed to update deal: %w", err)
	}

	// Handle error responses
	if resp.JSON200 == nil {
		return nil, fmt.Errorf("unexpected response status: %s", resp.Status())
	}

	return resp.JSON200, nil
}

// Delete deletes a deal by ID.
//
// Example:
//
//	err := dealsService.Delete(ctx, companyID, dealID)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println("Deal deleted successfully")
func (s *DealsService) Delete(ctx context.Context, companyID int64, dealID int64) error {
	// Build parameters
	params := &gen.DestroyDealParams{
		CompanyId: companyID,
	}

	// Call the generated client
	resp, err := s.genClient.DestroyDealWithResponse(ctx, dealID, params)
	if err != nil {
		return fmt.Errorf("failed to delete deal: %w", err)
	}

	// Check for error responses
	if resp.StatusCode() >= 400 {
		return fmt.Errorf("failed to delete deal: %s", resp.Status())
	}

	return nil
}

// ListIter returns an iterator for paginated deal results.
//
// The iterator transparently handles pagination, automatically fetching
// new pages as needed. This is more convenient than manually managing
// offset/limit parameters.
//
// Example:
//
//	typ := "expense"
//	opts := &accounting.ListDealsOptions{
//	    Type: &typ,
//	}
//	iter := dealsService.ListIter(ctx, companyID, opts)
//	for iter.Next() {
//	    deal := iter.Value()
//	    fmt.Printf("Deal ID: %d, Amount: %d\n", deal.Id, deal.Amount)
//	}
//	if err := iter.Err(); err != nil {
//	    log.Fatal(err)
//	}
func (s *DealsService) ListIter(ctx context.Context, companyID int64, opts *ListDealsOptions) Iterator[gen.Deal] {
	// Determine page size (limit)
	limit := int64(20) // Default
	if opts != nil && opts.Limit != nil {
		limit = *opts.Limit
	}

	// Create a fetcher function that captures the service and options
	fetcher := func(ctx context.Context, offset, limit int64) ([]gen.Deal, int64, error) {
		// Create a copy of options with updated offset/limit
		fetchOpts := &ListDealsOptions{}
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

		return result.Deals, result.TotalCount, nil
	}

	return NewPager(ctx, fetcher, limit)
}
