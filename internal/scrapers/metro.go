package scrapers

import (
	"log"
	"net/url"
	"strings"

	"github.com/dillonkmcquade/price-comparison/internal/database"
	"github.com/gocolly/colly/v2"
)

const METRO_URL string = "https://www.metro.ca/en/online-grocery/search"

// Scrapes metro and adds items to the db
func NewMetroScraper(db *database.Database, query string) *Scraper {
	metroUrl, err := url.Parse(METRO_URL)
	if err != nil {
		log.Fatal(err)
	}

	scraper := &Scraper{
		Url: *metroUrl,
	}
	SetQuery(&scraper.Url, "filter", query)

	scraper.Collector = colly.NewCollector(
		colly.AllowedDomains("www.metro.ca", "metro.ca"),
		// Cache responses to prevent multiple download of pages
		// even if the collector is restarted
		colly.CacheDir("./cache"),
		colly.MaxDepth(1),
		// Run requests in parallel
		colly.Async(),
	)
	scraper.Collector.AllowURLRevisit = false

	err = scraper.Collector.Limit(&colly.LimitRule{DomainGlob: "*", Parallelism: 2})
	if err != nil {
		log.Fatal(err)
	}

	scraper.Collector.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Attr("href")

		prefix := "/en/online-grocery/search-page-"
		if strings.HasPrefix(link, prefix) {
			err = e.Request.Visit(link)
			if err != nil && err != colly.ErrMaxDepth {
				log.Fatal(err)
			}

		}
	})

	scraper.Collector.OnHTML(".tile-product", func(e *colly.HTMLElement) {
		price, err := strToFloat(e.ChildText(".price-update"))
		if err != nil {
			log.Println("error parsing float")
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
			log.Println(err)
		}
	})

	scraper.Collector.OnRequest(func(r *colly.Request) {
		log.Println("visiting", r.URL.String())
	})

	return scraper
}
