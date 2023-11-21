package database

import (
	"errors"
	"fmt"
	"sync"
)

type Product struct {
	Id                   string  `json:"id"`
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
	products map[string]Product
}

/* This might be obsolete if a database is used */
func createId(p *Product) (string, error) {
	return fmt.Sprintf("%s-%s-%s-%f", p.Vendor, p.Brand, p.Name, p.Price), nil
}

// Insert product to database, returns error if an item already exists
func (db *Database) Insert(p *Product) error {
	id, err := createId(p)
	if err != nil {
		return err
	}
	_, hasKey := db.products[id]
	if !hasKey {
		db.mut.Lock()
		db.products[id] = *p
		defer db.mut.Unlock()
		return nil
	} else {
		return errors.New("item exists in database already")
	}
}

func NewDatabase() *Database {
	return &Database{
		products: map[string]Product{},
	}
}

// Returns all items as an array
func (db *Database) FindAll() []Product {
	var productsArray []Product
	for k, value := range db.products {
		value.Id = k
		productsArray = append(productsArray, value)
	}
	return productsArray
}

func (db *Database) FindOne(id string) (Product, error) {
	product, hasKey := db.products[id]
	if hasKey {
		return product, nil
	} else {
		return product, errors.New("item does not exist")
	}
}
