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
	"github.com/dillonkmcquade/price-comparison/internal/middleware"
	"github.com/dillonkmcquade/price-comparison/internal/scrapers"
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

	file, err := os.CreateTemp("/tmp", "price_comparison_errorLogs-")
	if err != nil {
		log.Fatal(err)
	}

	// Create json structured logger that writes to file above
	l := slog.New(slog.NewJSONHandler(file, nil))

	// Initialize engine (scraper container)
	engine := engine.NewEngine(l, db)

	// Register various scraper factories
	engine.Register(scrapers.NewIgaScraper)
	engine.Register(scrapers.NewMetroScraper)

	// HTTP Router
	mux := http.NewServeMux()
	fs := http.FileServer(http.Dir("client/dist/"))
	mux.Handle("/", fs)

	mux.Handle("/api/products", handlers.NewProductHandler(engine))

	router := middleware.Logger(middleware.Cors(mux))

	server := &http.Server{
		Addr:         ":3001",
		Handler:      router,
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
	if err = openBrowser(); err != nil {
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
