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
	"github.com/dillonkmcquade/price-comparison/internal/handlers"
	"github.com/dillonkmcquade/price-comparison/internal/scrapers"
)

func main() {
	// Initialize a new database
	db := data.NewDatabase("file::memory:?cache=shared")
	defer db.Close()

	// Initialize engine (scraper container)
	engine := engine.NewEngine(db)

	// Register various scraper factories
	engine.Register(scrapers.NewIgaScraper)
	engine.Register(scrapers.NewMetroScraper)

	// HTTP Router
	mux := http.NewServeMux()
	mux.Handle("/api/products", handlers.NewProductHandler(engine))

	server := &http.Server{
		Addr:         ":3001",
		Handler:      mux,
		ErrorLog:     log.New(os.Stderr, "", log.LstdFlags),
		IdleTimeout:  120 * time.Second,
		ReadTimeout:  1 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	go func() {
		log.Printf("Listening on port %s", server.Addr)
		err := server.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			log.Println(err)
		}
	}()

	// Listen for interrupt or terminate signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Shutdown when signal received
	log.Printf("Received %s, commencing graceful shutdown", <-sigChan)
	tc, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(tc); err != nil {
		log.Fatalf("Error shutting down server: %s", err)
	}

	log.Println("Server shutdown successfully")
}
