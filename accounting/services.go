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
	genClient *gen.ClientWithResponses
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
	genClient *gen.ClientWithResponses
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
	genClient *gen.ClientWithResponses
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
	genClient *gen.ClientWithResponses
}

// PartnersService provides operations for managing partners (取引先).
//
// Partners represent business partners such as clients, suppliers, and vendors
// that your company has transactions with.
//
// All methods require a context.Context for cancellation and timeouts.
//
// Example:
//
//	partners := accountingClient.Partners()
//	list, err := partners.List(ctx, companyID, nil)
//	partner, err := partners.Get(ctx, companyID, partnerID)
type PartnersService struct {
	client    *client.Client
	genClient *gen.ClientWithResponses
}

// AccountItemsService provides operations for managing account items (勘定科目).
//
// Account items represent accounts used in double-entry bookkeeping,
// such as cash, accounts receivable, sales, etc.
//
// All methods require a context.Context for cancellation and timeouts.
//
// Example:
//
//	accountItems := accountingClient.AccountItems()
//	list, err := accountItems.List(ctx, companyID, nil)
//	item, err := accountItems.Get(ctx, companyID, accountItemID)
type AccountItemsService struct {
	client    *client.Client
	genClient *gen.ClientWithResponses
}

// ItemsService provides operations for managing items (品目).
//
// Items represent product or service categories that can be associated
// with transactions.
//
// All methods require a context.Context for cancellation and timeouts.
//
// Example:
//
//	items := accountingClient.Items()
//	list, err := items.List(ctx, companyID, nil)
//	item, err := items.Get(ctx, companyID, itemID)
type ItemsService struct {
	client    *client.Client
	genClient *gen.ClientWithResponses
}

// SectionsService provides operations for managing sections (部門).
//
// Sections represent organizational units or departments within a company.
//
// All methods require a context.Context for cancellation and timeouts.
//
// Example:
//
//	sections := accountingClient.Sections()
//	list, err := sections.List(ctx, companyID, nil)
//	section, err := sections.Get(ctx, companyID, sectionID)
type SectionsService struct {
	client    *client.Client
	genClient *gen.ClientWithResponses
}

// TagsService provides operations for managing tags (メモタグ).
//
// Tags are labels that can be attached to transactions for additional
// categorization and filtering.
//
// All methods require a context.Context for cancellation and timeouts.
//
// Example:
//
//	tags := accountingClient.Tags()
//	list, err := tags.List(ctx, companyID, nil)
//	tag, err := tags.Get(ctx, companyID, tagID)
type TagsService struct {
	client    *client.Client
	genClient *gen.ClientWithResponses
}
