package scrapers

import (
	"log"
	"net/url"
	"strings"

	"github.com/dillonkmcquade/price-comparison/internal/database"
	"github.com/gocolly/colly/v2"
)

const METRO_URL string = "https://www.metro.ca/en/online-grocery/search"

func NewMetroScraper(db *database.Database, query string) *Scraper {
	metroUrl, err := url.Parse(METRO_URL)
	if err != nil {
		log.Fatal(err)
	}

	scraper := &Scraper{
		Url: *metroUrl,
	}
	setQuery(&scraper.Url, "filter", query)

	scraper.Collector = colly.NewCollector(
		// Visit only domains
		colly.AllowedDomains("www.metro.ca", "metro.ca"),

		// Cache responses to prevent multiple download of pages
		// even if the collector is restarted
		colly.CacheDir("./cache"),
		//
		colly.MaxDepth(4),
		// Run requests in parallel
		colly.Async(),
	)
	scraper.Collector.AllowURLRevisit = false

	scraper.Collector.Limit(&colly.LimitRule{DomainGlob: "*", Parallelism: 2})

	scraper.Collector.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Attr("href")

		prefix := strings.Join([]string{scraper.Url.Path, scraper.Url.Query().Encode()}, "?")
		if strings.HasPrefix(link, prefix) {
			e.Request.Visit(link)
		}
	})

	scraper.Collector.OnHTML(".item-product.js-product", func(e *colly.HTMLElement) {
		log.Printf("Reading from product %d of page %s", e.Index, e.Request.URL.String())
		product := &database.Product{
			Vendor:               "IGA",
			Brand:                e.ChildText(".item-product__brand"),
			Price:                e.ChildText("span.price"),
			Name:                 e.ChildText(".js-ga-productname"),
			Image:                e.ChildAttr(".js-ga-productimage > img", "src"),
			Size:                 e.ChildTexts(".item-product__info")[0],
			PricePerHundredGrams: e.ChildText(".item-product__info > div.text--small"),
		}
		err := db.Insert(product)
		if err != nil {
			log.Println(err)
		}
	})

	scraper.Collector.OnRequest(func(r *colly.Request) {
		log.Println("visiting", r.URL.String())
	})

	return scraper
}
