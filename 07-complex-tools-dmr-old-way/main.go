package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
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
		Seed:              openai.Int(0),
		Temperature:       openai.Opt(0.0),
	}

	e.Params = params

	completion, err := e.client.Chat.Completions.New(e.ctx, params)
	if err != nil {
		return nil, fmt.Errorf("error creating tool completion: %w", err)
	}
	return completion.Choices[0].Message.ToolCalls, nil

}

func (e *Engine) ChatCompletion(messages []openai.ChatCompletionMessageParamUnion, temperature float64) (string, error) {

	param := openai.ChatCompletionNewParams{
		Messages:    messages,
		Model:       e.model,
		Temperature: openai.Opt(temperature),
	}

	completion, err := e.client.Chat.Completions.New(e.ctx, param)

	if err != nil {
		return "", fmt.Errorf("error creating chat completion: %w", err)
	}

	return completion.Choices[0].Message.Content, nil
}

func (e *Engine) JSONCompletion(messages []openai.ChatCompletionMessageParamUnion) (string, error) {

	schema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"name": map[string]any{
				"type": "string",
			},
			"arguments": map[string]any{
				"type": "object",
			},
		},
		"required": []string{"name", "capital", "languages"},
	}

	schemaParam := openai.ResponseFormatJSONSchemaJSONSchemaParam{
		Name:        "country_info",
		Description: openai.String("Notable information about a country in the world"),
		Schema:      schema,
		Strict:      openai.Bool(true),
	}

	param := openai.ChatCompletionNewParams{
		Messages:    messages,
		Model:       e.model,
		Temperature: openai.Opt(0.0),
		ResponseFormat: openai.ChatCompletionNewParamsResponseFormatUnion{
			OfJSONSchema: &openai.ResponseFormatJSONSchemaParam{
				JSONSchema: schemaParam,
			},
		},
	}

	completion, err := e.client.Chat.Completions.New(e.ctx, param)

	if err != nil {
		return "", fmt.Errorf("error creating chat completion: %w", err)
	}

	return completion.Choices[0].Message.Content, nil
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
		//log.Fatalln("😡", err)
		// use the env variables from compose file if not found
	}

	llmToolEngine := NewEngine(WithDockerModelRunner(ctx), WithModel(os.Getenv("MODEL_RUNNER_TOOL_LLM")))

	fmt.Println("\n" + strings.Repeat("=", 50))
	fmt.Println("🛠️  Tools completion...")
	fmt.Println(strings.Repeat("=", 50))

	llmToolEngine.Tools(GetToolsCatalog())

	//llmToolEngine.Tools(GetToolsCatalog())

	jsonData, err := json.MarshalIndent(GetToolsCatalog(), "", "  ")
	if err != nil {
		log.Fatalln("😡 Error marshalling tools catalog:", err)
	}
	fmt.Println("🛠️  Tools catalog:", string(jsonData))

	systemPrompt := openai.SystemMessage(fmt.Sprintf(`You have access to the following tools
	[AVAILABLE_TOOLS]
		%s
	[/AVAILABLE_TOOLS]

	Search in the user question all the tool calls that match with the description of the tools.

	For each tool call, respond with a JSON object with the following structure: 
	{
	  "name": <name of the called tool>,
	  "arguments": {
	    <name of the argument>: <value of the argument>
	  }
	}
	
	search the name of the tool in the list of tools with the Name field.
	Do this for every tool call you make.
	You can call multiple tools in a single response, but you must return a JSON array with the tool calls.
	
	`, string(jsonData)))

	userQuestion := openai.UserMessage(`
		search the Dune book in books 
		search all books with a limit of 5 found books
		search all electronics with a limit of 3 found books

		add 3 iPad Pro 12.9 to the cart
		add 2 macbook air M3 to the cart
		add 5 Sapiens book to the cart and 2 Dune book to the cart
		remove iPad Pro 12.9 from the cart

		view the cart
	`)

	dmrToolCallsStr, err := llmToolEngine.JSONCompletion(
		[]openai.ChatCompletionMessageParamUnion{
			systemPrompt,
			userQuestion,
		})

	if err != nil {
		log.Fatalln("😡", err)
	}
	fmt.Println("🛠️  Tool calls:", dmrToolCallsStr)

}
