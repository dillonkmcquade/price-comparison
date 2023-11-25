package main

import (
	"log"

	data "github.com/dillonkmcquade/price-comparison/internal/database"
	"github.com/dillonkmcquade/price-comparison/internal/engine"
	"github.com/dillonkmcquade/price-comparison/internal/scrapers"
)

func main() {
	// Initialize new database
	db := data.NewDatabase("file::memory:?cache=shared")

	// Initialize engine (scraper container)
	engine := engine.NewEngine(db)

	// Register various scraper factories
	engine.Register(scrapers.NewIgaScraper)
	engine.Register(scrapers.NewMetroScraper)

	// Scrape
	engine.ScrapeAll("carrots")

	// Write results to file
	engine.Write("products.json")

	log.Println("Finished")
}
