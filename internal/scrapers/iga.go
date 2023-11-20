package scrapers

import (
	"log"
	"net/url"
	"strings"

	"github.com/dillonkmcquade/price-comparison/internal/database"
	"github.com/gocolly/colly/v2"
)

const IGA_URL string = "https://www.iga.net/en/search"

func setQuery(u *url.URL, k string, v string) *url.URL {
	q := u.Query()
	q.Set(k, v)
	u.RawQuery = q.Encode()
	return u
}

// Scrapes IGA and adds items to the db
func ScrapeIga(db *database.Database, query string) *Scraper {
	igaUrl, err := url.Parse(IGA_URL)
	if err != nil {
		log.Fatal(err)
	}
	scraper := &Scraper{
		Url: *igaUrl,
	}
	setQuery(igaUrl, "k", query)

	scraper.Collector = colly.NewCollector(
		// Visit only domains: coursera.org, www.coursera.org
		colly.AllowedDomains("www.iga.net", "iga.net"),

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

		setQuery(igaUrl, "page", "")
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
