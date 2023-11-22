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
func (e *Engine) Register(c ScraperFactory) {
	e.scraperFactories = append(e.scraperFactories, c)
}

// Runs all registered scrapers
func (e *Engine) ScrapeAll(query string) {
	if len(e.scraperFactories) == 0 {
		log.Fatal("No scrapers registered")
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
		return
	}
	defer file.Close()

	e := json.NewEncoder(file)
	e.SetIndent("", "  ")
	err = e.Encode(eng.db.FindAll())

	if err != nil {
		log.Println(err)
	}
}

// Create a new instance of an Engine
func NewEngine(db *database.Database) *Engine {
	return &Engine{
		db: db,
	}
}
