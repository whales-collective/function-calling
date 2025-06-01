# Function Calling and 🐥 Tiny Models
> Experiments 🧪

## Test 1: one simple tool, several calls

**Source code**: `01-one-tool`

**Engines**: 🐳 Docker Model Runner, 🦙 Ollama

**Models**:
- 🐳 Docker Model Runner: `ai/qwen2.5:0.5B-F16`
- 🦙 Ollama: `qwen2.5:0.5b`

**Tool**:
```golang
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
```

**Prompt**:
> No system instruction
```golang
userQuestion := openai.UserMessage(`
    Make a Vulcan salute to Spock
    Make a Vulcan salute to Bob Morane
    Make a Vulcan salute to Sam

    Make a Vulcan salute to John Doe
    Make a Vulcan salute to Jane Doe
    Make a Vulcan salute to Bill Gates
`)
```

**Results**:

```raw
🐳 vulcan_salute {"name":"Spock"}
🐳 vulcan_salute {"name":"Bob Morane"}
🐳 vulcan_salute {"name":"Sam"}
🐳 vulcan_salute {"name":"John Doe"}
🐳 vulcan_salute {"name":"Jane Doe"}
🐳 vulcan_salute {"name":"Bill Gates"}
🦙 vulcan_salute {"name":"Spock"}
🦙 vulcan_salute {"name":"Bob Morane"}
🦙 vulcan_salute {"name":"Sam"}
🦙 vulcan_salute {"name":"John Doe"}
🦙 vulcan_salute {"name":"Jane Doe"}
🦙 vulcan_salute {"name":"Bill Gates"}
✅ Tool calls are equal and have exactly the same number of items: 6
```
**Result**: Ollama == Docker Model Runner

## Test 2: three simple tools, several calls

**Source code**: `02-three-tools`

**Engines**: 🐳 Docker Model Runner, 🦙 Ollama

**Models**:
- 🐳 Docker Model Runner: `ai/qwen2.5:0.5B-F16`
- 🦙 Ollama: `qwen2.5:0.5b`

**Tool**:
```golang
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
```

**Prompt**:
> No system instruction
```golang
userQuestion := openai.UserMessage(`
    Make a Vulcan salute to Spock
    Say Hello to John Doe
    Add 10 and 32
    Make a Vulcan salute to Bob Morane
    Say Hello to Jane Doe
    Add 5 and 37
    Make a Vulcan salute to Sam
`)
```

**Results**:

```raw
🐳 vulcan_salute {"name":"Spock"}
🐳 say_hello {"name":"John Doe"}
🐳 addition {"number1":10,"number2":32}
🐳 say_hello {"name":"Jane Doe"}
🐳 addition {"number1":5,"number2":37}
🦙 vulcan_salute {"name":"Spock"}
🦙 say_hello {"name":"John Doe"}
🦙 addition {"number1":10,"number2":32}
🦙 vulcan_salute {"name":"Bob Morane"}
🦙 say_hello {"name":"Jane Doe"}
🦙 addition {"number1":5,"number2":37}
❌ Tool calls do not have the same number of items 5 vs 6
```
**Result**: Ollama wins


## Test 3: three simple tools, several calls
> ✋ bigger models: **`1.5b`**

**Source code**: `03-three-tools`

**Engines**: 🐳 Docker Model Runner, 🦙 Ollama

**Models**:
- 🐳 Docker Model Runner: `ai/qwen2.5:1.5B-F16`
- 🦙 Ollama: `qwen2.5:1.5b`

**Tool**:
```golang
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
```

**Prompt**:
> No system instruction
```golang
userQuestion := openai.UserMessage(`
    Make a Vulcan salute to Spock
    Say Hello to John Doe
    Add 10 and 32
    Make a Vulcan salute to Bob Morane
    Say Hello to Jane Doe
    Add 5 and 37
    Make a Vulcan salute to Sam
`)
```

**Results**:

```raw
🐳 vulcan_salute {"name":"Spock"}
🐳 say_hello {"name":"John Doe"}
🐳 addition {"number1":10,"number2":32}
🐳 vulcan_salute {"name":"Bob Morane"}
🐳 say_hello {"name":"Jane Doe"}
🐳 addition {"number1":5,"number2":37}
🦙 vulcan_salute {"name":"Spock"}
🦙 say_hello {"name":"John Doe"}
🦙 addition {"number1":10,"number2":32}
🦙 vulcan_salute {"name":"Bob Morane"}
🦙 say_hello {"name":"Jane Doe"}
🦙 addition {"number1":5,"number2":37}
✅ Tool calls are equal and have exactly the same number of items: 6
```
**Result**: Ollama == Docker Model Runner


## Test 4: the tools of chat2cart
> ✋ bigger models: **`1.5b`**

