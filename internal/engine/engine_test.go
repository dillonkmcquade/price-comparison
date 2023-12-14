package engine

import (
	"log/slog"
	"os"
	"testing"

	"github.com/dillonkmcquade/price-comparison/internal/database"
	"github.com/dillonkmcquade/price-comparison/internal/scrapers"
)

func TestRegister(t *testing.T) {
	l := slog.New(slog.NewJSONHandler(os.Stderr, nil))
	db := database.NewDatabase(":memory:")
	defer db.Close()
	engine, err := NewEngine(l, db)
	if err != nil {
		t.Error(err)
	}
	engine.Register(scrapers.NewIgaScraper)
	if len(engine.scraperFactories) == 0 {
		t.Error("Not registering scraper factory")
	}
}

func TestNewEngine(t *testing.T) {
	_, err := NewEngine(nil, nil)
	if err == nil {
		t.Error(err)
	}
}
