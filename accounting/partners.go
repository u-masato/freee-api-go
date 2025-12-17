package accounting

import (
	"context"
	"fmt"

	"github.com/u-masato/freee-api-go/internal/gen"
)

// ListPartnersOptions contains optional parameters for listing partners.
type ListPartnersOptions struct {
	// StartUpdateDate filters by update date start (yyyy-mm-dd)
	StartUpdateDate *string

	// EndUpdateDate filters by update date end (yyyy-mm-dd)
	EndUpdateDate *string

	// Offset for pagination (default: 0)
	Offset *int64

	// Limit for pagination (default: 50, min: 1, max: 3000)
	Limit *int64

	// Keyword searches partner code, name, long name, name kana, shortcut1/2
	// Multiple keywords separated by space/tab perform AND search
	Keyword *string
}

// ListPartnersResult contains the result of listing partners.
type ListPartnersResult struct {
	// Partners is the list of partners
	Partners *gen.PartnersResponse

	// Count is the number of partners returned in this response
	Count int
}

// List retrieves a list of partners for the specified company.
//
// This method returns all partners matching the optional filter criteria.
// Use ListPartnersOptions to filter by update date range or keyword.
//
// Example:
//
//	opts := &accounting.ListPartnersOptions{
//	    Keyword: stringPtr("株式会社"),
//	    Limit:   int64Ptr(100),
//	}
//	result, err := partnersService.List(ctx, companyID, opts)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	for _, partner := range result.Partners {
//	    fmt.Printf("Partner ID: %d, Name: %s\n", partner.Id, partner.Name)
//	}
func (s *PartnersService) List(ctx context.Context, companyID int64, opts *ListPartnersOptions) (*ListPartnersResult, error) {
	// Build parameters
	params := &gen.GetPartnersParams{
		CompanyId: companyID,
	}

	if opts != nil {
		params.StartUpdateDate = opts.StartUpdateDate
		params.EndUpdateDate = opts.EndUpdateDate
		params.Offset = opts.Offset
		params.Limit = opts.Limit
		params.Keyword = opts.Keyword
	}

	// Call the generated client
	resp, err := s.genClient.GetPartnersWithResponse(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to list partners: %w", err)
	}

	// Handle error responses
	if resp.JSON200 == nil {
		return nil, fmt.Errorf("unexpected response status: %s", resp.Status())
	}

	// Return the result
	return &ListPartnersResult{
		Partners: resp.JSON200,
		Count:    len(resp.JSON200.Partners),
	}, nil
}

// Get retrieves a single partner by ID.
//
// Example:
//
//	partner, err := partnersService.Get(ctx, companyID, partnerID)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("Partner: %+v\n", partner)
func (s *PartnersService) Get(ctx context.Context, companyID int64, partnerID int64) (*gen.PartnerResponse, error) {
	// Build parameters
	params := &gen.GetPartnerParams{
		CompanyId: companyID,
	}

	// Call the generated client
	resp, err := s.genClient.GetPartnerWithResponse(ctx, partnerID, params)
	if err != nil {
		return nil, fmt.Errorf("failed to get partner: %w", err)
	}

	// Handle error responses
	if resp.JSON200 == nil {
		return nil, fmt.Errorf("unexpected response status: %s", resp.Status())
	}

	return resp.JSON200, nil
}

// Create creates a new partner.
//
// The params parameter should contain all required fields for creating a partner,
// including company ID and name.
//
// Example:
//
//	params := gen.PartnerCreateParams{
//	    CompanyId: companyID,
//	    Name:      "株式会社テスト",
//	}
//	partner, err := partnersService.Create(ctx, params)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("Created partner ID: %d\n", partner.Partner.Id)
func (s *PartnersService) Create(ctx context.Context, params gen.PartnerCreateParams) (*gen.PartnerResponse, error) {
	// Call the generated client
	resp, err := s.genClient.CreatePartnerWithResponse(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to create partner: %w", err)
	}

	// Handle error responses
	if resp.JSON201 == nil {
		return nil, fmt.Errorf("unexpected response status: %s", resp.Status())
	}

	return resp.JSON201, nil
}

// Update updates an existing partner.
//
// The params parameter should contain the fields to update.
// Only specified fields will be updated.
//
// Example:
//
//	params := gen.PartnerUpdateParams{
//	    CompanyId: companyID,
//	    Name:      stringPtr("株式会社テスト更新"),
//	}
//	partner, err := partnersService.Update(ctx, partnerID, params)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("Updated partner ID: %d\n", partner.Partner.Id)
func (s *PartnersService) Update(ctx context.Context, partnerID int64, params gen.PartnerUpdateParams) (*gen.PartnerResponse, error) {
	// Call the generated client
	resp, err := s.genClient.UpdatePartnerWithResponse(ctx, partnerID, params)
	if err != nil {
		return nil, fmt.Errorf("failed to update partner: %w", err)
	}

	// Handle error responses
	if resp.JSON200 == nil {
		return nil, fmt.Errorf("unexpected response status: %s", resp.Status())
	}

	return resp.JSON200, nil
}

