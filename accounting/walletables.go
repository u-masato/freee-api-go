package accounting

import (
	"context"
	"fmt"

	"github.com/u-masato/freee-api-go/internal/gen"
)

// ListWalletablesOptions contains optional parameters for listing walletables.
type ListWalletablesOptions struct {
	// Type filters by account type (口座種別)
	// Values: "bank_account" (銀行口座), "credit_card" (クレジットカード), "wallet" (その他の決済口座)
	Type *string

	// WithBalance includes walletable balance information
	WithBalance *bool

	// WithLastSyncedAt includes last synced timestamp
	WithLastSyncedAt *bool

	// WithSyncStatus includes sync status
	WithSyncStatus *bool

	// StartUpdateDate filters by update date start (yyyy-mm-dd)
	StartUpdateDate *string

	// EndUpdateDate filters by update date end (yyyy-mm-dd)
	EndUpdateDate *string
}

// ListWalletablesResult contains the result of listing walletables.
type ListWalletablesResult struct {
	// Walletables is the list of walletables
	Walletables []gen.Walletable

	// Count is the number of walletables returned in this response
	Count int

	// UpToDate indicates whether the aggregation is up to date
	UpToDate *bool
}

// GetWalletableOptions contains optional parameters for retrieving a walletable.
type GetWalletableOptions struct {
	// WithLastSyncedAt includes last synced timestamp
	WithLastSyncedAt *bool

	// WithSyncStatus includes sync status
	WithSyncStatus *bool
}

// GetWalletableResult contains the result of retrieving a walletable.
type GetWalletableResult struct {
	// Walletable is the walletable details
	Walletable gen.Walletable

	// UpToDate indicates whether the aggregation is up to date
	UpToDate *bool
}

// List retrieves a list of walletables for the specified company.
//
// Example:
//
//	opts := &accounting.ListWalletablesOptions{
//	    Type:        stringPtr("bank_account"),
//	    WithBalance: boolPtr(true),
//	}
//	result, err := walletablesService.List(ctx, companyID, opts)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	for _, walletable := range result.Walletables {
//	    fmt.Printf("Walletable ID: %d, Name: %s\n", walletable.Id, walletable.Name)
//	}
func (s *WalletablesService) List(ctx context.Context, companyID int64, opts *ListWalletablesOptions) (*ListWalletablesResult, error) {
	params := &gen.GetWalletablesParams{
		CompanyId: companyID,
	}

	if opts != nil {
		params.WithBalance = opts.WithBalance
		params.WithLastSyncedAt = opts.WithLastSyncedAt
		params.WithSyncStatus = opts.WithSyncStatus
		params.StartUpdateDate = opts.StartUpdateDate
		params.EndUpdateDate = opts.EndUpdateDate

		if opts.Type != nil {
			walletableType := gen.GetWalletablesParamsType(*opts.Type)
			params.Type = &walletableType
		}
	}

	resp, err := s.genClient.GetWalletablesWithResponse(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to list walletables: %w", err)
	}

	if resp.JSON200 == nil {
		return nil, fmt.Errorf("unexpected response status: %s", resp.Status())
	}

	var upToDate *bool
	if resp.JSON200.Meta != nil {
		upToDate = resp.JSON200.Meta.UpToDate
	}

	return &ListWalletablesResult{
		Walletables: resp.JSON200.Walletables,
		Count:       len(resp.JSON200.Walletables),
		UpToDate:    upToDate,
	}, nil
}

// Get retrieves a single walletable by type and ID.
//
// Example:
//
//	opts := &accounting.GetWalletableOptions{
//	    WithSyncStatus: boolPtr(true),
//	}
//	result, err := walletablesService.Get(ctx, companyID, "bank_account", walletableID, opts)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("Walletable ID: %d, Name: %s\n", result.Walletable.Id, result.Walletable.Name)
func (s *WalletablesService) Get(ctx context.Context, companyID int64, walletableType string, walletableID int64, opts *GetWalletableOptions) (*GetWalletableResult, error) {
	params := &gen.GetWalletableParams{
		CompanyId: companyID,
	}

	if opts != nil {
		params.WithLastSyncedAt = opts.WithLastSyncedAt
		params.WithSyncStatus = opts.WithSyncStatus
	}

	pType := gen.GetWalletableParamsType(walletableType)
	resp, err := s.genClient.GetWalletableWithResponse(ctx, pType, walletableID, params)
	if err != nil {
		return nil, fmt.Errorf("failed to get walletable: %w", err)
	}

	if resp.JSON200 == nil {
		return nil, fmt.Errorf("unexpected response status: %s", resp.Status())
	}

	var upToDate *bool
	if resp.JSON200.Meta != nil {
		upToDate = resp.JSON200.Meta.UpToDate
	}

	return &GetWalletableResult{
		Walletable: resp.JSON200.Walletable,
		UpToDate:   upToDate,
	}, nil
}

// Create creates a new walletable.
//
// Example:
//
//	params := gen.WalletableCreateParams{
//	    CompanyId: companyID,
//	    Name:      "Main Wallet",
//	    Type:      "wallet",
//	}
//	walletable, err := walletablesService.Create(ctx, params)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("Created walletable ID: %d\n", walletable.Walletable.Id)
func (s *WalletablesService) Create(ctx context.Context, params gen.WalletableCreateParams) (*gen.WalletableCreateResponse, error) {
	resp, err := s.genClient.CreateWalletableWithResponse(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to create walletable: %w", err)
	}

	if resp.JSON201 == nil {
		return nil, fmt.Errorf("unexpected response status: %s", resp.Status())
	}

	return resp.JSON201, nil
}

// Update updates an existing walletable.
//
// Example:
//
//	params := gen.WalletableUpdateParams{
//	    CompanyId: companyID,
//	    Name:      "Updated Wallet",
//	}
//	walletable, err := walletablesService.Update(ctx, "wallet", walletableID, params)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("Updated walletable ID: %d\n", walletable.Id)
func (s *WalletablesService) Update(ctx context.Context, walletableType string, walletableID int64, params gen.WalletableUpdateParams) (*gen.WalletableUpdateResponse, error) {
	pType := gen.UpdateWalletableParamsType(walletableType)
	resp, err := s.genClient.UpdateWalletableWithResponse(ctx, pType, walletableID, params)
	if err != nil {
		return nil, fmt.Errorf("failed to update walletable: %w", err)
	}

	if resp.JSON200 == nil {
		return nil, fmt.Errorf("unexpected response status: %s", resp.Status())
	}

	walletable := resp.JSON200.Walletable
	return &walletable, nil
}

// Delete deletes a walletable by type and ID.
//
// Example:
//
//	err := walletablesService.Delete(ctx, companyID, "wallet", walletableID)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println("Walletable deleted successfully")
func (s *WalletablesService) Delete(ctx context.Context, companyID int64, walletableType string, walletableID int64) error {
	params := &gen.DestroyWalletableParams{
		CompanyId: companyID,
	}

	pType := gen.DestroyWalletableParamsType(walletableType)
	resp, err := s.genClient.DestroyWalletableWithResponse(ctx, pType, walletableID, params)
	if err != nil {
		return fmt.Errorf("failed to delete walletable: %w", err)
	}

	if resp.StatusCode() >= 400 {
		return fmt.Errorf("failed to delete walletable: %s", resp.Status())
	}

	return nil
}
