package scrapers

import (
	"log"
	"log/slog"
	"net/url"
	"os"
	"strings"

	"github.com/dillonkmcquade/price-comparison/internal/database"
	"github.com/gocolly/colly/v2"
)

const IGA_URL string = "https://www.iga.net/en/search"

// Scrapes IGA and adds items to the db
func NewIgaScraper(l *slog.Logger, db *database.Database, query string) *Scraper {
	igaUrl, err := url.Parse(IGA_URL)
	if err != nil {
		l.Error("Error parsing url", "error", err, "url", IGA_URL)
		os.Exit(1)
	}
	scraper := &Scraper{
		Url: *igaUrl,
		Collector: colly.NewCollector(
			colly.AllowedDomains("www.iga.net", "iga.net"),
			// Cache responses to prevent multiple download of pages
			// even if the collector is restarted
			colly.CacheDir("./cache"),
			colly.MaxDepth(1),
			// Run requests in parallel
			colly.Async(),
		),
	}
	SetQuery(&scraper.Url, "k", query)

	scraper.AllowURLRevisit = false

	err = scraper.Limit(&colly.LimitRule{DomainGlob: "*", Parallelism: 2})
	if err != nil {
		l.Error("Colly limit rule error", "error", err)
	}

	scraper.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Attr("href")

		SetQuery(&scraper.Url, "page", "")
		prefix := strings.Join([]string{scraper.Url.Path, scraper.Url.Query().Encode()}, "?")
		if strings.HasPrefix(link, prefix) {
			err = e.Request.Visit(link)
			if err != nil && err != colly.ErrMaxDepth {
				l.Error("Error visiting link", "error", err, "link", link)
				os.Exit(1)
			}
		}
	})

	scraper.OnHTML(".item-product.js-product", func(e *colly.HTMLElement) {
		price, err := strToFloat(e.ChildTexts("span.price")[0])
		if err != nil {
			l.Error("error parsing float", "error", err)
			return
		}

		brand := e.ChildText(".item-product__brand")
		if brand == "" {
			brand = "IGA"
		}

		var size string
		sizeSplit := e.ChildTexts(".item-product__info")
		if len(sizeSplit) == 0 {
			size = ""
		} else {
			size = sizeSplit[0]
		}
		product := &database.Product{
			Vendor:               "IGA",
			Brand:                brand,
			Price:                price,
			Name:                 e.ChildText(".js-ga-productname"),
			Image:                e.ChildAttr(".js-ga-productimage > img", "src"),
			Size:                 size,
			PricePerHundredGrams: e.ChildText(".item-product__info > div.text--small"),
		}

		/* No need to handle error here, unique constraint failures are expected */
		_, err = db.Insert(product)

		if err != nil {
			l.Error("Error executing database Insert", "error", err)
		}
	})

	scraper.OnRequest(func(r *colly.Request) {
		log.Println("visiting", r.URL.String())
	})

	return scraper
}
