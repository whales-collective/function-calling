package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"one-tool/cart"
	"one-tool/models"
	"one-tool/tools"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

func getOpenAIClient(chatURL string) openai.Client {
	client := openai.NewClient(
		option.WithBaseURL(chatURL),
		option.WithAPIKey(""),
	)
	return client
}

type Engine struct {
	ctx    context.Context
	client openai.Client
	model  string
	tools  []openai.ChatCompletionToolParam
	Params openai.ChatCompletionNewParams
}

func (e *Engine) Tools(tools []openai.ChatCompletionToolParam) {
	e.tools = tools
}

func (e *Engine) ToolCompletion(messages []openai.ChatCompletionMessageParamUnion) ([]openai.ChatCompletionMessageToolCall, error) {

	params := openai.ChatCompletionNewParams{
		Messages: messages,
		Model:    e.model,
		Tools:    e.tools,
		// Enable parallel tool calls for DMR, no need for this with Ollama
		ParallelToolCalls: openai.Bool(true),
		Temperature:       openai.Opt(0.0),
	}

	e.Params = params

	completion, err := e.client.Chat.Completions.New(e.ctx, params)
	if err != nil {
		return nil, fmt.Errorf("error creating tool completion: %w", err)
	}
	return completion.Choices[0].Message.ToolCalls, nil

}

func (e *Engine) ChatStreamCompletion(messages []openai.ChatCompletionMessageParamUnion, temperature float64, cbk func(content string)) {
	params := openai.ChatCompletionNewParams{
		Messages:    messages,
		Model:       e.model,
		Temperature: openai.Opt(temperature),
	}

	e.Params = params

	stream := e.client.Chat.Completions.NewStreaming(e.ctx, params)

	for stream.Next() {
		chunk := stream.Current()
		// Stream each chunk as it arrives
		if len(chunk.Choices) > 0 && chunk.Choices[0].Delta.Content != "" {
			//fmt.Print(chunk.Choices[0].Delta.Content)
			cbk(chunk.Choices[0].Delta.Content)
		}
	}
}

type EngineOption func(*Engine)

func NewEngine(options ...EngineOption) *Engine {
	engine := &Engine{}
	// Apply all options
	for _, option := range options {
		option(engine)
	}
	return engine
}

func WithDockerModelRunner(ctx context.Context) EngineOption {
	return func(engine *Engine) {
		engine.ctx = ctx
		engine.client = getOpenAIClient(os.Getenv("MODEL_RUNNER_BASE_URL"))
	}
}

func WithOllama(ctx context.Context) EngineOption {
	return func(engine *Engine) {
		engine.ctx = ctx
		engine.client = getOpenAIClient(os.Getenv("OLLAMA_BASE_URL"))
	}
}

func WithModel(model string) EngineOption {
	return func(engine *Engine) {
		engine.model = model
	}
}

