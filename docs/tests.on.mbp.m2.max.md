# Tests on MBP M2 Max
> `Docker Model Runner version v0.1.23`

## Models 
The test runs 2 completions:
- a tool completion with `ai/qwen2.5:1.5B-F16`
- a classic chat completion with `ai/llama3.2:latest`

**models.json**:
```json
{
  "models": [
    {
      "id": "sha256:436bb282b41968a83638482999980267ca8d7e8b5574604460efa9efff11cf59",
      "tags": [
        "ai/llama3.2:latest"
      ],
      "files": [
        "sha256:91651317fc958f8e6b4f1414cd71e2529ad335b4a6af9c3add2f5f09c822fba0",
        "sha256:0b4284c1f87029e67654c7953afa16279961632cf73dcfe33374c4c2f298fa35",
        "sha256:40e2777d7faa6beaf98400654170f414d8ab29b921b5163ad4ea0a1d39894201",
        "sha256:3c0bb09bb8fdf51bccc24e9959ab02c0fdb32c632ccb7e7c301c75db8ec499fc"
      ]
    },
    {
      "id": "sha256:b83c287163f67ba50ebebd583ae0f02fa3f9a8ebe5d4596c6d367460a75d88e6",
      "tags": [
        "ai/qwen2.5:1.5B-F16"
      ],
      "files": [
        "sha256:b6eaec3509f1d0373d1f4802654c4a7bfcde0768645d2d45e8885f6922b428ee",
        "sha256:609e2cb599f84aaa41d8ef29d8fdb04d164fab22e8d9292ca34a599d0f56a338",
        "sha256:5f24375ec7f1d1e316ac7789e5c694234810d1a9332a09f112b2095e97877d64"
      ]
    },
    {
      "id": "sha256:d23f1d398f07505076f12ed79161f3b70ff07f020a93513e39cf1ad8ecee7113",
      "tags": [
        "ai/qwen2.5:latest"
      ],
      "files": [
        "sha256:7848e617403f9c800ec80c974126803ec949388869dcd3d6ff6e886de0576b9b",
        "sha256:832dd9e00a68dd83b3c3fb9f5588dad7dcf337a0db50f7d9483f310cd292e92e",
        "sha256:822fb139bd6c6d7ef88ffb970f2b24ef87b028160f96da56c5ef70bb6903a4b9"
      ]
    }
  ]
}
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
> or `docker compose up`

### Tool completion result

```raw
==================================================
🛠️  Tools completion...
==================================================
0 . 🐳 search_products {"category":"books","query":"Dune"}
✅ Found 1 products:
  - Dune (books): $14.99
```

