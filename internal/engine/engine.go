package engine

import (
	"errors"
	"log/slog"
	"os"

	"github.com/dillonkmcquade/price-comparison/internal/database"
	"github.com/dillonkmcquade/price-comparison/internal/scrapers"
)

// # The Engine
//
//	The engine is what orchestrates all of the different scrapers and acts
//	as the container for the database. The register method allows you to
//	extend the engine to utilise various scraper configs through [engine.ScraperFactory]
type Engine struct {
	Db               *database.Database
	scraperFactories []*ScraperFactory
	log              *slog.Logger
}

type ScraperFactory func(*slog.Logger, *database.Database, string) *scrapers.Scraper

// Register a new scraper to the engine
func (e *Engine) Register(f ScraperFactory) {
	e.scraperFactories = append(e.scraperFactories, &f)
}

// Runs all registered scrapers
func (e *Engine) ScrapeAll(query string) error {
	if len(e.scraperFactories) == 0 {
		e.log.Error("No scrapers registered\n")
		os.Exit(1)
	}
	var err error
	for _, v := range e.scraperFactories {
		scraperFactory := *v
		scraper := scraperFactory(e.log, e.Db, query)
		err = scraper.Visit(scraper.Url.String())
		if err != nil {
			e.log.Error("Error visiting", "error", err)
			break
		}
		scraper.Wait()
	}
	return err
}

// Create a new instance of an Engine
func NewEngine(logger *slog.Logger, db *database.Database) (*Engine, error) {
	if db == nil {
		return &Engine{}, errors.New("No database connection provided")
	}
	if logger == nil {
		return &Engine{}, errors.New("No database connection provided")
	}
	return &Engine{
		Db:  db,
		log: logger,
	}, nil
}
