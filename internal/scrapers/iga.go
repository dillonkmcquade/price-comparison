package scrapers

import (
	"log"
	"net/url"
	"strings"

	"github.com/dillonkmcquade/price-comparison/internal/database"
	"github.com/gocolly/colly/v2"
)

const IGA_URL string = "https://www.iga.net/en/search"

// Scrapes IGA and adds items to the db
func NewIgaScraper(db *database.Database, query string) *Scraper {
	igaUrl, err := url.Parse(IGA_URL)
	if err != nil {
		log.Fatal(err)
	}
	scraper := &Scraper{
		Url: *igaUrl,
	}
	SetQuery(&scraper.Url, "k", query)

	scraper.Collector = colly.NewCollector(
		colly.AllowedDomains("www.iga.net", "iga.net"),
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

		SetQuery(&scraper.Url, "page", "")
		prefix := strings.Join([]string{scraper.Url.Path, scraper.Url.Query().Encode()}, "?")
		if strings.HasPrefix(link, prefix) {
			err = e.Request.Visit(link)
			if err != nil && err != colly.ErrMaxDepth {
				log.Fatal(err)
			}
		}
	})

	scraper.Collector.OnHTML(".item-product.js-product", func(e *colly.HTMLElement) {
		price, err := strToFloat(e.ChildTexts("span.price")[0])
		if err != nil {
			log.Println("error parsing float")
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
			log.Println(err)
		}
	})

	scraper.Collector.OnRequest(func(r *colly.Request) {
		log.Println("visiting", r.URL.String())
	})

	return scraper
}
