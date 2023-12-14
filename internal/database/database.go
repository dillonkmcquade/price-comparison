package database

import (
	"context"
	"database/sql"
	"log"
	"sync"
	"time"

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
	*sql.DB
	mut sync.Mutex
}

// Insert product to database, returns error from DB.Exec
func (db *Database) Insert(ctx context.Context, p *Product) (sql.Result, error) {
	context, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
	defer cancel()
	db.mut.Lock()
	result, err := db.ExecContext(context, `
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
func (db *Database) findCountByName(ctx context.Context, name string) (totalItems int, err error) {
	context, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	err = db.QueryRowContext(context, "SELECT count(*) from products where name like '%' || ? || '%'", name).Scan(&totalItems)
	return
}

// Returns the products in the database that contain the keyword
func (db *Database) FindByName(ctx context.Context, name string, page uint64) (*Result, error) {
	result := &Result{
		pageNumber:  page,
		SearchQuery: name,
	}

	// Get total count of products that match query
	totalItems, err := db.findCountByName(ctx, name)
	if err != nil {
		return result, err
	}
	result.TotalItems = totalItems

	context, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	// Find products that match query
	rows, err := db.QueryContext(context, `SELECT * FROM products WHERE name LIKE '%' || ? || '%' ORDER BY price ASC LIMIT 24 OFFSET ?`, name, page*24)
	if err != nil {
		return result, err
	}
	defer rows.Close()

	for rows.Next() {
		p := &Product{}
		e := rows.Scan(&p.Id, &p.Vendor, &p.Brand, &p.Name, &p.Price, &p.Image, &p.Size, &p.PricePerHundredGrams)
		if e != nil {
			log.Fatal(e)
		}
		result.Products = append(result.Products, p)
	}
	result.Count = len(result.Products)
	return result, err
}
