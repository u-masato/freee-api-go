package accounting

import (
	"context"
	"fmt"

	"github.com/muno/freee-api-go/internal/gen"
)

// ListSectionsOptions contains optional parameters for listing sections.
type ListSectionsOptions struct {
	// StartUpdateDate filters by update date start (yyyy-mm-dd)
	StartUpdateDate *string

	// EndUpdateDate filters by update date end (yyyy-mm-dd)
	EndUpdateDate *string
}

// ListSectionsResult contains the result of listing sections.
type ListSectionsResult struct {
	// Sections is the list of sections
	Sections []gen.Section

	// Count is the number of sections returned in this response
	Count int
}

// List retrieves a list of sections for the specified company.
//
// This method returns all sections matching the optional filter criteria.
// Use ListSectionsOptions to filter by update date range.
//
// Example:
//
//	opts := &accounting.ListSectionsOptions{
//	    StartUpdateDate: stringPtr("2024-01-01"),
//	}
//	result, err := sectionsService.List(ctx, companyID, opts)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	for _, section := range result.Sections {
//	    fmt.Printf("Section ID: %d, Name: %s\n", section.Id, section.Name)
//	}
func (s *SectionsService) List(ctx context.Context, companyID int64, opts *ListSectionsOptions) (*ListSectionsResult, error) {
	// Build parameters
	params := &gen.GetSectionsParams{
		CompanyId: companyID,
	}

	if opts != nil {
		params.StartUpdateDate = opts.StartUpdateDate
		params.EndUpdateDate = opts.EndUpdateDate
	}

	// Call the generated client
	resp, err := s.genClient.GetSectionsWithResponse(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to list sections: %w", err)
	}

	// Handle error responses
	if resp.JSON200 == nil {
		return nil, fmt.Errorf("unexpected response status: %s", resp.Status())
	}

	// Return the result
	return &ListSectionsResult{
		Sections: resp.JSON200.Sections,
		Count:    len(resp.JSON200.Sections),
	}, nil
}

// Get retrieves a single section by ID.
//
// Example:
//
//	section, err := sectionsService.Get(ctx, companyID, sectionID)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("Section: %+v\n", section)
func (s *SectionsService) Get(ctx context.Context, companyID int64, sectionID int64) (*gen.SectionResponse, error) {
	// Build parameters
	params := &gen.GetSectionParams{
		CompanyId: companyID,
	}

	// Call the generated client
	resp, err := s.genClient.GetSectionWithResponse(ctx, sectionID, params)
	if err != nil {
		return nil, fmt.Errorf("failed to get section: %w", err)
	}

	// Handle error responses
	if resp.JSON200 == nil {
		return nil, fmt.Errorf("unexpected response status: %s", resp.Status())
	}

	return resp.JSON200, nil
}

// Create creates a new section.
//
// The params parameter should contain all required fields for creating a section,
// including company ID and name.
//
// Example:
//
//	params := gen.SectionParams{
//	    CompanyId: companyID,
//	    Name:      "新規部門",
//	}
//	section, err := sectionsService.Create(ctx, params)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("Created section ID: %d\n", section.Section.Id)
func (s *SectionsService) Create(ctx context.Context, params gen.SectionParams) (*gen.SectionResponse, error) {
	// Call the generated client
	resp, err := s.genClient.CreateSectionWithResponse(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to create section: %w", err)
	}

	// Handle error responses
	if resp.JSON201 == nil {
		return nil, fmt.Errorf("unexpected response status: %s", resp.Status())
	}

	return resp.JSON201, nil
}

// Update updates an existing section.
//
// The params parameter should contain the fields to update.
//
// Example:
//
//	params := gen.SectionParams{
//	    CompanyId: companyID,
//	    Name:      "更新後部門",
//	}
//	section, err := sectionsService.Update(ctx, sectionID, params)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("Updated section ID: %d\n", section.Section.Id)
func (s *SectionsService) Update(ctx context.Context, sectionID int64, params gen.SectionParams) (*gen.SectionResponse, error) {
	// Call the generated client
	resp, err := s.genClient.UpdateSectionWithResponse(ctx, sectionID, params)
	if err != nil {
		return nil, fmt.Errorf("failed to update section: %w", err)
	}

	// Handle error responses
	if resp.JSON200 == nil {
		return nil, fmt.Errorf("unexpected response status: %s", resp.Status())
	}

	return resp.JSON200, nil
}

// Delete deletes a section by ID.
//
// Example:
//
//	err := sectionsService.Delete(ctx, companyID, sectionID)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println("Section deleted successfully")
func (s *SectionsService) Delete(ctx context.Context, companyID int64, sectionID int64) error {
	// Build parameters
	params := &gen.DestroySectionParams{
		CompanyId: companyID,
	}

	// Call the generated client
	resp, err := s.genClient.DestroySectionWithResponse(ctx, sectionID, params)
	if err != nil {
		return fmt.Errorf("failed to delete section: %w", err)
	}

	// Check for error responses
	if resp.StatusCode() >= 400 {
		return fmt.Errorf("failed to delete section: %s", resp.Status())
	}

	return nil
}
