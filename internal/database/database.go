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
	Mut      sync.Mutex
	Products map[string]Product
}

func (db *Database) Insert(p *Product) error {
	id := fmt.Sprintf("%s-%s-%s-%f", p.Vendor, p.Brand, p.Name, p.Price)
	_, hasKey := db.Products[id]
	if !hasKey {
		db.Mut.Lock()
		db.Products[id] = *p
		defer db.Mut.Unlock()
		return nil
	} else {
		return errors.New("item exists in database already")
	}
}

func NewDatabase() *Database {
	return &Database{
		Products: map[string]Product{},
	}
}

func (db *Database) FindAll() []Product {
	var productsArray []Product
	for k, v := range db.Products {
		v.Id = k
		productsArray = append(productsArray, v)
	}

	return productsArray
}