func GetToolsCatalog() []openai.ChatCompletionToolParam {

	searchProducts := openai.ChatCompletionToolParam{
		Function: openai.FunctionDefinitionParam{
			Name:        "search_products",
			Description: openai.String("Search for products by query, category, or price range"),
			Parameters: openai.FunctionParameters{
				"type": "object",
				"properties": map[string]interface{}{
					"query": map[string]interface{}{
						"type":        "string",
						"description": "Search query for product name or description",
					},
					"category": map[string]interface{}{
						"type":        "string",
						"description": "Product category (electronics, clothing, books, home, sports, beauty, toys, food)",
					},
					"limit": map[string]interface{}{
						"type":        "integer",
						"description": "Maximum number of results to return (default: 10)",
					},
				},
			},
		},
	}

	addToCart := openai.ChatCompletionToolParam{
		Function: openai.FunctionDefinitionParam{
			Name:        "add_to_cart",
			Description: openai.String("Add a quantity of a product to the shopping cart"),
			Parameters: openai.FunctionParameters{
				"type": "object",
				"properties": map[string]interface{}{
					"product_name": map[string]interface{}{
						"type":        "string",
						"description": "The name of the product to add",
					},
					"quantity": map[string]interface{}{
						"type":        "integer",
						"description": "Quantity to add (default: 1)",
					},
				},
				"required": []string{"product_name"},
			},
		},
	}

	removeFromCart := openai.ChatCompletionToolParam{
		Function: openai.FunctionDefinitionParam{
			Name:        "remove_from_cart",
			Description: openai.String("Remove a product from the shopping cart"),
			Parameters: openai.FunctionParameters{
				"type": "object",
				"properties": map[string]interface{}{
					"product_name": map[string]interface{}{
						"type":        "string",
						"description": "The name of the product to remove",
					},
				},
				"required": []string{"product_name"},
			},
		},
	}

	viewCart := openai.ChatCompletionToolParam{
		Function: openai.FunctionDefinitionParam{
			Name:        "view_cart",
			Description: openai.String("View the current shopping cart contents and totals"),
			Parameters: openai.FunctionParameters{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
		},
	}

	updateQuantity := openai.ChatCompletionToolParam{
		Function: openai.FunctionDefinitionParam{
			Name:        "update_quantity",
			Description: openai.String("Update the quantity of a product in the cart"),
			Parameters: openai.FunctionParameters{
				"type": "object",
				"properties": map[string]interface{}{
					"product_name": map[string]interface{}{
						"type":        "string",
						"description": "The name of the product to update",
					},
					"quantity": map[string]interface{}{
						"type":        "integer",
						"description": "New quantity (use 0 to remove)",
					},
				},
				"required": []string{"product_name", "quantity"},
			},
		},
	}

	checkOut := openai.ChatCompletionToolParam{
		Function: openai.FunctionDefinitionParam{
			Name:        "checkout",
			Description: openai.String("Process checkout for the current cart"),
			Parameters: openai.FunctionParameters{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
		},
	}

	tools := []openai.ChatCompletionToolParam{
		searchProducts,
		addToCart,
		removeFromCart,
		viewCart,
		updateQuantity,
		checkOut,
	}
	return tools
}

func main() {
	ctx := context.Background()
	err := godotenv.Load()
	if err != nil {
		log.Fatalln("😡", err)
	}

	products, err := models.LoadProducts("products.json")
	if err != nil {
		log.Fatalln("😡", err)
	}
	// Create a new cart
	cart := cart.NewCart()

	dmrEngine := NewEngine(WithDockerModelRunner(ctx), WithModel(os.Getenv("MODEL_RUNNER_TOOL_LLM")))
	dmrChatEngine := NewEngine(WithDockerModelRunner(ctx), WithModel(os.Getenv("MODEL_RUNNER_CHAT_LLM")))

	dmrEngine.Tools(GetToolsCatalog())

	userQuestion := openai.UserMessage(`
		search the Dune book in books 
		search all books with a limit of 5 found books
		search all electronics with a limit of 3 found books

		add 3 iPad Pro 12.9 to the cart
		add 2 macbook air M3 to the cart
		add 5 Sapiens book to the cart

		view the cart

		remove iPad Pro 12.9 from the cart
		remove macbook air M3 and Sapiens book from the cart

		view the cart

		update the quantity of macbook air M3 to 1
		update the quantity of iPad Pro 12.9 to 2
		update the quantity of Sapiens book to 23

		view the cart

		checkout
	`)

	// No Sysystem message
	dmrToolCalls, err := dmrEngine.ToolCompletion(
		[]openai.ChatCompletionMessageParamUnion{
			userQuestion,
		},
	)
	if err != nil {
		log.Fatalln("😡", err)
	}

	// Return early if there are no tool calls
	if len(dmrToolCalls) == 0 {
		fmt.Println("😠 No function call")
		fmt.Println()
		return
	}

	// Display the tool calls
	for idx, toolCall := range dmrToolCalls {
		fmt.Println(idx,".", "🐳", toolCall.Function.Name, toolCall.Function.Arguments)

		switch toolCall.Function.Name {
		case "search_products":
			var args struct {
				Query    string `json:"query"`
				Category string `json:"category"`
				Limit    int    `json:"limit"`
			}
			err = json.Unmarshal([]byte(toolCall.Function.Arguments), &args)
			if err != nil {
				log.Fatalln("😡 Error unmarshalling search_products arguments:", err)
			}
			results := tools.SearchProducts(products, args.Query, args.Category, args.Limit)
			if len(results) == 0 {
				fmt.Println("😠 No products found for query:", args.Query, "category:", args.Category)
			} else {
				fmt.Println("✅ Found", len(results), "products:")
				content := fmt.Sprintf("Found %d products for query '%s' in category '%s':", len(results), args.Query, args.Category)
				for _, product := range results {
					fmt.Printf("  - %s (%s): $%.2f\n", product.Name, product.Category, product.Price)
					content += fmt.Sprintf("\n  - %s (%s): $%.2f", product.Name, product.Category, product.Price)
				}
				// Append the content to the messages
				dmrEngine.Params.Messages = append(dmrEngine.Params.Messages, openai.ToolMessage(
					content, toolCall.ID,
				))
			}

		case "add_to_cart":
			var args struct {
				ProductName string `json:"product_name"`
				Quantity    int    `json:"quantity"`
			}
			err = json.Unmarshal([]byte(toolCall.Function.Arguments), &args)
			if err != nil {
				log.Fatalln("😡 Error unmarshalling add_to_cart arguments:", err)
			}
			if args.Quantity <= 0 {
				fmt.Println("😠 Invalid quantity for adding to cart:", args.Quantity)
			} else {
				err := cart.AddToCart(products, args.ProductName, args.Quantity)
				if err != nil {
					fmt.Println("😠 Error adding to cart:", err)
				} else {
					fmt.Printf("✅ Added %d of '%s' to the cart\n", args.Quantity, args.ProductName)
					content := fmt.Sprintf("Added %d of '%s' to the cart", args.Quantity, args.ProductName)
					// Append the content to the messages
					dmrEngine.Params.Messages = append(dmrEngine.Params.Messages, openai.ToolMessage(
						content, toolCall.ID,
					))
				}
			}
		case "remove_from_cart":
			var args struct {
				ProductName string `json:"product_name"`
			}
			err = json.Unmarshal([]byte(toolCall.Function.Arguments), &args)
			if err != nil {
				log.Fatalln("😡 Error unmarshalling remove_from_cart arguments:", err)
			}
			if args.ProductName == "" {
				fmt.Println("😠 Invalid product name for removal")
			} else {
				err := cart.RemoveFromCart(products, args.ProductName, 1) // Default to removing 1 item
				if err != nil {
					fmt.Println("😠 Error removing from cart:", err)
				} else {
					fmt.Printf("✅ Removed '%s' from the cart\n", args.ProductName)
					content := fmt.Sprintf("Removed '%s' from the cart", args.ProductName)
					// Append the content to the messages
					dmrEngine.Params.Messages = append(dmrEngine.Params.Messages, openai.ToolMessage(
						content, toolCall.ID,
					))
				}
			}

		case "view_cart":
			fmt.Println("🛒 Viewing cart contents:")
			cart.DisplayCart()
			// Append the cart contents to the messages
			dmrEngine.Params.Messages = append(dmrEngine.Params.Messages, openai.ToolMessage(
				cart.PrintCart(), toolCall.ID,
			))

		case "update_quantity":
			var args struct {
				ProductName string `json:"product_name"`
				Quantity    int    `json:"quantity"`
			}
			err = json.Unmarshal([]byte(toolCall.Function.Arguments), &args)
			if err != nil {
				log.Fatalln("😡 Error unmarshalling update_quantity arguments:", err)
			}
			if args.Quantity < 0 {
				fmt.Println("😠 Invalid quantity for updating:", args.Quantity)
			} else {
				err := cart.UpdateCartQuantity(products, args.ProductName, args.Quantity)
				if err != nil {
					fmt.Println("😠 Error updating quantity:", err)
				} else {
					fmt.Printf("✅ Updated '%s' quantity to %d\n", args.ProductName, args.Quantity)
					content := fmt.Sprintf("Updated '%s' quantity to %d", args.ProductName, args.Quantity)
					// Append the content to the messages
					dmrEngine.Params.Messages = append(dmrEngine.Params.Messages, openai.ToolMessage(
						content, toolCall.ID,
					))
				}
			}
		case "checkout":
			fmt.Println("✅ Checkout completed successfully!")
			content := "Checkout completed successfully!"
			// Append the content to the messages
			dmrEngine.Params.Messages = append(dmrEngine.Params.Messages, openai.ToolMessage(
				content, toolCall.ID,
			))
		default:
			fmt.Println("😠 Unknown tool call:", toolCall.Function.Name)
			content := fmt.Sprintf("Unknown tool call: %s", toolCall.Function.Name)
			// Append the content to the messages
			dmrEngine.Params.Messages = append(dmrEngine.Params.Messages, openai.ToolMessage(
				content, toolCall.ID,
			))
		}
	} // End of tool calls loop

	messages := []openai.ChatCompletionMessageParamUnion{
		openai.SystemMessage(`You are a helpful assistant that can search products, manage a shopping cart`),
	}
	messages = append(messages, dmrEngine.Params.Messages...)
	messages = append(messages, openai.UserMessage(`
		Make a summary of the previous conversation and the actions taken.
		Include the total price of the cart and the number of items in it.
		Also, provide a list of all products that were added to the cart, removed, or updated.
		Make sure to include the final cart contents and the total price.
		You can use the following format for the summary:
		Cart Summary:
		- Total Items: <number of items>
		- Total Price: <total price>
		- Products Added: <list of products added>
		- Products Removed: <list of products removed>
		- Products Updated: <list of products updated>
		- Final Cart Contents: <list of products in the cart>
		Make sure to format the response in a way that is easy to read and understand.
	`))

	fmt.Println("\n" + strings.Repeat("=", 50))
	fmt.Println("🛠️  Using DMR Chat Engine for chat completion...")
	fmt.Println(strings.Repeat("=", 50))

	dmrChatEngine.ChatStreamCompletion(messages, 0.9, func(content string) {
		fmt.Print(content)
	})
	fmt.Println("\n" + strings.Repeat("=", 50))

}
