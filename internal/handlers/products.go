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
	*engine.Engine
	queries map[string]string
}

type Page struct {
	Page       int    `json:"page"`       // The current page
	TotalPages int    `json:"totalPages"` // Total number of pages
	LastPage   string `json:"lastPage"`   // Url used to navigate to page n-1
	NextPage   string `json:"nextPage"`   // Url used to navigate to page n+1
	*database.Result
}

func NewProductHandler(e *engine.Engine) *ProductHandler {
	return &ProductHandler{
		Engine:  e,
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
		err = p.ScrapeAll(searchQuery)
		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		p.queries[searchQuery] = ""
	}

	// retrieve items from db
	result, err := p.Db.FindByName(searchQuery, pageNumber)
	if err != nil {
		log.Println(err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if result.Count == 0 {
		http.Error(w, "404 Not Found", http.StatusNotFound)
		return
	}

	// Create pagination struct
	page := &Page{
		Page:     pageNumber,
		NextPage: fmt.Sprintf("http://localhost:3001/api/products?search=%s&page=%d", searchQuery, pageNumber+1),
		LastPage: fmt.Sprintf("http://localhost:3001/api/products?search=%s&page=%d", searchQuery, pageNumber-1),
		Result:   result,
	}

	if pageNumber == 0 {
		page.LastPage = ""
	}
	if result.Count < 24 {
		page.NextPage = ""
	}

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
