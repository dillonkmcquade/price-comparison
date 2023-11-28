package database

import (
	"database/sql"
	"fmt"
	"log"
	"math"
	"sync"

	_ "github.com/mattn/go-sqlite3"
)

type Product struct {
	Id                   int     `json:"id"`
	Vendor               string  `json:"vendor"`
	Brand                string  `json:"brand"`
	Name                 string  `json:"name"`
	Price                float64 `json:"price"`
	Image                string  `json:"image"`
	Size                 string  `json:"size"`
	PricePerHundredGrams string  `json:"pricePerHundredGrams"`
}

type Database struct {
	mut      sync.Mutex
	products *sql.DB
}

// Insert product to database, returns error from DB.Exec
func (db *Database) Insert(p *Product) (sql.Result, error) {
	db.mut.Lock()
	result, err := db.products.Exec(`
        INSERT INTO products 
            (vendor, brand, name, price, image, size, price_per_hundred_grams)
        VALUES 
            (?, ?, ?, ?, ?, ?, ?)`, p.Vendor, p.Brand, p.Name, p.Price, p.Image, p.Size, p.PricePerHundredGrams)
	db.mut.Unlock()
	return result, err
}

func NewDatabase(dataSourceName string) *Database {
	db, err := sql.Open("sqlite3", dataSourceName)
	if err != nil {
		log.Fatal(err)
	}
	_, err = db.Exec(`CREATE TABLE products (
        id INTEGER PRIMARY KEY,
        vendor VARCHAR(5),
        brand CHARACTER VARYING,
        name CHARACTER VARYING,
        price DOUBLE PRECISION,
        image TEXT,
        size CHARACTER VARYING,
        price_per_hundred_grams CHARACTER VARYING,
        UNIQUE(vendor, brand, name, size)
        )`)
	if err != nil {
		log.Fatal(err)
	}
	return &Database{
		products: db,
	}
}

func (db *Database) FindAll() ([]*Product, error) {
	var productsArray []*Product
	rows, err := db.products.Query("SELECT * FROM products")
	if err != nil {
		rows.Close()
		return productsArray, err
	}
	defer rows.Close()
	for rows.Next() {
		p := &Product{}
		err := rows.Scan(&p.Id, &p.Vendor, &p.Brand, &p.Name, &p.Price, &p.Image, &p.Size, &p.PricePerHundredGrams)
		if err != nil {
			log.Fatal(err)
		}
		productsArray = append(productsArray, p)
	}
	return productsArray, err
}

func (db *Database) FindById(id string) (*Product, error) {
	var product *Product
	err := db.products.QueryRow("SELECT * FROM products WHERE id = ?", id).Scan(&product)
	return product, err
}

type Page struct {
	Page       int        `json:"page"`
	TotalPages int        `json:"totalPages"`
	LastPage   string     `json:"lastPage"`
	NextPage   string     `json:"nextPage"`
	Count      int        `json:"count"`
	Products   []*Product `json:"products"`
}

func (db *Database) FindByName(name string, page int) (*Page, error) {
	paginate := &Page{
		Page: page,
	}
	paginate.NextPage = fmt.Sprintf("http://localhost:3001/products?search=%s&page=%d", name, page+1)
	paginate.LastPage = fmt.Sprintf("http://localhost:3001/products?search=%s&page=%d", name, page-1)
	if page == 0 {
		paginate.LastPage = ""
	}

	var totalItems int
	err := db.products.QueryRow("SELECT count(*) from products where name like '%' || ? || '%'", name).Scan(&totalItems)
	rows, err := db.products.Query(`SELECT * FROM products WHERE name LIKE '%' || ? || '%' ORDER BY price ASC LIMIT 25 OFFSET ?`, name, page*25)
	if err != nil {
		rows.Close()
		return paginate, err
	}
	defer rows.Close()
	for rows.Next() {
		p := &Product{}
		err := rows.Scan(&p.Id, &p.Vendor, &p.Brand, &p.Name, &p.Price, &p.Image, &p.Size, &p.PricePerHundredGrams)
		if err != nil {
			log.Fatal(err)
		}
		paginate.Products = append(paginate.Products, p)
	}

	paginate.Count = len(paginate.Products)
	paginate.TotalPages = int(math.Ceil(float64(totalItems) / 25.00))
	if paginate.Count < 25 {
		paginate.NextPage = ""
	}

	return paginate, err
}
