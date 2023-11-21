package database

import (
	"bytes"
	"encoding/json"
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

type Database[T any] struct {
	mut      sync.Mutex
	products map[string]T
}

/* This might be obsolete if a database is used */
func createId[T any](product *T) (string, error) {
	buf := new(bytes.Buffer)
	json.NewEncoder(buf).Encode(product)
	fmt.Println(buf.String())
	v, ok := any(product).(Product)
	if ok {
		return fmt.Sprintf("%s-%s-%s-%f", v.Vendor, v.Brand, v.Name, v.Price), nil
	} else {
		return "", errors.New("value not a valid product")
	}
}

// Insert product to database, returns error if an item already exists
func (db *Database[T]) Insert(p *T) error {
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

func NewDatabase[T any]() *Database[T] {
	return &Database[T]{
		products: map[string]T{},
	}
}

// Returns all items as an array
func (db *Database[T]) FindAll() []T {
	var productsArray []T
	for _, value := range db.products {
		productsArray = append(productsArray, value)
	}
	return productsArray
}

func (db *Database[T]) FindOne(id string) (T, error) {
	product, hasKey := db.products[id]
	if hasKey {
		return product, nil
	} else {
		return product, errors.New("item does not exist")
	}
}
