package engine

import (
	"github.com/dillonkmcquade/price-comparison/internal/scrapers"
)

type Engine struct {
	scrapers []scrapers.Scraper
}

func (e *Engine) Register(c *scrapers.Scraper) {
	e.scrapers = append(e.scrapers, *c)
}

func (e *Engine) ScrapeAll() {
	for _, scraper := range e.scrapers {
		go scraper.Visit()
	}
}

func NewEngine() *Engine {
	return &Engine{}
}
