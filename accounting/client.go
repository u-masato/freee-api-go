// Package accounting provides a user-friendly Facade for the freee Accounting API.
//
// This package wraps the generated API client (internal/gen) and provides:
//   - Service-based organization (Deals, Journals, WalletTxns, Transfers)
//   - Context propagation
//   - Simplified error handling
//   - Iterator/Pager support for pagination (future phases)
//
// Example usage:
//
//	client := client.NewClient(
//	    client.WithTokenSource(tokenSource),
//	)
//	accountingClient := accounting.NewClient(client)
//
//	// Access service-specific operations
//	deals := accountingClient.Deals()
//	journals := accountingClient.Journals()
package accounting

import (
	"github.com/muno/freee-api-go/client"
	"github.com/muno/freee-api-go/internal/gen"
)

// Client is the main facade for the freee Accounting API.
//
// It provides access to service-specific clients (Deals, Journals, etc.)
// and manages the underlying generated API client.
//
// Example:
//
//	client := client.NewClient(
//	    client.WithTokenSource(tokenSource),
//	)
//	accountingClient := accounting.NewClient(client)
//
//	// Use service-specific clients
//	deals := accountingClient.Deals()
//	journals := accountingClient.Journals()
type Client struct {
	// client is the base freee API client
	client *client.Client

	// genClient is the generated OpenAPI client
	genClient *gen.Client

	// Service clients (lazy initialization)
	deals     *DealsService
	journals  *JournalsService
	walletTxn *WalletTxnService
	transfers *TransfersService
}

// NewClient creates a new accounting facade client.
//
// The provided client.Client should be configured with appropriate
// authentication (OAuth2 token source) and transport settings.
//
// Example:
//
//	baseClient := client.NewClient(
//	    client.WithTokenSource(tokenSource),
//	)
//	accountingClient := accounting.NewClient(baseClient)
func NewClient(c *client.Client) (*Client, error) {
	// Create the generated client with the base client's configuration
	genClient, err := gen.NewClient(
		c.BaseURL(),
		gen.WithHTTPClient(c.HTTPClient()),
	)
	if err != nil {
		return nil, err
	}

	return &Client{
		client:    c,
		genClient: genClient,
	}, nil
}

// Deals returns the DealsService for managing deals (取引).
//
// The service is lazily initialized on first access.
//
// Example:
//
//	deals := accountingClient.Deals()
//	list, err := deals.List(ctx, companyID, nil)
func (c *Client) Deals() *DealsService {
	if c.deals == nil {
		c.deals = &DealsService{
			client:    c.client,
			genClient: c.genClient,
		}
	}
	return c.deals
}

// Journals returns the JournalsService for managing journals (仕訳).
//
// The service is lazily initialized on first access.
//
// Example:
//
//	journals := accountingClient.Journals()
//	list, err := journals.List(ctx, companyID, nil)
func (c *Client) Journals() *JournalsService {
	if c.journals == nil {
		c.journals = &JournalsService{
			client:    c.client,
			genClient: c.genClient,
		}
	}
	return c.journals
}

// WalletTxns returns the WalletTxnService for managing wallet transactions (口座明細).
//
// The service is lazily initialized on first access.
//
// Example:
//
//	walletTxns := accountingClient.WalletTxns()
//	list, err := walletTxns.List(ctx, companyID, nil)
func (c *Client) WalletTxns() *WalletTxnService {
	if c.walletTxn == nil {
		c.walletTxn = &WalletTxnService{
			client:    c.client,
			genClient: c.genClient,
		}
	}
	return c.walletTxn
}

// Transfers returns the TransfersService for managing transfers (取引（振替）).
//
// The service is lazily initialized on first access.
//
// Example:
//
//	transfers := accountingClient.Transfers()
//	list, err := transfers.List(ctx, companyID, nil)
func (c *Client) Transfers() *TransfersService {
	if c.transfers == nil {
		c.transfers = &TransfersService{
			client:    c.client,
			genClient: c.genClient,
		}
	}
	return c.transfers
}

// BaseClient returns the underlying base client.
//
// This can be useful for advanced use cases where direct access
// to the base client is needed.
func (c *Client) BaseClient() *client.Client {
	return c.client
}

// GenClient returns the underlying generated API client.
//
// This is intended for advanced use cases or when the facade
// doesn't yet provide a specific operation. Use with caution
// as this exposes the internal generated API.
func (c *Client) GenClient() *gen.Client {
	return c.genClient
}
