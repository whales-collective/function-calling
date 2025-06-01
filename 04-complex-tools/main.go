package main

import (
	"context"
	"fmt"
	"log"
	"os"

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

	completion, err := e.client.Chat.Completions.New(e.ctx, params)
	if err != nil {
		return nil, fmt.Errorf("error creating tool completion: %w", err)
	}
	return completion.Choices[0].Message.ToolCalls, nil

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
			Description: openai.String("Add a product to the shopping cart"),
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

	dmrEngine := NewEngine(WithDockerModelRunner(ctx), WithModel(os.Getenv("MODEL_RUNNER_LLM")))
	ollamaEngine := NewEngine(WithOllama(ctx), WithModel(os.Getenv("OLLAMA_LLM")))

	dmrEngine.Tools(GetToolsCatalog())
	ollamaEngine.Tools(GetToolsCatalog())

	userQuestion := openai.UserMessage(`
		search the Dune book in books 
		search all fom books with a limit of 5 found books
		search all books with a range price betwwen 10 and 20

		add 3 ipad pro to the cart
		add 2 macbook pro to the cart
		add Sapiens book to the cart

		remove ipad pro from the cart
		remove  macbook pro and Sapiens book from the cart

		view the cart

		update the quantity of macbook pro to 1
		update the quantity of ipad pro to 0
		update the quantity of Sapiens book to 23

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

	// No Sysystem message
	ollamaToolCalls, err := ollamaEngine.ToolCompletion(
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

	if len(ollamaToolCalls) == 0 {
		fmt.Println("😠 No function call")
		fmt.Println()
		return
	}

	// Display the tool calls
	for _, toolCall := range dmrToolCalls {
		fmt.Println("🐳", toolCall.Function.Name, toolCall.Function.Arguments)
	}

	for _, toolCall := range ollamaToolCalls {
		fmt.Println("🦙", toolCall.Function.Name, toolCall.Function.Arguments)
	}

	// Check if both tool calls are equal and have the same number of items
	if len(dmrToolCalls) != len(ollamaToolCalls) {
		fmt.Println("❌ Tool calls do not have the same number of items", len(dmrToolCalls), "vs", len(ollamaToolCalls))
		return
	}

	for i := range dmrToolCalls {
		if dmrToolCalls[i].Function.Name != ollamaToolCalls[i].Function.Name {
			fmt.Println("😠 Tool calls are not equal")
			return
		}
		if dmrToolCalls[i].Function.Arguments != ollamaToolCalls[i].Function.Arguments {
			fmt.Println("😠 Tool calls are not equal")
			return
		}
	}

	fmt.Println("✅ Tool calls are equal and have exactly the same number of items:", len(dmrToolCalls))

}
