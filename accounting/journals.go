package accounting

import (
	"context"
	"fmt"

	"github.com/muno/freee-api-go/internal/gen"
)

// Note: JournalsService type is declared in services.go

// DownloadJournalsOptions contains optional parameters for downloading journals.
type DownloadJournalsOptions struct {
	// Encoding specifies the character encoding (文字コード)
	// Values: "sjis" (Shift-JIS), "utf-8" (UTF-8)
	Encoding *string

	// VisibleTags specifies items to output as auxiliary subjects or comments
	// (補助科目やコメントとして出力する項目)
	// Values: "partner", "item", "tag", "section", "description",
	//         "wallet_txn_description", "segment_1_tag", "segment_2_tag",
	//         "segment_3_tag", "all"
	VisibleTags *[]string

	// VisibleIds specifies additional ID items to output (追加出力するID項目)
	// Values: "deal_id", "transfer_id", "manual_journal_id"
	VisibleIds *[]string

	// StartDate filters by start date (取得開始日 yyyy-mm-dd)
	StartDate *string

	// EndDate filters by end date (取得終了日 yyyy-mm-dd)
	EndDate *string
}

// DownloadJournalsResult contains the result of downloading journals.
type DownloadJournalsResult struct {
	// Journals contains the journal download information
	Journals gen.JournalsResponse
}

// ListManualJournalsOptions contains optional parameters for listing manual journals.
type ListManualJournalsOptions struct {
	// StartIssueDate filters by issue date start (発生日で絞込：開始日 yyyy-mm-dd)
	StartIssueDate *string

	// EndIssueDate filters by issue date end (発生日で絞込：終了日 yyyy-mm-dd)
	EndIssueDate *string

	// EntrySide filters by debit/credit side (貸借で絞込)
	// Values: "credit" (貸方), "debit" (借方)
	EntrySide *string

	// AccountItemId filters by account item ID (勘定科目IDで絞込)
	AccountItemId *int64

	// MinAmount filters by minimum amount (金額で絞込：下限)
	MinAmount *int64

	// MaxAmount filters by maximum amount (金額で絞込：上限)
	MaxAmount *int64

	// PartnerId filters by partner ID (取引先IDで絞込)
	PartnerId *int64

	// PartnerCode filters by partner code (取引先コードで絞込)
	PartnerCode *string

	// ItemId filters by item ID (品目IDで絞込)
	ItemId *int64

	// SectionId filters by section ID (部門IDで絞込)
	SectionId *int64

	// Segment1TagId filters by segment 1 tag ID (セグメント1IDで絞込)
	Segment1TagId *int64

	// Segment2TagId filters by segment 2 tag ID (セグメント2IDで絞込)
	Segment2TagId *int64

	// Segment3TagId filters by segment 3 tag ID (セグメント3IDで絞込)
	Segment3TagId *int64

	// CommentStatus filters by comment status (コメント状態で絞込)
	// Values: "posted" (コメントあり), "raised" (未解決), "resolved" (解決済み)
	CommentStatus *string

	// CommentImportant filters by important comment flag (重要コメントで絞込)
	CommentImportant *bool

	// Adjustment filters by adjustment transaction (決算整理仕訳で絞込)
	// Values: "only" (決算整理仕訳のみ), "without" (決算整理仕訳以外)
	Adjustment *string

	// TxnNumber filters by transaction number (仕訳番号で絞込)
	TxnNumber *string

	// Offset for pagination (デフォルト: 0)
	Offset *int64

	// Limit for pagination (デフォルト: 20, 最大: 500)
	Limit *int64
}

// ListManualJournalsResult contains the result of listing manual journals.
type ListManualJournalsResult struct {
	// ManualJournals is the list of manual journals
	ManualJournals []gen.ManualJournal
}

// Download initiates a journal download request.
//
// This method requests the freee API to generate journal data in the specified format.
// The actual download must be handled separately using the returned download URL.
//
// Example:
//
//	opts := &accounting.DownloadJournalsOptions{
//	    StartDate: stringPtr("2024-01-01"),
//	    EndDate:   stringPtr("2024-01-31"),
//	    Encoding:  stringPtr("utf-8"),
//	}
//	result, err := journalsService.Download(ctx, companyID, "csv", opts)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("Download ID: %d, Status: %s\n", result.Journals.Id, result.Journals.Status)
func (s *JournalsService) Download(ctx context.Context, companyID int64, downloadType string, opts *DownloadJournalsOptions) (*DownloadJournalsResult, error) {
	// Build parameters
	params := &gen.GetJournalsParams{
		CompanyId:    companyID,
		DownloadType: gen.GetJournalsParamsDownloadType(downloadType),
	}

	if opts != nil {
		if opts.Encoding != nil {
			encoding := gen.GetJournalsParamsEncoding(*opts.Encoding)
			params.Encoding = &encoding
		}

		if opts.VisibleTags != nil {
			visibleTags := make([]gen.GetJournalsParamsVisibleTags, len(*opts.VisibleTags))
			for i, tag := range *opts.VisibleTags {
				visibleTags[i] = gen.GetJournalsParamsVisibleTags(tag)
			}
			params.VisibleTags = &visibleTags
		}

		if opts.VisibleIds != nil {
			visibleIds := make([]gen.GetJournalsParamsVisibleIds, len(*opts.VisibleIds))
			for i, id := range *opts.VisibleIds {
				visibleIds[i] = gen.GetJournalsParamsVisibleIds(id)
			}
			params.VisibleIds = &visibleIds
		}

		params.StartDate = opts.StartDate
		params.EndDate = opts.EndDate
	}

	// Call the generated client
	resp, err := s.genClient.GetJournalsWithResponse(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to download journals: %w", err)
	}

	// Handle error responses
	if resp.JSON202 == nil {
		return nil, fmt.Errorf("unexpected response status: %s", resp.Status())
	}

	// Return the result
	return &DownloadJournalsResult{
		Journals: *resp.JSON202,
	}, nil
}

