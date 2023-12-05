package scrapers

import (
	"net/url"
	"strconv"
	"strings"

	"github.com/gocolly/colly/v2"
)

/*
 |-------------------------------------------------------------------------
 | Extending the list of different scrapers
 |-------------------------------------------------------------------------
 |   Add more URL's as needed here.
 |   You can also add more scrapers by creating a new scraper factory
 |   and registering it on the engine.
 |
*/

const (
	IGA_URL   string = "https://www.iga.net/en/search"
	METRO_URL string = "https://www.metro.ca/en/online-grocery/search"
)

// Acts as a wrapper around [colly.Collector].
type Scraper struct {
	*colly.Collector
	Url url.URL
}

/*
 |-------------------------------------------------------------------------
 | Utils
 |-------------------------------------------------------------------------
*/

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
