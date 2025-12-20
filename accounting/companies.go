package accounting

import (
	"context"
	"fmt"

	"github.com/u-masato/freee-api-go/internal/gen"
)

// ListCompaniesResult contains the result of listing companies.
type ListCompaniesResult struct {
	// Companies is the response containing companies
	Companies *gen.CompanyIndexResponse

	// Count is the number of companies returned in this response
	Count int
}

// GetCompanyOptions contains optional parameters for retrieving a company.
//
// Note: These flags are include-only parameters. Set to true to include
// the corresponding data in the response.
type GetCompanyOptions struct {
	// Details includes account items, taxes, items, partners, sections, tags, and walletables
	Details *bool

	// AccountItems includes account items in the response
	AccountItems *bool

	// Taxes includes tax codes in the response
	Taxes *bool

	// Items includes items in the response
	Items *bool

	// Partners includes partners in the response
	Partners *bool

	// Sections includes sections in the response
	Sections *bool

	// Tags includes tags in the response
	Tags *bool

	// Walletables includes walletables in the response
	Walletables *bool
}

// List retrieves a list of companies the user belongs to.
//
// Example:
//
//	result, err := companiesService.List(ctx)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	for _, company := range result.Companies.Companies {
//	    fmt.Printf("Company ID: %d, Name: %s\n", company.Id, *company.Name)
//	}
func (s *CompaniesService) List(ctx context.Context) (*ListCompaniesResult, error) {
	resp, err := s.genClient.GetCompaniesWithResponse(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list companies: %w", err)
	}

	if resp.JSON200 == nil {
		return nil, fmt.Errorf("unexpected response status: %s", resp.Status())
	}

	return &ListCompaniesResult{
		Companies: resp.JSON200,
		Count:     len(resp.JSON200.Companies),
	}, nil
}

// Get retrieves a single company by ID.
//
// Example:
//
//	opts := &accounting.GetCompanyOptions{
//	    Details: boolPtr(true),
//	}
//	company, err := companiesService.Get(ctx, companyID, opts)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("Company ID: %d, Name: %s\n", company.Company.Id, company.Company.DisplayName)
func (s *CompaniesService) Get(ctx context.Context, companyID int64, opts *GetCompanyOptions) (*gen.CompanyResponse, error) {
	params := buildGetCompanyParams(opts)

	resp, err := s.genClient.GetCompanyWithResponse(ctx, companyID, params)
	if err != nil {
		return nil, fmt.Errorf("failed to get company: %w", err)
	}

	if resp.JSON200 == nil {
		return nil, fmt.Errorf("unexpected response status: %s", resp.Status())
	}

	return resp.JSON200, nil
}

func buildGetCompanyParams(opts *GetCompanyOptions) *gen.GetCompanyParams {
	if opts == nil {
		return nil
	}

	params := &gen.GetCompanyParams{}

	if opts.Details != nil && *opts.Details {
		details := gen.GetCompanyParamsDetailsTrue
		params.Details = &details
	}
	if opts.AccountItems != nil && *opts.AccountItems {
		accountItems := gen.GetCompanyParamsAccountItemsTrue
		params.AccountItems = &accountItems
	}
	if opts.Taxes != nil && *opts.Taxes {
		taxes := gen.GetCompanyParamsTaxesTrue
		params.Taxes = &taxes
	}
	if opts.Items != nil && *opts.Items {
		items := gen.GetCompanyParamsItemsTrue
		params.Items = &items
	}
	if opts.Partners != nil && *opts.Partners {
		partners := gen.GetCompanyParamsPartnersTrue
		params.Partners = &partners
	}
	if opts.Sections != nil && *opts.Sections {
		sections := gen.GetCompanyParamsSectionsTrue
		params.Sections = &sections
	}
	if opts.Tags != nil && *opts.Tags {
		tags := gen.GetCompanyParamsTagsTrue
		params.Tags = &tags
	}
	if opts.Walletables != nil && *opts.Walletables {
		walletables := gen.GetCompanyParamsWalletablesTrue
		params.Walletables = &walletables
	}

	return params
}
