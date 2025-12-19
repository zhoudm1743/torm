package db

import (
	"fmt"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// **Property 8: Pagination Completeness**
// **Validates: Requirements 8.1, 8.2, 8.3, 8.4**

// setupPaginationTestTable creates a test table for pagination tests
func setupPaginationTestTable(conn Connection, tableName string) error {
	driver := conn.GetDriver()
	var createSQL string

	switch driver {
	case "postgres":
		createSQL = fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s (
			id SERIAL PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			value INTEGER DEFAULT 0
		)`, tableName)
	case "mysql":
		createSQL = fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s (
			id INT AUTO_INCREMENT PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			value INT DEFAULT 0
		)`, tableName)
	default: // sqlite
		createSQL = fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			value INTEGER DEFAULT 0
		)`, tableName)
	}

	_, err := conn.Exec(createSQL)
	return err
}

// insertPaginationTestData inserts test data for pagination tests
func insertPaginationTestData(conn Connection, tableName string, count int) error {
	for i := 0; i < count; i++ {
		q := NewQuery(conn).Table(tableName)
		_, err := q.Insert(map[string]interface{}{
			"name":  fmt.Sprintf("item_%d", i),
			"value": i,
		})
		if err != nil {
			return err
		}
	}
	return nil
}

// TestProperty8_PaginationCompleteness tests pagination completeness
// For any dataset, iterating through all pages via Paginate() SHALL return
// every record exactly once, with no duplicates and no gaps.
func TestProperty8_PaginationCompleteness(t *testing.T) {
	// Setup SQLite in-memory database for testing
	manager := NewDbManager()
	defer manager.CloseAll()

	config := &Config{
		Driver:   "sqlite",
		Database: ":memory:",
	}

	err := manager.SetConfig("pagination_test", config)
	if err != nil {
		t.Fatalf("Failed to set config: %v", err)
	}

	conn, err := manager.Connection("pagination_test")
	if err != nil {
		t.Fatalf("Failed to get connection: %v", err)
	}

	tableName := "pagination_test_table"
	err = setupPaginationTestTable(conn, tableName)
	if err != nil {
		t.Fatalf("Failed to setup test table: %v", err)
	}
	defer dropTestTable(conn, tableName)

	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	// Property 8.1: Paginate returns correct metadata
	// Validates: Requirements 8.1, 8.2
	properties.Property("Paginate returns correct metadata", prop.ForAll(
		func(totalRecords int, perPage int) bool {
			if totalRecords <= 0 || totalRecords > 50 || perPage <= 0 || perPage > 20 {
				return true // Skip invalid inputs
			}

			// Clean and insert test data
			cleanupTestTable(conn, tableName)
			err := insertPaginationTestData(conn, tableName, totalRecords)
			if err != nil {
				t.Logf("Failed to insert test data: %v", err)
				return false
			}

			// Test pagination
			q := NewQuery(conn).Table(tableName).OrderBy("id", "ASC")
			paginator, err := q.Paginate(1, perPage)
			if err != nil {
				t.Logf("Paginate failed: %v", err)
				return false
			}

			// Verify total count
			if paginator.Total != int64(totalRecords) {
				t.Logf("Total mismatch: expected %d, got %d", totalRecords, paginator.Total)
				return false
			}

			// Verify per page
			if paginator.PerPage != perPage {
				t.Logf("PerPage mismatch: expected %d, got %d", perPage, paginator.PerPage)
				return false
			}

			// Verify current page
			if paginator.CurrentPage != 1 {
				t.Logf("CurrentPage mismatch: expected 1, got %d", paginator.CurrentPage)
				return false
			}

			// Verify last page calculation
			expectedLastPage := (totalRecords + perPage - 1) / perPage
			if paginator.LastPage != expectedLastPage {
				t.Logf("LastPage mismatch: expected %d, got %d", expectedLastPage, paginator.LastPage)
				return false
			}

			// Verify data count on first page
			expectedDataCount := perPage
			if totalRecords < perPage {
				expectedDataCount = totalRecords
			}
			if len(paginator.Data) != expectedDataCount {
				t.Logf("Data count mismatch: expected %d, got %d", expectedDataCount, len(paginator.Data))
				return false
			}

			return true
		},
		gen.IntRange(1, 50),
		gen.IntRange(1, 20),
	))

	// Property 8.2: Iterating all pages returns all records exactly once
	// Validates: Requirements 8.1, 8.2
	properties.Property("Iterating all pages returns all records exactly once", prop.ForAll(
		func(totalRecords int, perPage int) bool {
			if totalRecords <= 0 || totalRecords > 30 || perPage <= 0 || perPage > 10 {
				return true // Skip invalid inputs
			}

			// Clean and insert test data
			cleanupTestTable(conn, tableName)
			err := insertPaginationTestData(conn, tableName, totalRecords)
			if err != nil {
				t.Logf("Failed to insert test data: %v", err)
				return false
			}

			// Collect all IDs from all pages
			seenIDs := make(map[interface{}]bool)
			totalDataCount := 0

			expectedLastPage := (totalRecords + perPage - 1) / perPage

			for page := 1; page <= expectedLastPage; page++ {
				q := NewQuery(conn).Table(tableName).OrderBy("id", "ASC")
				paginator, err := q.Paginate(page, perPage)
				if err != nil {
					t.Logf("Paginate failed on page %d: %v", page, err)
					return false
				}

				for _, row := range paginator.Data {
					id := row["id"]
					if seenIDs[id] {
						t.Logf("Duplicate ID found: %v", id)
						return false
					}
					seenIDs[id] = true
					totalDataCount++
				}
			}

			// Verify all records were returned
			if totalDataCount != totalRecords {
				t.Logf("Total data count mismatch: expected %d, got %d", totalRecords, totalDataCount)
				return false
			}

			return true
		},
		gen.IntRange(1, 30),
		gen.IntRange(1, 10),
	))

	// Property 8.3: SimplePaginate correctly indicates hasMore
	// Validates: Requirements 8.3
	properties.Property("SimplePaginate correctly indicates hasMore", prop.ForAll(
		func(totalRecords int, perPage int) bool {
			if totalRecords <= 0 || totalRecords > 30 || perPage <= 0 || perPage > 10 {
				return true // Skip invalid inputs
			}

			// Clean and insert test data
			cleanupTestTable(conn, tableName)
			err := insertPaginationTestData(conn, tableName, totalRecords)
			if err != nil {
				t.Logf("Failed to insert test data: %v", err)
				return false
			}

			expectedLastPage := (totalRecords + perPage - 1) / perPage

			// Test each page
			for page := 1; page <= expectedLastPage; page++ {
				q := NewQuery(conn).Table(tableName).OrderBy("id", "ASC")
				simplePaginator, err := q.SimplePaginate(page, perPage)
				if err != nil {
					t.Logf("SimplePaginate failed on page %d: %v", page, err)
					return false
				}

				// hasMore should be true if there are more pages
				expectedHasMore := page < expectedLastPage
				if simplePaginator.HasMore != expectedHasMore {
					t.Logf("HasMore mismatch on page %d: expected %v, got %v", page, expectedHasMore, simplePaginator.HasMore)
					return false
				}
			}

			return true
		},
		gen.IntRange(1, 30),
		gen.IntRange(1, 10),
	))

	// Property 8.4: From and To values are correct
	// Validates: Requirements 8.2
	properties.Property("From and To values are correct", prop.ForAll(
		func(totalRecords int, perPage int, page int) bool {
			if totalRecords <= 0 || totalRecords > 50 || perPage <= 0 || perPage > 20 || page <= 0 {
				return true // Skip invalid inputs
			}

			// Clean and insert test data
			cleanupTestTable(conn, tableName)
			err := insertPaginationTestData(conn, tableName, totalRecords)
			if err != nil {
				t.Logf("Failed to insert test data: %v", err)
				return false
			}

			expectedLastPage := (totalRecords + perPage - 1) / perPage
			if page > expectedLastPage {
				page = expectedLastPage // Clamp to valid page
			}

			q := NewQuery(conn).Table(tableName).OrderBy("id", "ASC")
			paginator, err := q.Paginate(page, perPage)
			if err != nil {
				t.Logf("Paginate failed: %v", err)
				return false
			}

			// Calculate expected From and To
			expectedFrom := (page-1)*perPage + 1
			expectedTo := page * perPage
			if expectedTo > totalRecords {
				expectedTo = totalRecords
			}

			// If no data, From and To should be 0
			if len(paginator.Data) == 0 {
				if paginator.From != 0 || paginator.To != 0 {
					t.Logf("From/To should be 0 for empty page, got From=%d, To=%d", paginator.From, paginator.To)
					return false
				}
				return true
			}

			if paginator.From != expectedFrom {
				t.Logf("From mismatch: expected %d, got %d", expectedFrom, paginator.From)
				return false
			}

			// To should be From + len(Data) - 1
			actualTo := paginator.From + len(paginator.Data) - 1
			if paginator.To != actualTo {
				t.Logf("To mismatch: expected %d, got %d", actualTo, paginator.To)
				return false
			}

			return true
		},
		gen.IntRange(1, 50),
		gen.IntRange(1, 20),
		gen.IntRange(1, 10),
	))

	// Property 8.5: Paginator helper methods work correctly
	// Validates: Requirements 8.1, 8.2
	properties.Property("Paginator helper methods work correctly", prop.ForAll(
		func(totalRecords int, perPage int, page int) bool {
			if totalRecords <= 0 || totalRecords > 50 || perPage <= 0 || perPage > 20 || page <= 0 {
				return true // Skip invalid inputs
			}

			// Clean and insert test data
			cleanupTestTable(conn, tableName)
			err := insertPaginationTestData(conn, tableName, totalRecords)
			if err != nil {
				t.Logf("Failed to insert test data: %v", err)
				return false
			}

			expectedLastPage := (totalRecords + perPage - 1) / perPage
			if page > expectedLastPage {
				page = expectedLastPage // Clamp to valid page
			}

			q := NewQuery(conn).Table(tableName).OrderBy("id", "ASC")
			paginator, err := q.Paginate(page, perPage)
			if err != nil {
				t.Logf("Paginate failed: %v", err)
				return false
			}

			// Test HasPages
			expectedHasPages := expectedLastPage > 1
			if paginator.HasPages() != expectedHasPages {
				t.Logf("HasPages mismatch: expected %v, got %v", expectedHasPages, paginator.HasPages())
				return false
			}

			// Test HasMorePages
			expectedHasMorePages := page < expectedLastPage
			if paginator.HasMorePages() != expectedHasMorePages {
				t.Logf("HasMorePages mismatch: expected %v, got %v", expectedHasMorePages, paginator.HasMorePages())
				return false
			}

			// Test OnFirstPage
			expectedOnFirstPage := page <= 1
			if paginator.OnFirstPage() != expectedOnFirstPage {
				t.Logf("OnFirstPage mismatch: expected %v, got %v", expectedOnFirstPage, paginator.OnFirstPage())
				return false
			}

			// Test OnLastPage
			expectedOnLastPage := page >= expectedLastPage
			if paginator.OnLastPage() != expectedOnLastPage {
				t.Logf("OnLastPage mismatch: expected %v, got %v", expectedOnLastPage, paginator.OnLastPage())
				return false
			}

			// Test PreviousPage
			expectedPreviousPage := page - 1
			if expectedPreviousPage < 1 {
				expectedPreviousPage = 1
			}
			if paginator.PreviousPage() != expectedPreviousPage {
				t.Logf("PreviousPage mismatch: expected %d, got %d", expectedPreviousPage, paginator.PreviousPage())
				return false
			}

			// Test NextPage
			expectedNextPage := page + 1
			if expectedNextPage > expectedLastPage {
				expectedNextPage = expectedLastPage
			}
			if paginator.NextPage() != expectedNextPage {
				t.Logf("NextPage mismatch: expected %d, got %d", expectedNextPage, paginator.NextPage())
				return false
			}

			return true
		},
		gen.IntRange(1, 50),
		gen.IntRange(1, 20),
		gen.IntRange(1, 10),
	))

	// Property 8.6: SimplePaginator helper methods work correctly
	// Validates: Requirements 8.3
	properties.Property("SimplePaginator helper methods work correctly", prop.ForAll(
		func(totalRecords int, perPage int, page int) bool {
			if totalRecords <= 0 || totalRecords > 30 || perPage <= 0 || perPage > 10 || page <= 0 {
				return true // Skip invalid inputs
			}

			// Clean and insert test data
			cleanupTestTable(conn, tableName)
			err := insertPaginationTestData(conn, tableName, totalRecords)
			if err != nil {
				t.Logf("Failed to insert test data: %v", err)
				return false
			}

			expectedLastPage := (totalRecords + perPage - 1) / perPage
			if page > expectedLastPage {
				page = expectedLastPage // Clamp to valid page
			}

			q := NewQuery(conn).Table(tableName).OrderBy("id", "ASC")
			simplePaginator, err := q.SimplePaginate(page, perPage)
			if err != nil {
				t.Logf("SimplePaginate failed: %v", err)
				return false
			}

			// Test HasMorePages
			expectedHasMore := page < expectedLastPage
			if simplePaginator.HasMorePages() != expectedHasMore {
				t.Logf("HasMorePages mismatch: expected %v, got %v", expectedHasMore, simplePaginator.HasMorePages())
				return false
			}

			// Test OnFirstPage
			expectedOnFirstPage := page <= 1
			if simplePaginator.OnFirstPage() != expectedOnFirstPage {
				t.Logf("OnFirstPage mismatch: expected %v, got %v", expectedOnFirstPage, simplePaginator.OnFirstPage())
				return false
			}

			// Test PreviousPage
			expectedPreviousPage := page - 1
			if expectedPreviousPage < 1 {
				expectedPreviousPage = 1
			}
			if simplePaginator.PreviousPage() != expectedPreviousPage {
				t.Logf("PreviousPage mismatch: expected %d, got %d", expectedPreviousPage, simplePaginator.PreviousPage())
				return false
			}

			// Test NextPage
			expectedNextPage := page + 1
			if !expectedHasMore {
				expectedNextPage = page
			}
			if simplePaginator.NextPage() != expectedNextPage {
				t.Logf("NextPage mismatch: expected %d, got %d", expectedNextPage, simplePaginator.NextPage())
				return false
			}

			return true
		},
		gen.IntRange(1, 30),
		gen.IntRange(1, 10),
		gen.IntRange(1, 5),
	))

	// Property 8.7: Empty result pagination works correctly
	// Validates: Requirements 8.1, 8.2
	properties.Property("Empty result pagination works correctly", prop.ForAll(
		func(perPage int) bool {
			if perPage <= 0 || perPage > 20 {
				return true // Skip invalid inputs
			}

			// Clean table (no data)
			cleanupTestTable(conn, tableName)

			q := NewQuery(conn).Table(tableName).OrderBy("id", "ASC")
			paginator, err := q.Paginate(1, perPage)
			if err != nil {
				t.Logf("Paginate failed: %v", err)
				return false
			}

			// Verify empty result
			if paginator.Total != 0 {
				t.Logf("Total should be 0, got %d", paginator.Total)
				return false
			}

			if len(paginator.Data) != 0 {
				t.Logf("Data should be empty, got %d items", len(paginator.Data))
				return false
			}

			if paginator.From != 0 || paginator.To != 0 {
				t.Logf("From and To should be 0 for empty result, got From=%d, To=%d", paginator.From, paginator.To)
				return false
			}

			if paginator.LastPage != 1 {
				t.Logf("LastPage should be 1 for empty result, got %d", paginator.LastPage)
				return false
			}

			// Test IsEmpty
			if !paginator.IsEmpty() {
				t.Log("IsEmpty should return true for empty result")
				return false
			}

			// Test IsNotEmpty
			if paginator.IsNotEmpty() {
				t.Log("IsNotEmpty should return false for empty result")
				return false
			}

			return true
		},
		gen.IntRange(1, 20),
	))

	// Property 8.8: GetPageRange returns correct page range
	// Validates: Requirements 8.1
	properties.Property("GetPageRange returns correct page range", prop.ForAll(
		func(totalRecords int, perPage int, page int, onEachSide int) bool {
			if totalRecords <= 0 || totalRecords > 100 || perPage <= 0 || perPage > 10 || page <= 0 || onEachSide <= 0 || onEachSide > 5 {
				return true // Skip invalid inputs
			}

			// Clean and insert test data
			cleanupTestTable(conn, tableName)
			err := insertPaginationTestData(conn, tableName, totalRecords)
			if err != nil {
				t.Logf("Failed to insert test data: %v", err)
				return false
			}

			expectedLastPage := (totalRecords + perPage - 1) / perPage
			if page > expectedLastPage {
				page = expectedLastPage // Clamp to valid page
			}

			q := NewQuery(conn).Table(tableName).OrderBy("id", "ASC")
			paginator, err := q.Paginate(page, perPage)
			if err != nil {
				t.Logf("Paginate failed: %v", err)
				return false
			}

			pageRange := paginator.GetPageRange(onEachSide)

			// Verify page range is not empty
			if len(pageRange) == 0 {
				t.Log("Page range should not be empty")
				return false
			}

			// Verify all pages in range are valid
			for _, p := range pageRange {
				if p < 1 || p > expectedLastPage {
					t.Logf("Invalid page in range: %d (lastPage=%d)", p, expectedLastPage)
					return false
				}
			}

			// Verify current page is in range (if within bounds)
			found := false
			for _, p := range pageRange {
				if p == page {
					found = true
					break
				}
			}
			if !found {
				t.Logf("Current page %d not found in range %v", page, pageRange)
				return false
			}

			// Verify pages are in ascending order
			for i := 1; i < len(pageRange); i++ {
				if pageRange[i] <= pageRange[i-1] {
					t.Logf("Pages not in ascending order: %v", pageRange)
					return false
				}
			}

			return true
		},
		gen.IntRange(1, 100),
		gen.IntRange(1, 10),
		gen.IntRange(1, 15),
		gen.IntRange(1, 5),
	))

	properties.TestingRun(t)
}
