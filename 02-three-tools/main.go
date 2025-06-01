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

	vulcanSaluteTool := openai.ChatCompletionToolParam{
		Function: openai.FunctionDefinitionParam{
			Name:        "vulcan_salute",
			Description: openai.String("Give a vulcan salute to the given person name"),
			Parameters: openai.FunctionParameters{
				"type": "object",
				"properties": map[string]interface{}{
					"name": map[string]string{
						"type": "string",
					},
				},
				"required": []string{"name"},
			},
		},
	}

	sayHelloTool := openai.ChatCompletionToolParam{
		Function: openai.FunctionDefinitionParam{
			Name:        "say_hello",
			Description: openai.String("Say hello to the given person name"),
			Parameters: openai.FunctionParameters{
				"type": "object",
				"properties": map[string]interface{}{
					"name": map[string]string{
						"type": "string",
					},
				},
				"required": []string{"name"},
			},
		},
	}

	additionTool := openai.ChatCompletionToolParam{
		Function: openai.FunctionDefinitionParam{
			Name:        "addition",
			Description: openai.String("Add two numbers together"),
			Parameters: openai.FunctionParameters{
				"type": "object",
				"properties": map[string]interface{}{
					"number1": map[string]string{
						"type": "number",
					},
					"number2": map[string]string{
						"type": "number",
					},
				},
				"required": []string{"number1", "number2"},
			},
		},
	}

	tools := []openai.ChatCompletionToolParam{
		vulcanSaluteTool, sayHelloTool, additionTool,
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
		Make a Vulcan salute to Spock
		Say Hello to John Doe
		Add 10 and 32
		Make a Vulcan salute to Bob Morane
		Say Hello to Jane Doe
		Add 5 and 37
		Make a Vulcan salute to Sam
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
