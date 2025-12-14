package accounting

import (
	"github.com/muno/freee-api-go/client"
	"github.com/muno/freee-api-go/internal/gen"
)

// DealsService provides operations for managing deals (取引).
//
// Deals represent financial transactions in freee, such as income and expenses.
// This service wraps the generated API client and provides a user-friendly interface
// for CRUD operations on deals.
//
// All methods require a context.Context for cancellation and timeouts.
//
// Example:
//
//	deals := accountingClient.Deals()
//	// Future: list, err := deals.List(ctx, companyID, nil)
//	// Future: deal, err := deals.Get(ctx, companyID, dealID)
type DealsService struct {
	client    *client.Client
	genClient *gen.Client
}

// JournalsService provides operations for managing journals (仕訳).
//
// Journals represent accounting entries in the double-entry bookkeeping system.
// This includes both automatic journals (generated from deals) and manual journals.
//
// All methods require a context.Context for cancellation and timeouts.
//
// Example:
//
//	journals := accountingClient.Journals()
//	// Future: list, err := journals.List(ctx, companyID, nil)
//	// Future: manualJournals, err := journals.ListManual(ctx, companyID, nil)
type JournalsService struct {
	client    *client.Client
	genClient *gen.Client
}

// WalletTxnService provides operations for managing wallet transactions (口座明細).
//
// Wallet transactions represent entries in bank accounts, credit cards, and other
// financial accounts registered in freee. These can be used for automatic deal
// creation and reconciliation.
//
// All methods require a context.Context for cancellation and timeouts.
//
// Example:
//
//	walletTxns := accountingClient.WalletTxns()
//	// Future: list, err := walletTxns.List(ctx, companyID, walletableID, nil)
//	// Future: txn, err := walletTxns.Get(ctx, companyID, txnID)
type WalletTxnService struct {
	client    *client.Client
	genClient *gen.Client
}

// TransfersService provides operations for managing transfers (取引（振替）).
//
// Transfers represent movements of money between accounts, such as bank transfers
// or cash withdrawals. These are distinct from regular deals as they involve
// multiple accounts.
//
// All methods require a context.Context for cancellation and timeouts.
//
// Example:
//
//	transfers := accountingClient.Transfers()
//	// Future: list, err := transfers.List(ctx, companyID, nil)
//	// Future: transfer, err := transfers.Get(ctx, companyID, transferID)
type TransfersService struct {
	client    *client.Client
	genClient *gen.Client
}
