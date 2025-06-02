# Tests on MBA M4
> `Docker Model Runner version v0.1.23`

## Models 
The test runs 2 completions:
- a tool completion with `ai/qwen2.5:1.5B-F16`
- a classic chat completion with `ai/llama3.2:latest`

**models.json**:
```json

```

## Run the test

**First** update the `.env` file:
```bash
# From a container
#MODEL_RUNNER_BASE_URL=http://model-runner.docker.internal/engines/llama.cpp/v1/
# Locally
MODEL_RUNNER_BASE_URL=http://localhost:12434/engines/llama.cpp/v1/
MODEL_RUNNER_TOOL_LLM=ai/qwen2.5:1.5B-F16
MODEL_RUNNER_CHAT_LLM=ai/llama3.2:latest

```

**Then**, start the test:
```bash
cd 05-complex-tools-dmr
go run main.go
```

### Tool completion result

```raw

```

