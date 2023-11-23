package database

import (
	"database/sql"
	"log"
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
func (db *Database) Insert(p *Product) error {
	db.mut.Lock()
	_, err := db.products.Exec(`
        INSERT INTO products 
            (vendor, brand, name, price, image, size, price_per_hundred_grams)
        VALUES 
            (?, ?, ?, ?, ?, ?, ?)`, p.Vendor, p.Brand, p.Name, p.Price, p.Image, p.Size, p.PricePerHundredGrams)
	db.mut.Unlock()
	return err
}

func NewDatabase() *Database {
	db, err := sql.Open("sqlite3", "file::memory:?cache=shared")
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

func (db *Database) FindOne(id string) (*Product, error) {
	var product *Product
	err := db.products.QueryRow("SELECT * FROM products WHERE id = ?", id).Scan(&product)
	return product, err
}
