package database

import (
	"encoding/json"
	"io"
)

type Page struct {
	Page       uint64     `json:"page"`       // The current page
	TotalPages int        `json:"totalPages"` // Total number of pages
	LastPage   string     `json:"lastPage"`   // Url used to navigate to page n-1
	NextPage   string     `json:"nextPage"`   // Url used to navigate to page n+1
	TotalItems int        `json:"totalItems"` // The total amount of products that match the keyword
	Count      int        `json:"count"`      // The amount of rows returned by this query
	Products   []*Product `json:"products"`   // A slice of *Product
}

// Writes the page as JSON to w
func (p *Page) ToJSON(w io.Writer) error {
	return json.NewEncoder(w).Encode(p)
}
