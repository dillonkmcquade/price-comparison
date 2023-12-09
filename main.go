package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	data "github.com/dillonkmcquade/price-comparison/internal/database"
	"github.com/dillonkmcquade/price-comparison/internal/engine"
	"github.com/dillonkmcquade/price-comparison/internal/handlers"
	m "github.com/dillonkmcquade/price-comparison/internal/middleware"
	"github.com/dillonkmcquade/price-comparison/internal/scrapers"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func openBrowser() error {
	var browser *exec.Cmd
	if runtime.GOOS == "darwin" {
		cmd, err := exec.LookPath("open")
		if err != nil {
			return err
		}
		browser = exec.Command(cmd, "http://localhost:3001")
	} else {
		cmd, err := exec.LookPath("xdg-open")
		if err != nil {
			return err
		}
		browser = exec.Command(cmd, "http://localhost:3001")
	}
	return browser.Run()
}

func main() {
	// Initialize a new database
	db := data.NewDatabase("file::memory:?cache=shared")
	defer db.Close()

	// temp file for logging errors
	file, err := os.CreateTemp("/tmp", "price_comparison_errorLogs-")
	if err != nil {
		log.Fatal(err)
	}

	// Create json structured logger that writes to file above
	l := slog.New(slog.NewJSONHandler(file, nil))

	// Initialize engine (scraper container)
	// Register various scraper factories
	engine := engine.NewEngine(l, db)
	engine.Register(scrapers.NewIgaScraper)
	engine.Register(scrapers.NewMetroScraper)

	// HTTP Router
	assets := http.FileServer(http.Dir("./client/dist/assets"))
	mux := chi.NewRouter()
	mux.Use(middleware.Logger)
	mux.Use(middleware.Recoverer)
	mux.Use(m.Cors)
	mux.Handle("/", http.FileServer(http.Dir("./client/dist")))
	mux.Handle("/assets/*", http.StripPrefix("/assets", assets))

	mux.Route("/api", func(r chi.Router) {
		r.Handle("/products", handlers.NewProductHandler(engine))
	})

	server := &http.Server{
		Addr:         ":3001",
		Handler:      mux,
		IdleTimeout:  120 * time.Second,
		ReadTimeout:  1 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	go func() {
		log.Printf("Listening on port %s", server.Addr)
		err := server.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			l.Error("Server error.", err)
		}
	}()

	// Listen for interrupt or terminate signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// launch browser automatically
	err = openBrowser()
	if err != nil {
		l.Error("failed to open browser", "error", err)
	}
	// Shutdown when signal received
	log.Printf("Received %s, commencing graceful shutdown", <-sigChan)
	tc, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(tc); err != nil {
		l.Error("Error shutting down server", "error", err)
	}

	log.Println("Server shutdown successfully")
}
