package scrapers

import (
	"log"
	"net/url"
	"strconv"
	"strings"

	"github.com/dillonkmcquade/price-comparison/internal/database"
	"github.com/gocolly/colly/v2"
)

const IGA_URL string = "https://www.iga.net/en/search"

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

	scraper.Collector.Limit(&colly.LimitRule{DomainGlob: "*", Parallelism: 2})

	scraper.Collector.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Attr("href")

		SetQuery(&scraper.Url, "page", "")
		prefix := strings.Join([]string{scraper.Url.Path, scraper.Url.Query().Encode()}, "?")
		if strings.HasPrefix(link, prefix) {
			e.Request.Visit(link)
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
		product := &database.Product{
			Vendor:               "IGA",
			Brand:                brand,
			Price:                price,
			Name:                 e.ChildText(".js-ga-productname"),
			Image:                e.ChildAttr(".js-ga-productimage > img", "src"),
			Size:                 e.ChildTexts(".item-product__info")[0],
			PricePerHundredGrams: e.ChildText(".item-product__info > div.text--small"),
		}

		/* No need to handle error here, unique constraint failures are expected */
		db.Insert(product)
	})

	scraper.Collector.OnRequest(func(r *colly.Request) {
		log.Println("visiting", r.URL.String())
	})

	return scraper
}
