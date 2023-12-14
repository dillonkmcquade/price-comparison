package database

import "testing"

func TestPaginated(t *testing.T) {
	result := &Result{
		TotalItems: 1,
		Count:      1,
		Products: []*Product{
			{
				Id:                   1,
				Vendor:               "IGA",
				PricePerHundredGrams: "$1.00 / 100g",
				Price:                50.99,
				Brand:                "IGA",
				Name:                 "crackers",
				Size:                 "100g",
			},
		},
		pageNumber:  0,
		SearchQuery: "crackers",
	}

	page := result.Paginated()
	if page.LastPage != "" {
		t.Error("Last page should be empty")
	}
	if page.NextPage != "" {
		t.Error("Next page should be empty")
	}
	if page.TotalPages != 1 {
		t.Error("TotalPages should be 1")
	}
	if page.Count != len(page.Products) {
		t.Error("Page count should equal the number of products")
	}
}
