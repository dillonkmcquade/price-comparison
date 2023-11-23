package engine

import (
	"encoding/json"
	"log"
	"os"

	"github.com/dillonkmcquade/price-comparison/internal/database"
	"github.com/dillonkmcquade/price-comparison/internal/scrapers"
)

type Engine struct {
	db               *database.Database
	scraperFactories []ScraperFactory
}

type ScraperFactory func(*database.Database, string) *scrapers.Scraper

// Register a new scraper to the engine
func (e *Engine) Register(f ScraperFactory) {
	e.scraperFactories = append(e.scraperFactories, f)
}

// Runs all registered scrapers
func (e *Engine) ScrapeAll(query string) {
	if len(e.scraperFactories) == 0 {
		log.Fatal("No scrapers registered\n")
	}
	for _, scraperFactory := range e.scraperFactories {
		scraper := scraperFactory(e.db, query)
		scraper.Visit()
	}
}

// Writes the contents of the db to a file
func (eng *Engine) Write(filePath string) {
	file, err := os.Create(filePath)
	if err != nil {
		log.Fatalf("Cannot create file %q: %s\n", filePath, err)
	}
	defer file.Close()

	e := json.NewEncoder(file)
	e.SetIndent("", "  ")
	products, err := eng.db.FindAll()
	if err != nil {
		log.Fatal(err)
	}

	err = e.Encode(products)

	if err != nil {
		log.Fatal(err)
	}
}

// Create a new instance of an Engine
func NewEngine(db *database.Database) *Engine {
	return &Engine{
		db: db,
	}
}
