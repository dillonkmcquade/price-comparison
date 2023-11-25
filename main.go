package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

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
	log.Println("Finished Scraping all items")

	mux := http.NewServeMux()

	// mux.Handle("/products", )

	server := &http.Server{
		Addr:         ":3001",
		Handler:      mux,
		ErrorLog:     log.New(os.Stderr, "", log.LstdFlags),
		IdleTimeout:  120 * time.Second,
		ReadTimeout:  2 * time.Second,
		WriteTimeout: 2 * time.Second,
	}

	go func() {
		log.Printf("Listening on port %s", server.Addr)
		err := server.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			log.Println(err)
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	log.Printf("Received %s, commencing graceful shutdown", <-sigChan)

	tc, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(tc); err != nil {
		log.Fatalf("Error shutting down server: %s", err)
	}

	log.Println("Server shutdown successfully")
}
