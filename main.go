package main

import (
	"encoding/json"
	"log"
	"os"

	data "github.com/dillonkmcquade/price-comparison/internal/database"
	"github.com/dillonkmcquade/price-comparison/internal/engine"
	"github.com/dillonkmcquade/price-comparison/internal/scrapers"
)

func main() {
	fName := "products.json"
	file, err := os.Create(fName)
	if err != nil {
		log.Fatalf("Cannot create file %q: %s\n", fName, err)
		return
	}
	defer file.Close()

	// Initialize new database
	db := data.NewDatabase[data.Product]()

	engine := engine.NewEngine(db)

	engine.Register(scrapers.ScrapeIga)
	engine.Register(scrapers.NewMetroScraper)

	// Scrape
	engine.ScrapeAll("carrots")

	e := json.NewEncoder(file)
	e.SetIndent("", "  ")
	err = e.Encode(db.FindAll())
	if err != nil {
		log.Println(err)
	}
	log.Println("Finished")
}
