package handlers

import (
	"log"
	"net/http"
	"strconv"

	"github.com/dillonkmcquade/price-comparison/internal/engine"
)

type ProductHandler struct {
	*engine.Engine
	queries map[string]string
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
	pageNumber, err := strconv.ParseUint(r.URL.Query().Get("page"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid page number", http.StatusBadRequest)
		return
	}
	if searchQuery == "" {
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
	err = result.Paginated().ToJSON(w)
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
