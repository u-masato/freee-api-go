# Iterator Example

This example demonstrates how to use the Iterator pattern to transparently handle pagination when retrieving data from the freee API.

## Overview

The Iterator pattern provides a simple and efficient way to iterate through paginated API results without manually managing offset/limit parameters. The iterator automatically fetches new pages as needed.

## Features

- **Transparent Pagination**: Automatically fetches pages as you iterate
- **Memory Efficient**: Only one page of results is kept in memory at a time
- **Error Handling**: Errors are captured and can be checked after iteration
- **Early Termination**: You can break out of iteration at any time
- **Configurable Page Size**: Control how many items to fetch per API request

## Prerequisites

1. OAuth2 token saved in `token.json` (see the `examples/oauth` directory)
2. Environment variables:
   - `FREEE_COMPANY_ID`: Your freee company ID
   - `FREEE_CLIENT_ID`: Your OAuth2 client ID
   - `FREEE_CLIENT_SECRET`: Your OAuth2 client secret
   - `FREEE_REDIRECT_URL`: Your OAuth2 redirect URL

## Running the Example

```bash
# Set environment variables
export FREEE_COMPANY_ID="your_company_id"
export FREEE_CLIENT_ID="your_client_id"
export FREEE_CLIENT_SECRET="your_client_secret"
export FREEE_REDIRECT_URL="http://localhost:8080/callback"

# Run the example
go run main.go
```

## Usage Patterns

### Basic Iteration

```go
iter := accountingClient.Deals.ListIter(ctx, companyID, opts)
for iter.Next() {
    deal := iter.Value()
    fmt.Printf("Deal ID: %d, Amount: %d\n", deal.Id, deal.Amount)
}
if err := iter.Err(); err != nil {
    log.Fatal(err)
}
```

### Custom Page Size

```go
limit := int64(50)
opts := &accounting.ListDealsOptions{
    Type:  stringPtr("expense"),
    Limit: &limit, // Fetch 50 items per page
}
iter := accountingClient.Deals.ListIter(ctx, companyID, opts)
```

### Early Termination

```go
iter := accountingClient.Deals.ListIter(ctx, companyID, opts)
count := 0
for iter.Next() {
    deal := iter.Value()
    count++
    if count >= 10 {
        break // Stop after 10 items
    }
}
// Always check for errors, even after early termination
if err := iter.Err(); err != nil {
    log.Fatal(err)
}
```

### Filter and Aggregate

```go
iter := accountingClient.Deals.ListIter(ctx, companyID, opts)
total := int64(0)
for iter.Next() {
    deal := iter.Value()
    if deal.Amount > 100000 {
        total += deal.Amount
    }
}
if err := iter.Err(); err != nil {
    log.Fatal(err)
}
```

## How It Works

The iterator pattern works by:

1. **Lazy Loading**: Pages are only fetched when needed
2. **State Management**: The iterator keeps track of:
   - Current page of items
   - Current position within the page
   - Total number of items across all pages
   - Any errors that occurred
3. **Automatic Pagination**: When you reach the end of a page, the next page is automatically fetched

## Performance Considerations

- **Page Size**: Larger page sizes (up to 100) reduce API calls but use more memory
- **Early Termination**: If you only need a few items, break early to avoid unnecessary API calls
- **Context Cancellation**: Use context cancellation to stop iteration if needed

## Comparison with Manual Pagination

### Without Iterator (Manual)

```go
offset := int64(0)
limit := int64(20)
for {
    opts := &accounting.ListDealsOptions{
        Offset: &offset,
        Limit:  &limit,
    }
    result, err := dealsService.List(ctx, companyID, opts)
    if err != nil {
        log.Fatal(err)
    }

    for _, deal := range result.Deals {
        // Process deal
    }

    offset += int64(len(result.Deals))
    if offset >= result.TotalCount {
        break
    }
}
```

### With Iterator (Automatic)

```go
iter := dealsService.ListIter(ctx, companyID, opts)
for iter.Next() {
    deal := iter.Value()
    // Process deal
}
if err := iter.Err(); err != nil {
    log.Fatal(err)
}
```

The iterator pattern significantly simplifies the code and reduces the chance of pagination errors.
