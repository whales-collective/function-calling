package cart

import (
	"fmt"
	"one-tool/models"
	"strings"
)

// CartItem represents an item in the shopping cart
type CartItem struct {
	Product  models.Product `json:"product"`
	Quantity int            `json:"quantity"`
}

// Cart represents a shopping cart
type Cart struct {
	Items []CartItem `json:"items"`
}

// NewCart creates a new empty cart
func NewCart() *Cart {
	return &Cart{
		Items: make([]CartItem, 0),
	}
}

// AddToCart adds a product to the cart by name and quantity
// Returns error if product not found or insufficient stock
func (c *Cart) AddToCart(products []models.Product, productName string, quantity int) error {
	if quantity <= 0 {
		return fmt.Errorf("quantity must be greater than 0")
	}

	// Find the product by name (case-insensitive exact match)
	var productIndex = -1
	var foundProduct models.Product

	for i, product := range products {
		if strings.EqualFold(product.Name, productName) {
			productIndex = i
			foundProduct = product
			break
		}
	}

	if productIndex == -1 {
		return fmt.Errorf("product '%s' not found", productName)
	}

	// Check if there's enough stock
	if foundProduct.Stock < quantity {
		return fmt.Errorf("insufficient stock for '%s'. Available: %d, Requested: %d",
			productName, foundProduct.Stock, quantity)
	}

	// Update stock in the products slice
	products[productIndex].Stock -= quantity

	// Check if product already exists in cart
	for i, item := range c.Items {
		if item.Product.ID == foundProduct.ID {
			// Update quantity if product already in cart
			c.Items[i].Quantity += quantity
			return nil
		}
	}

	// Add new item to cart
	cartItem := CartItem{
		Product:  foundProduct,
		Quantity: quantity,
	}
	c.Items = append(c.Items, cartItem)

	return nil
}

// UpdateCartQuantity updates the quantity of a product in the cart by name
// If newQuantity is 0, the item is removed from the cart
// Returns error if product not found in cart or insufficient stock
func (c *Cart) UpdateCartQuantity(products []models.Product, productName string, newQuantity int) error {
	if newQuantity < 0 {
		return fmt.Errorf("quantity cannot be negative")
	}

	// Find the item in cart
	cartItemIndex := -1
	for i, item := range c.Items {
		if strings.EqualFold(item.Product.Name, productName) {
			cartItemIndex = i
			break
		}
	}

	if cartItemIndex == -1 {
		return fmt.Errorf("product '%s' not found in cart", productName)
	}

	cartItem := c.Items[cartItemIndex]
	currentQuantity := cartItem.Quantity
	quantityDifference := newQuantity - currentQuantity

	// Find product in products slice to check/update stock
	productIndex := -1
	for i, product := range products {
		if product.ID == cartItem.Product.ID {
			productIndex = i
			break
		}
	}

	if productIndex == -1 {
		return fmt.Errorf("product '%s' not found in inventory", productName)
	}

	// If increasing quantity, check if enough stock available
	if quantityDifference > 0 {
		if products[productIndex].Stock < quantityDifference {
			return fmt.Errorf("insufficient stock for '%s'. Available: %d, Additional needed: %d", 
				productName, products[productIndex].Stock, quantityDifference)
		}
	}

	// Update stock in products slice
	products[productIndex].Stock -= quantityDifference

	// Update cart
	if newQuantity == 0 {
		// Remove item from cart
		c.Items = append(c.Items[:cartItemIndex], c.Items[cartItemIndex+1:]...)
	} else {
		// Update quantity
		c.Items[cartItemIndex].Quantity = newQuantity
	}

	return nil
}



// RemoveFromCart removes a product from the cart and restores stock
func (c *Cart) RemoveFromCart(products []models.Product, productName string, quantity int) error {
	if quantity <= 0 {
		return fmt.Errorf("quantity must be greater than 0")
	}

	// Find the item in cart
	cartItemIndex := -1
	for i, item := range c.Items {
		if strings.EqualFold(item.Product.Name, productName) {
			cartItemIndex = i
			break
		}
	}

	if cartItemIndex == -1 {
		return fmt.Errorf("product '%s' not found in cart", productName)
	}

	cartItem := c.Items[cartItemIndex]

	if cartItem.Quantity < quantity {
		return fmt.Errorf("cannot remove %d items. Only %d in cart", quantity, cartItem.Quantity)
	}

	// Find product in products slice to restore stock
	for i, product := range products {
		if product.ID == cartItem.Product.ID {
			products[i].Stock += quantity
			break
		}
	}

	// Update cart
	if cartItem.Quantity == quantity {
		// Remove item completely from cart
		c.Items = append(c.Items[:cartItemIndex], c.Items[cartItemIndex+1:]...)
	} else {
		// Reduce quantity
		c.Items[cartItemIndex].Quantity -= quantity
	}

	return nil
}

// GetCartTotal calculates the total price of items in the cart
func (c *Cart) GetCartTotal() float64 {
	total := 0.0
	for _, item := range c.Items {
		total += item.Product.Price * float64(item.Quantity)
	}
	return total
}

// GetCartItemCount returns the total number of items in the cart
func (c *Cart) GetCartItemCount() int {
	count := 0
	for _, item := range c.Items {
		count += item.Quantity
	}
	return count
}

// PrintCart returns a formatted string of the cart contents
func (c *Cart) PrintCart() string {
	if len(c.Items) == 0 {
		return "Cart is empty"
	}

	var result strings.Builder
	result.WriteString("Shopping Cart:\n")
	result.WriteString("==============\n")
	
	for _, item := range c.Items {
		total := item.Product.Price * float64(item.Quantity)
		result.WriteString(fmt.Sprintf("- %s x%d @ $%.2f each = $%.2f\n", 
			item.Product.Name, item.Quantity, item.Product.Price, total))
	}
	
	result.WriteString(fmt.Sprintf("Total Items: %d\n", c.GetCartItemCount()))
	result.WriteString(fmt.Sprintf("Total Price: $%.2f", c.GetCartTotal()))
	
	return result.String()
}

// DisplayCart prints the cart contents to console (convenience method)
func (c *Cart) DisplayCart() {
	fmt.Println(c.PrintCart())
}

// ClearCart empties the cart and restores all stock
func (c *Cart) ClearCart(products []models.Product) {
	// Restore stock for all items in cart
	for _, item := range c.Items {
		for i, product := range products {
			if product.ID == item.Product.ID {
				products[i].Stock += item.Quantity
				break
			}
		}
	}

	// Clear cart
	c.Items = make([]CartItem, 0)
}
