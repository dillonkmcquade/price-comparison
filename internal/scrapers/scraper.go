package scrapers

import (
	"log"
	"net/url"
	"strconv"
	"strings"

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

// Adds a query parameter to a url
func SetQuery(u *url.URL, k string, v string) *url.URL {
	q := u.Query()
	q.Set(k, v)
	u.RawQuery = q.Encode()
	return u
}

// converts string price `$59.99` to float 59.99
func strToFloat(s string) (float64, error) {
	split := strings.Split(s, "$")[1]
	price, err := strconv.ParseFloat(split, 64)
	return price, err
}
