package accounting

import (
	"context"
	"fmt"

	"github.com/u-masato/freee-api-go/internal/gen"
)

// ListTransfersOptions contains optional parameters for listing transfers.
type ListTransfersOptions struct {
	// StartDate filters by transfer date start (振替日：開始日 yyyy-mm-dd)
	StartDate *string

	// EndDate filters by transfer date end (振替日：終了日 yyyy-mm-dd)
	EndDate *string

	// Offset for pagination (デフォルト: 0)
	Offset *int64

	// Limit for pagination (デフォルト: 20, 最小: 1, 最大: 100)
	Limit *int64
}

// ListTransfersResult contains the result of listing transfers.
type ListTransfersResult struct {
	// Transfers is the list of transfers
	Transfers []gen.Transfer
}

// List retrieves a list of transfers for the specified company.
//
// This method returns all transfers matching the optional filter criteria.
// Use ListTransfersOptions to filter by date range and control pagination.
//
// Example:
//
//	opts := &accounting.ListTransfersOptions{
//	    StartDate: stringPtr("2024-01-01"),
//	    EndDate:   stringPtr("2024-01-31"),
//	    Limit:     int64Ptr(50),
//	    Offset:    int64Ptr(0),
//	}
//	result, err := transfersService.List(ctx, companyID, opts)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	for _, transfer := range result.Transfers {
//	    fmt.Printf("Transfer ID: %d, Amount: %d\n", transfer.Id, transfer.Amount)
//	}
func (s *TransfersService) List(ctx context.Context, companyID int64, opts *ListTransfersOptions) (*ListTransfersResult, error) {
	// Build parameters
	params := &gen.GetTransfersParams{
		CompanyId: companyID,
	}

	if opts != nil {
		params.StartDate = opts.StartDate
		params.EndDate = opts.EndDate
		params.Offset = opts.Offset
		params.Limit = opts.Limit
	}

	// Call the generated client
	resp, err := s.genClient.GetTransfersWithResponse(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to list transfers: %w", err)
	}

	// Handle error responses
	if resp.JSON200 == nil {
		return nil, fmt.Errorf("unexpected response status: %s", resp.Status())
	}

	// Return the result
	return &ListTransfersResult{
		Transfers: resp.JSON200.Transfers,
	}, nil
}

// ListIter returns an iterator for paginated transfer results.
//
// The iterator transparently handles pagination, automatically fetching
// new pages as needed. This is more convenient than manually managing
// offset/limit parameters.
//
// Note: The freee API does not provide total_count for transfers,
// so the iterator will fetch pages until an empty result is returned.
//
// Example:
//
//	startDate := "2024-01-01"
//	opts := &accounting.ListTransfersOptions{
//	    StartDate: &startDate,
//	}
//	iter := transfersService.ListIter(ctx, companyID, opts)
//	for iter.Next() {
//	    transfer := iter.Value()
//	    fmt.Printf("Transfer ID: %d, Amount: %d\n", transfer.Id, transfer.Amount)
//	}
//	if err := iter.Err(); err != nil {
//	    log.Fatal(err)
//	}
func (s *TransfersService) ListIter(ctx context.Context, companyID int64, opts *ListTransfersOptions) Iterator[gen.Transfer] {
	// Determine page size (limit)
	limit := int64(20) // Default
	if opts != nil && opts.Limit != nil {
		limit = *opts.Limit
	}

	// Create a fetcher function that captures the service and options
	fetcher := func(ctx context.Context, offset, limit int64) ([]gen.Transfer, int64, error) {
		// Create a copy of options with updated offset/limit
		fetchOpts := &ListTransfersOptions{}
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

		// Since the API doesn't provide total_count, we return 0
		// The pager will continue until an empty array is returned
		return result.Transfers, 0, nil
	}

	return NewPager(ctx, fetcher, limit)
}

// Get retrieves a single transfer by ID.
//
// Example:
//
//	transfer, err := transfersService.Get(ctx, companyID, transferID)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("Transfer: %+v\n", transfer)
func (s *TransfersService) Get(ctx context.Context, companyID int64, transferID int64) (*gen.TransferResponse, error) {
	// Build parameters
	params := &gen.GetTransferParams{
		CompanyId: companyID,
	}

	// Call the generated client
	resp, err := s.genClient.GetTransferWithResponse(ctx, transferID, params)
	if err != nil {
		return nil, fmt.Errorf("failed to get transfer: %w", err)
	}

	// Handle error responses
	if resp.JSON200 == nil {
		return nil, fmt.Errorf("unexpected response status: %s", resp.Status())
	}

	return resp.JSON200, nil
}

// Create creates a new transfer.
//
// The params parameter should contain all required fields for creating a transfer,
// including company ID, date, amount, and wallet information.
//
// Example:
//
//	params := gen.TransferParams{
//	    CompanyId:          companyID,
//	    Date:               "2024-01-15",
//	    Amount:             10000,
//	    FromWalletableId:   123,
//	    FromWalletableType: "bank_account",
//	    ToWalletableId:     456,
//	    ToWalletableType:   "wallet",
//	    Description:        stringPtr("Transfer to cash"),
//	}
//	transfer, err := transfersService.Create(ctx, params)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("Created transfer ID: %d\n", transfer.Transfer.Id)
func (s *TransfersService) Create(ctx context.Context, params gen.TransferParams) (*gen.TransferResponse, error) {
	// Call the generated client
	resp, err := s.genClient.CreateTransferWithResponse(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to create transfer: %w", err)
	}

	// Handle error responses
	if resp.JSON201 == nil {
		return nil, fmt.Errorf("unexpected response status: %s", resp.Status())
	}

	return resp.JSON201, nil
}

// Update updates an existing transfer.
//
// The params parameter should contain the fields to update.
// All fields in TransferParams are required even for updates.
//
// Example:
//
//	params := gen.TransferParams{
//	    CompanyId:          companyID,
//	    Date:               "2024-01-20",
//	    Amount:             15000,
//	    FromWalletableId:   123,
//	    FromWalletableType: "bank_account",
//	    ToWalletableId:     456,
//	    ToWalletableType:   "wallet",
//	    Description:        stringPtr("Updated transfer"),
//	}
//	transfer, err := transfersService.Update(ctx, transferID, params)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("Updated transfer ID: %d\n", transfer.Transfer.Id)
func (s *TransfersService) Update(ctx context.Context, transferID int64, params gen.TransferParams) (*gen.TransferResponse, error) {
	// Call the generated client
	resp, err := s.genClient.UpdateTransferWithResponse(ctx, transferID, params)
	if err != nil {
		return nil, fmt.Errorf("failed to update transfer: %w", err)
	}

	// Handle error responses
	if resp.JSON200 == nil {
		return nil, fmt.Errorf("unexpected response status: %s", resp.Status())
	}

	return resp.JSON200, nil
}

// Delete deletes a transfer by ID.
//
// Example:
//
//	err := transfersService.Delete(ctx, companyID, transferID)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println("Transfer deleted successfully")
func (s *TransfersService) Delete(ctx context.Context, companyID int64, transferID int64) error {
	// Build parameters
	params := &gen.DestroyTransferParams{
		CompanyId: companyID,
	}

	// Call the generated client
	resp, err := s.genClient.DestroyTransferWithResponse(ctx, transferID, params)
	if err != nil {
		return fmt.Errorf("failed to delete transfer: %w", err)
	}

	// Check for error responses
	if resp.StatusCode() >= 400 {
		return fmt.Errorf("failed to delete transfer: %s", resp.Status())
	}

	return nil
}