**Source code**: `04-complex-tools`

**Engines**: 🐳 Docker Model Runner, 🦙 Ollama

**Models**:
- 🐳 Docker Model Runner: `ai/qwen2.5:1.5B-F16`
- 🦙 Ollama: `qwen2.5:1.5b`

**Tool**: the tools of chat2cart

**Prompt**:
> No system instruction
```golang
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
```

**Results**:

```raw
🐳 search_products {"category":"books","query":"Dune"}
🐳 search_products {"category":"books","limit":5,"query":""}
🐳 search_products {"category":"books","query":"10-20"}
🐳 add_to_cart {"product_name":"iPad Pro"}
🐳 add_to_cart {"product_name":"MacBook Pro"}
🐳 add_to_cart {"product_name":"Sapiens"}
🐳 remove_from_cart {"product_name":"iPad Pro"}
🐳 remove_from_cart {"product_name":"MacBook Pro"}
🐳 remove_from_cart {"product_name":"Sapiens"}
🐳 view_cart {}
🐳 update_quantity {"product_name":"MacBook Pro","quantity":1}
🐳 update_quantity {"product_name":"iPad Pro","quantity":0}
🐳 update_quantity {"product_name":"Sapiens","quantity":23}
🐳 checkout {}
🦙 search_products {"category":"books","limit":5,"query":""}
🦙 search_products {"category":"books","limit":null,"query":"10-20"}
🦙 add_to_cart {"product_name":"iPad Pro"}
🦙 add_to_cart {"product_name":"MacBook Pro"}
🦙 add_to_cart {"product_name":"Sapiens Book"}
❌ Tool calls do not have the same number of items 14 vs 5
```
**Result**: Docker Model Runner is a lot better 🎉

## How to improve the results?

I think that for each user message, we need to execute 2 completions and not only one:
- a **tools** completion (with a model that has tools support) with temperature set to zero, and using `ParallelToolCalls: openai.Bool(true)` (then you can avoid looping over the messages to execute the tools one by one: `for currentIteration < maxIterations`)
- a **classic chat** completion (this can be done with another model, more "gifted" for formatting), with the **same prompt + the execution results**, and there we can raise the temperature again

✋ Using `ParallelToolCalls: openai.Bool(true)` allows to execute all the tools in parallel (no need to loop over the messages).

**First completion**: a tool completion

```golang
// Create a list of messages for the completion request
messages := []openai.ChatCompletionMessageParamUnion{
    openai.SystemMessage(systemInstructions),
    openai.UserMessage(userQuestion),
}

// Create the tool completion parameters
params := openai.ChatCompletionNewParams{
    Messages:          messages,
    ParallelToolCalls: openai.Bool(true),   // <-- this is a tool completion (no need to loop over the messages)
    Tools:             openAITools,         // <-- this is a tool completion
    Seed:              openai.Int(0),
    Model:             modelTools,          // <-- a model with tools support
    Temperature:       openai.Opt(0.0),
}

// Make initial completion request to detect the tools
completion, _ := dmrClient.Chat.Completions.New(ctx, params)

// Check if the completion contains any tool calls
detectedToolCalls := completion.Choices[0].Message.ToolCalls

for _, toolCall := range detectedToolCalls {
    // toolCall.Function.Arguments is a JSON String
    // Convert the JSON string to a (map[string]any)
    var args map[string]any
    err = json.Unmarshal([]byte(toolCall.Function.Arguments), &args)
    if err != nil {
        log.Println("😡 Failed to unmarshal arguments:", err)
    }

    // Call the tool with the arguments
    toolResponse, err := CallTool(ctx, toolCall.Function.Name, args)
    if err != nil {
        log.Println("😡 Failed to call tool:", err)
    }
    if toolResponse != nil && len(toolResponse.Content) > 0 && toolResponse.Content[0].TextContent != nil {
        // add the result as a tool message to the list of messages
        messages = append(
            messages,
            openai.ToolMessage(
                toolResponse.Content[0].TextContent.Text,
                toolCall.ID,
            ),
        )

    }
}
fmt.Println("🎉 tools execution completed.")
```

**Then, second completion**: a "classical" chat completion

```golang	
// New request param
params = openai.ChatCompletionNewParams{
    Messages:    messages,          // <-- the messages now contain the tool results    
    Model:       modelChat,         // <-- you can use another model
    Temperature: openai.Opt(0.9),   // <-- you can make the model more creative
}

stream := dmrClient.Chat.Completions.NewStreaming(ctx, params)

for stream.Next() {
    chunk := stream.Current()
    // Stream each chunk as it arrives
    if len(chunk.Choices) > 0 && chunk.Choices[0].Delta.Content != "" {
        fmt.Print(chunk.Choices[0].Delta.Content)
    }
}
```
