package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/dillonkmcquade/price-comparison/internal/engine"
)

type ProductHandler struct {
	engine *engine.Engine
}

func NewProductHandler(e *engine.Engine) *ProductHandler {
	return &ProductHandler{
		engine: e,
	}
}

func (p *ProductHandler) get(w http.ResponseWriter, r *http.Request) {
	searchQuery := r.URL.Query().Get("search")
	if searchQuery == "" {
		http.Error(w, "Missing search parameters", http.StatusNotFound)
		return
	}
	page, err := strconv.Atoi(r.URL.Query().Get("page"))
	if err != nil {
		http.Error(w, "Missing search parameters", http.StatusNotFound)
		return
	}

	products, err := p.engine.Db.FindByName(searchQuery, page)
	if err != nil {
		log.Println(err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if len(products.Products) < 10 {
		p.engine.ScrapeAll(searchQuery)
	}
	products, err = p.engine.Db.FindByName(searchQuery, page)
	if err != nil {
		log.Println(err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	err = json.NewEncoder(w).Encode(products)
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
