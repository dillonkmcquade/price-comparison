package scrapers

import (
	"log/slog"
	"net/url"
	"os"
	"strings"

	"github.com/dillonkmcquade/price-comparison/internal/database"
	"github.com/gocolly/colly/v2"
)

// Scrapes metro and adds items to the db
func NewMetroScraper(l *slog.Logger, db *database.Database, query string) *Scraper {
	metroUrl, err := url.Parse(METRO_URL)
	if err != nil {
		l.Error("Error parsing url", "error", err, "url", METRO_URL)
		os.Exit(1)
	}

	scraper := &Scraper{
		Url: *metroUrl,
		Collector: colly.NewCollector(
			colly.AllowedDomains("www.metro.ca", "metro.ca"),
			// Cache responses to prevent multiple download of pages
			// even if the collector is restarted
			colly.CacheDir("./cache"),
			colly.MaxDepth(1),
			// Run requests in parallel
			colly.Async(),
		),
	}
	SetQuery(&scraper.Url, "filter", query)

	scraper.AllowURLRevisit = false

	err = scraper.Limit(&colly.LimitRule{DomainGlob: "*", Parallelism: 2})
	if err != nil {
		l.Error("colly limit rule error", "error", err)
		os.Exit(1)
	}

	scraper.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Attr("href")

		prefix := "/en/online-grocery/search-page-"
		if strings.HasPrefix(link, prefix) {
			err = e.Request.Visit(link)
			if err != nil && err != colly.ErrMaxDepth {
				l.Error("error visiting link", "error", err, "link", link)
				os.Exit(1)
			}

		}
	})

	scraper.OnHTML(".tile-product", func(e *colly.HTMLElement) {
		price, err := strToFloat(e.ChildText(".price-update"))
		if err != nil {
			l.Error("error parsing float")
			return
		}
		brand := e.ChildText(".head__brand")
		if brand == "" {
			brand = "metro"
		}
		product := &database.Product{
			Vendor:               "metro",
			Brand:                brand,
			Price:                price,
			Name:                 e.ChildText(".head__title"),
			Image:                e.ChildAttr(".defaultable-picture > img", "src"),
			Size:                 e.ChildText(".head__unit-details"),
			PricePerHundredGrams: e.ChildText(".pricing__secondary-price > span"),
		}
		/* No need to handle errors here, unique constraint failures are expected and intentional */
		_, err = db.Insert(product)
		if err != nil {
			l.Error("error executing database insert", "error", err)
		}
	})

	return scraper
}
