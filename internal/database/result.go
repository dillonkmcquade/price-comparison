package database

import (
	"fmt"
	"math"
)

type Result struct {
	TotalItems  int        // The total amount of products that match the keyword
	Count       int        // The amount of rows returned by this query
	Products    []*Product // A slice of *Product
	SearchQuery string     // The keyword contained in the product name
	pageNumber  uint64     // The current requested page
}

// Transform the database result to a pagination response
func (r *Result) Paginated() *Page {
	page := &Page{
		Page:       r.pageNumber,
		NextPage:   fmt.Sprintf("http://localhost:3001/api/products?search=%s&page=%d", r.SearchQuery, r.pageNumber+1),
		LastPage:   fmt.Sprintf("http://localhost:3001/api/products?search=%s&page=%d", r.SearchQuery, r.pageNumber-1),
		TotalPages: int(math.Ceil(float64(r.TotalItems) / 24)),
		TotalItems: r.TotalItems,
		Count:      r.Count,
		Products:   r.Products,
	}
	if r.pageNumber == 0 {
		page.LastPage = ""
	}
	if r.Count < 24 {
		page.NextPage = ""
	}

	return page
}
