# Workflow Validation Fixtures

This directory contains workflow JSON fixtures used for MVP validation testing.

## airbnb-validation.json

This workflow simulates the complete Airbnb price monitoring scraper flow.

### Workflow Flow

```
[http_request] --> [html_parser] --> [transform]
     |                  |                 |
  Fetch HTML      Extract data      Format output
  from URL        with CSS           as JSON
                  selectors
```

### Tasks

1. **fetch_listings** (`http_request`)
   - Fetches HTML page from the target URL
   - Output stored as `fetch_listings_result` in context
   - Contains: `{status_code, headers, body}`

2. **parse_listings** (`html_parser`)
   - Reads HTML from `fetch_listings_result.body`
   - Extracts using CSS selectors:
     - `.listing-title` -> titles
     - `.price-amount` -> prices
     - `.listing-location` -> locations
     - `.rating-score` -> ratings
     - `.listing-image[src]` -> images (attribute)
     - `.view-link[href]` -> urls (attribute)
   - Output stored as `parse_listings_result`
   - Returns: `[{titles: [...], prices: [...], ...}]`

3. **format_output** (`transform`)
   - Transforms parsed data into final JSON structure
   - Uses Go text/template syntax
   - Output stored as `format_output_result`
   - Returns: `{listing_count: N, listings: [...]}`

### Usage

The workflow JSON requires the `{{MOCK_SERVER_URL}}` placeholder to be replaced with the actual target URL before execution.

For E2E testing, see `internal/e2e/airbnb_validation_test.go`.

### Expected Context Snapshot

After successful execution, the `context_snapshot` will contain:

```json
{
  "fetch_listings_result": {
    "status_code": 200,
    "headers": {...},
    "body": "<html>..."
  },
  "parse_listings_result": [{
    "titles": ["Beautiful Beach House", "Cozy Mountain Cabin", "Modern City Loft"],
    "prices": ["$350", "$275", "$450"],
    "locations": ["Malibu, CA", "Aspen, CO", "New York, NY"],
    "ratings": ["4.9", "4.8", "4.7"]
  }],
  "format_output_result": {
    "listing_count": 3,
    "listings": [...]
  }
}
```
