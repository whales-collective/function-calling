package tools

import (
	"one-tool/models"
	"strings"
)

// SearchProducts searches for products by name and/or category with optional limit
// name: partial or full product name (case-insensitive), empty string to ignore
// category: exact category match (case-insensitive), empty string to ignore
// limit: maximum number of results to return, 0 or negative for no limit
func SearchProducts(products []models.Product, name, category string, limit int) []models.Product {
	var results []models.Product

	// Convert search terms to lowercase for case-insensitive comparison
	searchName := strings.ToLower(name)
	searchCategory := strings.ToLower(category)

	for _, product := range products {
		// Check if we've reached the limit
		if limit > 0 && len(results) >= limit {
			break
		}

		productName := strings.ToLower(product.Name)
		productCategory := strings.ToLower(product.Category)

		// Check name match (partial match if name is provided)
		nameMatch := name == "" || strings.Contains(productName, searchName)

		// Check category match (exact match if category is provided)
		categoryMatch := category == "" || productCategory == searchCategory

		// Include product if both conditions are met
		if nameMatch && categoryMatch {
			results = append(results, product)
		}
	}

	return results
}

// SearchProductsByNameOnly searches products by name only (partial match) with optional limit
func SearchProductsByNameOnly(products []models.Product, name string, limit int) []models.Product {
	return SearchProducts(products, name, "", limit)
}

// SearchProductsByCategory searches products by category only (exact match) with optional limit
func SearchProductsByCategory(products []models.Product, category string, limit int) []models.Product {
	return SearchProducts(products, "", category, limit)
}

// Legacy functions without limit (for backward compatibility)
func SearchProductsUnlimited(products []models.Product, name, category string) []models.Product {
	return SearchProducts(products, name, category, 0)
}

func SearchProductsByNameOnlyUnlimited(products []models.Product, name string) []models.Product {
	return SearchProducts(products, name, "", 0)
}

func SearchProductsByCategoryUnlimited(products []models.Product, category string) []models.Product {
	return SearchProducts(products, "", category, 0)
}
