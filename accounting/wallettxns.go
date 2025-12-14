package accounting

import (
	"context"
	"fmt"

	"github.com/muno/freee-api-go/internal/gen"
)

// ListWalletTxnsOptions contains optional parameters for listing wallet transactions.
type ListWalletTxnsOptions struct {
	// WalletableType filters by account type (口座区分)
	// Values: "bank_account" (銀行口座), "credit_card" (クレジットカード), "wallet" (現金)
	// Note: WalletableType and WalletableId must be specified together
	WalletableType *string

	// WalletableId filters by account ID (口座ID)
	// Note: WalletableType and WalletableId must be specified together
	WalletableId *int64

	// StartDate filters by transaction date start (取引日：開始日 yyyy-mm-dd)
	StartDate *string

	// EndDate filters by transaction date end (取引日：終了日 yyyy-mm-dd)
	EndDate *string

	// EntrySide filters by income/expense type (入金／出金)
	// Values: "income" (入金), "expense" (出金)
	EntrySide *string

	// Offset for pagination (デフォルト: 0)
	Offset *int64

	// Limit for pagination (デフォルト: 20, 最小: 1, 最大: 100)
	Limit *int64
}

// ListWalletTxnsResult contains the result of listing wallet transactions.
type ListWalletTxnsResult struct {
	// WalletTxns is the list of wallet transactions
	WalletTxns []gen.WalletTxn
}

// List retrieves a list of wallet transactions for the specified company.
//
// This method returns all wallet transactions matching the optional filter criteria.
// Use ListWalletTxnsOptions to filter by account, dates, entry side, etc.
//
// Example:
//
//	opts := &accounting.ListWalletTxnsOptions{
//	    WalletableType: stringPtr("bank_account"),
//	    WalletableId:   int64Ptr(12345),
//	    Limit:          int64Ptr(50),
//	    Offset:         int64Ptr(0),
//	}
//	result, err := walletTxnService.List(ctx, companyID, opts)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	for _, txn := range result.WalletTxns {
//	    fmt.Printf("Txn ID: %d, Amount: %d, Description: %s\n", txn.Id, txn.Amount, txn.Description)
//	}
func (s *WalletTxnService) List(ctx context.Context, companyID int64, opts *ListWalletTxnsOptions) (*ListWalletTxnsResult, error) {
	// Build parameters
	params := &gen.GetWalletTxnsParams{
		CompanyId: companyID,
	}

	if opts != nil {
		params.WalletableId = opts.WalletableId
		params.StartDate = opts.StartDate
		params.EndDate = opts.EndDate
		params.Offset = opts.Offset
		params.Limit = opts.Limit

		if opts.WalletableType != nil {
			walletableType := gen.GetWalletTxnsParamsWalletableType(*opts.WalletableType)
			params.WalletableType = &walletableType
		}
		if opts.EntrySide != nil {
			entrySide := gen.GetWalletTxnsParamsEntrySide(*opts.EntrySide)
			params.EntrySide = &entrySide
		}
	}

	// Call the generated client
	resp, err := s.genClient.GetWalletTxnsWithResponse(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to list wallet transactions: %w", err)
	}

	// Handle error responses
	if resp.JSON200 == nil {
		return nil, fmt.Errorf("unexpected response status: %s", resp.Status())
	}

	// Return the result
	return &ListWalletTxnsResult{
		WalletTxns: resp.JSON200.WalletTxns,
	}, nil
}

// Get retrieves a single wallet transaction by ID.
//
// Example:
//
//	txn, err := walletTxnService.Get(ctx, companyID, txnID)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("Wallet Transaction: %+v\n", txn)
func (s *WalletTxnService) Get(ctx context.Context, companyID int64, txnID int64) (*gen.WalletTxnResponse, error) {
	// Build parameters
	params := &gen.GetWalletTxnParams{
		CompanyId: companyID,
	}

	// Call the generated client
	resp, err := s.genClient.GetWalletTxnWithResponse(ctx, txnID, params)
	if err != nil {
		return nil, fmt.Errorf("failed to get wallet transaction: %w", err)
	}

	// Handle error responses
	if resp.JSON200 == nil {
		return nil, fmt.Errorf("unexpected response status: %s", resp.Status())
	}

	return resp.JSON200, nil
}

// Create creates a new wallet transaction.
//
// The params parameter should contain all required fields for creating a wallet transaction,
// including company ID, walletable ID, walletable type, date, amount, entry side, etc.
//
// Example:
//
//	params := gen.WalletTxnParams{
//	    CompanyId:      companyID,
//	    WalletableId:   12345,
//	    WalletableType: "bank_account",
//	    Date:           "2024-01-15",
//	    Amount:         10000,
//	    EntrySide:      "income",
//	    Description:    stringPtr("Payment received"),
//	}
//	txn, err := walletTxnService.Create(ctx, params)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("Created wallet transaction ID: %d\n", txn.WalletTxn.Id)
func (s *WalletTxnService) Create(ctx context.Context, params gen.WalletTxnParams) (*gen.WalletTxnResponse, error) {
	// Call the generated client
	resp, err := s.genClient.CreateWalletTxnWithResponse(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to create wallet transaction: %w", err)
	}

	// Handle error responses
	if resp.JSON201 == nil {
		return nil, fmt.Errorf("unexpected response status: %s", resp.Status())
	}

	return resp.JSON201, nil
}

// Delete deletes a wallet transaction by ID.
//
// Example:
//
//	err := walletTxnService.Delete(ctx, companyID, txnID)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println("Wallet transaction deleted successfully")
func (s *WalletTxnService) Delete(ctx context.Context, companyID int64, txnID int64) error {
	// Build parameters
	params := &gen.DestroyWalletTxnParams{
		CompanyId: companyID,
	}

	// Call the generated client
	resp, err := s.genClient.DestroyWalletTxnWithResponse(ctx, txnID, params)
	if err != nil {
		return fmt.Errorf("failed to delete wallet transaction: %w", err)
	}

	// Check for error responses
	if resp.StatusCode() >= 400 {
		return fmt.Errorf("failed to delete wallet transaction: %s", resp.Status())
	}

	return nil
}