// List retrieves a list of manual journals for the specified company.
//
// This method returns all manual journals matching the optional filter criteria.
// Use ListManualJournalsOptions to filter by date, amount, account items, etc.
//
// Example:
//
//	opts := &accounting.ListManualJournalsOptions{
//	    StartIssueDate: stringPtr("2024-01-01"),
//	    EndIssueDate:   stringPtr("2024-01-31"),
//	    Limit:          int64Ptr(100),
//	    Offset:         int64Ptr(0),
//	}
//	result, err := journalsService.List(ctx, companyID, opts)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	for _, journal := range result.ManualJournals {
//	    fmt.Printf("Journal ID: %d, Issue Date: %s\n", journal.Id, journal.IssueDate)
//	}
func (s *JournalsService) List(ctx context.Context, companyID int64, opts *ListManualJournalsOptions) (*ListManualJournalsResult, error) {
	// Build parameters
	params := &gen.GetManualJournalsParams{
		CompanyId: companyID,
	}

	if opts != nil {
		params.StartIssueDate = opts.StartIssueDate
		params.EndIssueDate = opts.EndIssueDate
		params.AccountItemId = opts.AccountItemId
		params.MinAmount = opts.MinAmount
		params.MaxAmount = opts.MaxAmount
		params.PartnerId = opts.PartnerId
		params.PartnerCode = opts.PartnerCode
		params.ItemId = opts.ItemId
		params.SectionId = opts.SectionId
		params.Segment1TagId = opts.Segment1TagId
		params.Segment2TagId = opts.Segment2TagId
		params.Segment3TagId = opts.Segment3TagId
		params.CommentImportant = opts.CommentImportant
		params.TxnNumber = opts.TxnNumber
		params.Offset = opts.Offset
		params.Limit = opts.Limit

		if opts.EntrySide != nil {
			entrySide := gen.GetManualJournalsParamsEntrySide(*opts.EntrySide)
			params.EntrySide = &entrySide
		}

		if opts.CommentStatus != nil {
			commentStatus := gen.GetManualJournalsParamsCommentStatus(*opts.CommentStatus)
			params.CommentStatus = &commentStatus
		}

		if opts.Adjustment != nil {
			adjustment := gen.GetManualJournalsParamsAdjustment(*opts.Adjustment)
			params.Adjustment = &adjustment
		}
	}

	// Call the generated client
	resp, err := s.genClient.GetManualJournalsWithResponse(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to list manual journals: %w", err)
	}

	// Handle error responses
	if resp.JSON200 == nil {
		return nil, fmt.Errorf("unexpected response status: %s", resp.Status())
	}

	// Return the result
	return &ListManualJournalsResult{
		ManualJournals: resp.JSON200.ManualJournals,
	}, nil
}

// ListIter returns an iterator for paginated manual journal results.
//
// The iterator transparently handles pagination, automatically fetching
// new pages as needed. This is more convenient than manually managing
// offset/limit parameters.
//
// Note: The freee API does not provide total_count for manual journals,
// so the iterator will fetch pages until an empty result is returned.
//
// Example:
//
//	startDate := "2024-01-01"
//	opts := &accounting.ListManualJournalsOptions{
//	    StartIssueDate: &startDate,
//	}
//	iter := journalsService.ListIter(ctx, companyID, opts)
//	for iter.Next() {
//	    journal := iter.Value()
//	    fmt.Printf("Journal ID: %d, Issue Date: %s\n", journal.Id, journal.IssueDate)
//	}
//	if err := iter.Err(); err != nil {
//	    log.Fatal(err)
//	}
func (s *JournalsService) ListIter(ctx context.Context, companyID int64, opts *ListManualJournalsOptions) Iterator[gen.ManualJournal] {
	// Determine page size (limit)
	limit := int64(20) // Default
	if opts != nil && opts.Limit != nil {
		limit = *opts.Limit
	}

	// Create a fetcher function that captures the service and options
	fetcher := func(ctx context.Context, offset, limit int64) ([]gen.ManualJournal, int64, error) {
		// Create a copy of options with updated offset/limit
		fetchOpts := &ListManualJournalsOptions{}
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
		return result.ManualJournals, 0, nil
	}

	return NewPager(ctx, fetcher, limit)
}
