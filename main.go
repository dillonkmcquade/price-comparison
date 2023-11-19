package main

import (
	"encoding/json"
	"fmt"
	"log"
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
	mut      *sync.Mutex
	products map[string]Product
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
		mut:      &sync.Mutex{},
		products: map[string]Product{},
	}

	c := colly.NewCollector(
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
	c.AllowURLRevisit = false

	c.Limit(&colly.LimitRule{DomainGlob: "*", Parallelism: 2})

	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Attr("href")
		if strings.HasPrefix(link, "/en/search?k=carrots&page=") {
			e.Request.Visit(link)
		}
	})

	c.OnHTML(".item-product.js-product", func(e *colly.HTMLElement) {
		log.Printf("Reading from product %d of page %s", e.Index, e.Request.URL)
		product := Product{
			Vendor:               "IGA",
			Brand:                e.ChildText(".item-product__brand"),
			Price:                e.ChildText("span.price"),
			Name:                 e.ChildText(".js-ga-productname"),
			Image:                e.ChildAttr(".js-ga-productimage > img", "src"),
			Size:                 e.ChildTexts(".item-product__info")[0],
			PricePerHundredGrams: e.ChildText(".item-product__info > div.text--small"),
		}
		id := fmt.Sprintf("%s%s%s", product.Brand, product.Name, product.Price)
		if _, ok := db.products[id]; !ok {
			db.mut.Lock()
			db.products[id] = product
			defer db.mut.Unlock()
		}
	})

	c.OnRequest(func(r *colly.Request) {
		log.Println("visiting", r.URL.String())
	})

	err = c.Visit("https://www.iga.net/en/search?k=carrots")

	// Wait for goroutines to finish
	c.Wait()

	if err != nil {
		log.Println(err)
	}

	e := json.NewEncoder(file)

	e.SetIndent("", "  ")

	err = e.Encode(db.products)
	if err != nil {
		log.Println(err)
	}

	log.Println("Success")
}
