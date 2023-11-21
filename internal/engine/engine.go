package engine

import (
	"log"

	"github.com/dillonkmcquade/price-comparison/internal/database"
	"github.com/dillonkmcquade/price-comparison/internal/scrapers"
)

type Engine struct {
	db               *database.Database[database.Product]
	scraperFunctions []ScraperFunction
}

type ScraperFunction func(*database.Database[database.Product], string) *scrapers.Scraper

// Register a new scraper to the engine
func (e *Engine) Register(c ScraperFunction) {
	e.scraperFunctions = append(e.scraperFunctions, c)
}

// Runs all registered scrapers
func (e *Engine) ScrapeAll(query string) {
	if len(e.scraperFunctions) == 0 {
		log.Fatal("No scrapers registered")
	}
	for _, scraperFunction := range e.scraperFunctions {
		scraper := scraperFunction(e.db, query)
		scraper.Visit()
		log.Printf("Scraping %s", scraper.Url.String())
	}
}

// Create a new instance of an Engine
func NewEngine(db *database.Database[database.Product]) *Engine {
	return &Engine{
		db: db,
	}
}
