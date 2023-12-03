package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"strconv"

	"github.com/dillonkmcquade/price-comparison/internal/database"
	"github.com/dillonkmcquade/price-comparison/internal/engine"
)

type ProductHandler struct {
	engine  *engine.Engine
	queries map[string]string
}

type Page struct {
	Page       int                 `json:"page"`       // The current page
	TotalPages int                 `json:"totalPages"` // Total number of pages
	TotalItems int                 `json:"totalItems"` // Total number of items
	LastPage   string              `json:"lastPage"`   // Url used to navigate to page n-1
	NextPage   string              `json:"nextPage"`   // Url used to navigate to page n+1
	Count      int                 `json:"count"`      // The number of products being returned
	Products   []*database.Product `json:"products"`
}

func NewProductHandler(e *engine.Engine) *ProductHandler {
	return &ProductHandler{
		engine:  e,
		queries: map[string]string{},
	}
}

func (p *ProductHandler) get(w http.ResponseWriter, r *http.Request) {
	// Validate query parameters
	searchQuery := r.URL.Query().Get("search")
	if searchQuery == "" {
		http.Error(w, "Missing search parameters", http.StatusNotFound)
		return
	}
	pageNumber, err := strconv.Atoi(r.URL.Query().Get("page"))
	if err != nil {
		http.Error(w, "Missing search parameters", http.StatusNotFound)
		return
	}

	// Only scrape new search queries
	if _, ok := p.queries[searchQuery]; !ok {
		p.engine.ScrapeAll(searchQuery)
		p.queries[searchQuery] = ""
	}

	// retrieve items from db
	result, err := p.engine.Db.FindByName(searchQuery, pageNumber)
	if err != nil {
		log.Println(err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if result.RowCount == 0 {
		http.Error(w, "404 Not Found", http.StatusNotFound)
		return
	}

	// Create pagination struct
	page := &Page{
		Page:     pageNumber,
		NextPage: fmt.Sprintf("http://localhost:3001/api/products?search=%s&page=%d", searchQuery, pageNumber+1),
		LastPage: fmt.Sprintf("http://localhost:3001/api/products?search=%s&page=%d", searchQuery, pageNumber-1),
	}

	if pageNumber == 0 {
		page.LastPage = ""
	}
	if page.Count < 24 {
		page.NextPage = ""
	}

	page.Count = result.RowCount
	page.Products = result.Products
	page.TotalItems = result.TotalProducts
	page.TotalPages = int(math.Ceil(float64(page.TotalItems) / 24))

	err = json.NewEncoder(w).Encode(page)
	if err != nil {
		log.Println(err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

func (p *ProductHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		p.get(w, r)
	default:
		http.Error(w, "Method not allowed.", http.StatusMethodNotAllowed)
	}
}
