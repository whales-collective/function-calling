package models

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

type Product struct {
	ID          string  `json:"ID"`
	Name        string  `json:"Name"`
	Description string  `json:"Description"`
	Price       float64 `json:"Price"`
	Category    string  `json:"Category"`
	Stock       int     `json:"Stock"`
}

type ProductCatalog struct {
	Products []Product `json:"products"`
}

func LoadProducts(filename string) ([]Product, error) {
	// Read the JSON file
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("error reading file: %v", err)
	}

	// Parse JSON into ProductCatalog struct
	var catalog ProductCatalog
	err = json.Unmarshal(data, &catalog)
	if err != nil {
		return nil, fmt.Errorf("error parsing JSON: %v", err)
	}

	return catalog.Products, nil
}
