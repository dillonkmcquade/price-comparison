package scrapers

import (
	"log"
	"net/url"

	"github.com/gocolly/colly/v2"
)

type Scraper struct {
	Url       url.URL
	Collector *colly.Collector
}

// Calls the Scraper.Collector.Visit function on the Scraper.Url
func (s *Scraper) Visit() {
	err := s.Collector.Visit(s.Url.String())
	if err != nil {
		log.Fatal(err)
	}

	s.Collector.Wait()
}