// Delete deletes a partner by ID.
//
// Example:
//
//	err := partnersService.Delete(ctx, companyID, partnerID)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println("Partner deleted successfully")
func (s *PartnersService) Delete(ctx context.Context, companyID int64, partnerID int64) error {
	// Build parameters
	params := &gen.DestroyPartnerParams{
		CompanyId: companyID,
	}

	// Call the generated client
	resp, err := s.genClient.DestroyPartnerWithResponse(ctx, partnerID, params)
	if err != nil {
		return fmt.Errorf("failed to delete partner: %w", err)
	}

	// Check for error responses
	if resp.StatusCode() >= 400 {
		return fmt.Errorf("failed to delete partner: %s", resp.Status())
	}

	return nil
}

// ListIter returns an iterator for paginated partner results.
//
// The iterator transparently handles pagination, automatically fetching
// new pages as needed. This is more convenient than manually managing
// offset/limit parameters.
//
// Note: The Partners API does not return total_count, so the iterator
// will fetch pages until an empty page is returned.
//
// Example:
//
//	keyword := "株式会社"
//	opts := &accounting.ListPartnersOptions{
//	    Keyword: &keyword,
//	}
//	iter := partnersService.ListIter(ctx, companyID, opts)
//	for iter.Next() {
//	    partner := iter.Value()
//	    fmt.Printf("Partner ID: %d, Name: %s\n", partner.Id, partner.Name)
//	}
//	if err := iter.Err(); err != nil {
//	    log.Fatal(err)
//	}
//
// PartnerListItem is the type for individual partner items in list responses.
// This is a type alias for the inline struct used in PartnersResponse.
type PartnerListItem = struct {
	AddressAttributes *struct {
		PrefectureCode *int64  `json:"prefecture_code"`
		StreetName1    *string `json:"street_name1"`
		StreetName2    *string `json:"street_name2"`
		Zipcode        *string `json:"zipcode"`
	} `json:"address_attributes,omitempty"`
	Available                    bool    `json:"available"`
	Code                         *string `json:"code"`
	CompanyId                    int64   `json:"company_id"`
	ContactName                  *string `json:"contact_name"`
	CountryCode                  *string `json:"country_code,omitempty"`
	DefaultTitle                 *string `json:"default_title"`
	Email                        *string `json:"email"`
	Id                           int64   `json:"id"`
	InvoiceRegistrationNumber    *string `json:"invoice_registration_number"`
	LongName                     *string `json:"long_name"`
	Name                         string  `json:"name"`
	NameKana                     *string `json:"name_kana"`
	OrgCode                      *int64  `json:"org_code"`
	PartnerBankAccountAttributes *struct {
		AccountName     *string                                                              `json:"account_name"`
		AccountNumber   *string                                                              `json:"account_number"`
		AccountType     *gen.PartnersResponsePartnersPartnerBankAccountAttributesAccountType `json:"account_type"`
		BankCode        *string                                                              `json:"bank_code"`
		BankName        *string                                                              `json:"bank_name"`
		BankNameKana    *string                                                              `json:"bank_name_kana"`
		BranchCode      *string                                                              `json:"branch_code"`
		BranchKana      *string                                                              `json:"branch_kana"`
		BranchName      *string                                                              `json:"branch_name"`
		LongAccountName *string                                                              `json:"long_account_name"`
	} `json:"partner_bank_account_attributes,omitempty"`
	PartnerDocSettingAttributes *struct {
		SendingMethod *gen.PartnersResponsePartnersPartnerDocSettingAttributesSendingMethod `json:"sending_method"`
	} `json:"partner_doc_setting_attributes,omitempty"`
	PayerWalletableId       *int64                                               `json:"payer_walletable_id"`
	Phone                   *string                                              `json:"phone"`
	QualifiedInvoiceIssuer  *bool                                                `json:"qualified_invoice_issuer,omitempty"`
	Shortcut1               *string                                              `json:"shortcut1"`
	Shortcut2               *string                                              `json:"shortcut2"`
	TransferFeeHandlingSide *gen.PartnersResponsePartnersTransferFeeHandlingSide `json:"transfer_fee_handling_side,omitempty"`
	UpdateDate              string                                               `json:"update_date"`
}

func (s *PartnersService) ListIter(ctx context.Context, companyID int64, opts *ListPartnersOptions) Iterator[PartnerListItem] {
	// Determine page size (limit)
	limit := int64(50) // Default for partners API
	if opts != nil && opts.Limit != nil {
		limit = *opts.Limit
	}

	// Create a fetcher function that captures the service and options
	fetcher := func(ctx context.Context, offset, limit int64) ([]PartnerListItem, int64, error) {
		// Create a copy of options with updated offset/limit
		fetchOpts := &ListPartnersOptions{}
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

		// Partners API doesn't return total_count, so we use -1 to indicate unknown
		// The pager will continue fetching until an empty page is returned
		totalCount := int64(-1)
		if result.Count < int(limit) {
			// This is the last page
			totalCount = offset + int64(result.Count)
		}

		return result.Partners.Partners, totalCount, nil
	}

	return NewPager(ctx, fetcher, limit)
}
