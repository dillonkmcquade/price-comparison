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

type Result struct {
	TotalProducts int        // The total amount of products that match the keyword
	RowCount      int        // The amount of rows returned by this query
	Products      []*Product // A slice of *Product
}

type Database struct {
	*sql.DB
	mut sync.Mutex
}

// Insert product to database, returns error from DB.Exec
func (db *Database) Insert(p *Product) (sql.Result, error) {
	db.mut.Lock()
	result, err := db.Exec(`
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
		DB: db,
	}
}

// Returns the total row count in the database that contains the keyword
func (db *Database) findCountByName(name string) (totalItems int, err error) {
	err = db.QueryRow("SELECT count(*) from products where name like '%' || ? || '%'", name).Scan(&totalItems)
	return
}

// Returns the products in the database that contain the keyword
func (db *Database) FindByName(name string, page int) (*Result, error) {
	result := &Result{}

	// Get total count of products that match query
	totalItems, err := db.findCountByName(name)
	if err != nil {
		return result, err
	}
	result.TotalProducts = totalItems

	// Find products that match query
	rows, err := db.Query(`SELECT * FROM products WHERE name LIKE '%' || ? || '%' ORDER BY price ASC LIMIT 24 OFFSET ?`, name, page*24)
	if err != nil {
		return result, err
	}
	defer rows.Close()

	for rows.Next() {
		p := &Product{}
		err := rows.Scan(&p.Id, &p.Vendor, &p.Brand, &p.Name, &p.Price, &p.Image, &p.Size, &p.PricePerHundredGrams)
		if err != nil {
			log.Fatal(err)
		}
		result.Products = append(result.Products, p)
	}
	result.RowCount = len(result.Products)
	return result, err
}
