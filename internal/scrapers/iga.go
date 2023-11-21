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
func setQuery(u *url.URL, k string, v string) *url.URL {
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
func ScrapeIga(db *database.Database[database.Product], query string) *Scraper {
	igaUrl, err := url.Parse(IGA_URL)
	if err != nil {
		log.Fatal(err)
	}
	scraper := &Scraper{
		Url: *igaUrl,
	}
	setQuery(&scraper.Url, "k", query)

	scraper.Collector = colly.NewCollector(
		colly.AllowedDomains("www.iga.net", "iga.net"),
		// Cache responses to prevent multiple download of pages
		// even if the collector is restarted
		colly.CacheDir("./cache"),
		colly.MaxDepth(2),
		// Run requests in parallel
		colly.Async(),
	)
	scraper.Collector.AllowURLRevisit = false

	scraper.Collector.Limit(&colly.LimitRule{DomainGlob: "*", Parallelism: 2})

	scraper.Collector.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Attr("href")

		setQuery(&scraper.Url, "page", "")
		prefix := strings.Join([]string{scraper.Url.Path, scraper.Url.Query().Encode()}, "?")
		if strings.HasPrefix(link, prefix) {
			e.Request.Visit(link)
		}
	})

	scraper.Collector.OnHTML(".item-product.js-product", func(e *colly.HTMLElement) {
		log.Printf("Reading from product %d of page %s", e.Index, e.Request.URL.String())

		price, err := strToFloat(e.ChildTexts("span.price")[0])
		if err != nil {
			log.Println("error parsing float")
			return
		}

		product := &database.Product{
			Vendor:               "IGA",
			Brand:                e.ChildText(".item-product__brand"),
			Price:                price,
			Name:                 e.ChildText(".js-ga-productname"),
			Image:                e.ChildAttr(".js-ga-productimage > img", "src"),
			Size:                 e.ChildTexts(".item-product__info")[0],
			PricePerHundredGrams: e.ChildText(".item-product__info > div.text--small"),
		}
		err = db.Insert(product)
		if err != nil {
			log.Println(err)
		}
	})

	scraper.Collector.OnRequest(func(r *colly.Request) {
		log.Println("visiting", r.URL.String())
	})

	return scraper
}
