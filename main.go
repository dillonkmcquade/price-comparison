package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"
	"sync"

	"github.com/gocolly/colly/v2"
)

type Product struct {
	Vendor               string `json:"vendor"`
	Brand                string `json:"brand"`
	Name                 string `json:"name"`
	Price                string `json:"price"`
	Image                string `json:"image"`
	Size                 string `json:"size"`
	PricePerHundredGrams string `json:"pricePerHundredGrams"`
}

type Database struct {
	mut      sync.Mutex
	products map[string]Product
}

const IGA_URL string = "https://www.iga.net/en/search"

func setQuery(u *url.URL, k string, v string) *url.URL {
	q := u.Query()
	q.Set(k, v)
	u.RawQuery = q.Encode()
	return u
}

func main() {
	fName := "products.json"
	file, err := os.Create(fName)
	if err != nil {
		log.Fatalf("Cannot create file %q: %s\n", fName, err)
		return
	}
	defer file.Close()

	// products := []Product{}
	db := Database{
		products: map[string]Product{},
	}

	igaUrl, err := url.Parse(IGA_URL)
	if err != nil {
		log.Fatal(err)
	}
	query := "carrots"
	setQuery(igaUrl, "k", query)

	collector := colly.NewCollector(
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
	collector.AllowURLRevisit = false

	collector.Limit(&colly.LimitRule{DomainGlob: "*", Parallelism: 2})

	collector.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Attr("href")

		setQuery(igaUrl, "page", "")
		prefix := strings.Join([]string{igaUrl.Path, igaUrl.Query().Encode()}, "?")
		if strings.HasPrefix(link, prefix) {
			e.Request.Visit(link)
		}
	})

	collector.OnHTML(".item-product.js-product", func(e *colly.HTMLElement) {
		log.Printf("Reading from product %d of page %s", e.Index, e.Request.URL.String())
		product := Product{
			Vendor:               "IGA",
			Brand:                e.ChildText(".item-product__brand"),
			Price:                e.ChildText("span.price"),
			Name:                 e.ChildText(".js-ga-productname"),
			Image:                e.ChildAttr(".js-ga-productimage > img", "src"),
			Size:                 e.ChildTexts(".item-product__info")[0],
			PricePerHundredGrams: e.ChildText(".item-product__info > div.text--small"),
		}
		id := fmt.Sprintf("%s-%s-%s", product.Brand, product.Name, product.Price)
		if _, hasKey := db.products[id]; !hasKey {
			db.mut.Lock()
			db.products[id] = product
			defer db.mut.Unlock()
		}
	})

	collector.OnRequest(func(r *colly.Request) {
		log.Println("visiting", r.URL.String())
	})

	err = collector.Visit(igaUrl.String())

	// Wait for goroutines to finish
	collector.Wait()
	if err != nil {
		log.Fatal(err)
	}

	e := json.NewEncoder(file)
	e.SetIndent("", "  ")
	err = e.Encode(db.products)
	if err != nil {
		log.Println(err)
	}
	log.Println("Finished")
}
