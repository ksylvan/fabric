# Fabric REST API

Fabric's REST API provides HTTP access to all core functionality: chat completions, pattern management, contexts, sessions, and more.

## Quick Start

Start the server:

```bash
fabric --serve
```

The server listens on `http://127.0.0.1:8080` by default.

Test it:

```bash
curl http://localhost:8080/patterns/names
```

## Interactive API Documentation

Fabric includes Swagger/OpenAPI documentation with an interactive UI:

- **Swagger UI**: [http://localhost:8080/swagger/index.html](http://localhost:8080/swagger/index.html)
- **OpenAPI JSON**: [http://localhost:8080/swagger/doc.json](http://localhost:8080/swagger/doc.json)
- **OpenAPI YAML**: [http://localhost:8080/swagger/swagger.yaml](http://localhost:8080/swagger/swagger.yaml)

The Swagger UI lets you:

- Browse all available endpoints
- View request/response schemas
- Test API calls directly in your browser
- See authentication requirements

**Note:** Swagger documentation endpoints are publicly accessible even when API key authentication is enabled. Only the actual API endpoints require authentication

## Server Options

| Flag | Description | Default |
| ------ | ------------- | --------- |
| `--serve` | Start the REST API server | - |
| `--address` | Server address and port | `127.0.0.1:8080` |
| `--api-key` | Enable API key authentication | (none) |

Example with custom configuration:

```bash
fabric --serve --address 0.0.0.0:9090 --api-key my_secret_key
```

## Authentication

When you set an API key with `--api-key`, all requests must include:

```http
X-API-Key: your-api-key-here
```

Example:

```bash
curl -H "X-API-Key: my_secret_key" http://localhost:8080/patterns/names
```

Without an API key, Fabric only allows loopback bind addresses such as `127.0.0.1:8080` or `localhost:8080`. To expose the REST API on `0.0.0.0`, `:8080`, or another non-loopback interface, set `--api-key`.

## Endpoints

### Chat Completions

Stream AI responses using Server-Sent Events (SSE).

**Endpoint:** `POST /chat`

**Request:**

```json
{
  "prompts": [
    {
      "userInput": "Explain quantum computing",
      "vendor": "openai",
      "model": "gpt-5.2",
      "patternName": "explain",
      "contextName": "",
      "strategyName": "",
      "variables": {}
    }
  ],
  "language": "en",
  "temperature": 0.7,
  "topP": 0.9,
  "frequencyPenalty": 0,
  "presencePenalty": 0,
  "thinking": 0
}
```

**Prompt Fields:**

| Field | Required | Default | Description |
| ------- | ---------- | --------- | ------------- |
| `userInput` | **Yes** | - | Your message or question |
| `vendor` | **Yes** | - | AI provider: `openai`, `anthropic`, `gemini`, `ollama`, etc. |
| `model` | **Yes** | - | Model name: `gpt-5.2`, `claude-sonnet-4.5`, `gemini-2.0-flash-exp`, etc. |
| `patternName` | No | `""` | Pattern to apply (from `~/.config/fabric/patterns/`) |
| `contextName` | No | `""` | Context to prepend (from `~/.config/fabric/contexts/`) |
| `strategyName` | No | `""` | Strategy to use (from `~/.config/fabric/strategies/`) |
| `variables` | No | `{}` | Variable substitutions for patterns (e.g., `{"role": "expert"}`) |

**Chat Options:**

| Field | Required | Default | Description |
| ------- | ---------- | --------- | ------------- |
| `language` | No | `"en"` | Language code for responses |
| `temperature` | No | `0.7` | Randomness (0.0-1.0) |
| `topP` | No | `0.9` | Nucleus sampling (0.0-1.0) |
| `frequencyPenalty` | No | `0.0` | Reduce repetition (-2.0 to 2.0) |
| `presencePenalty` | No | `0.0` | Encourage new topics (-2.0 to 2.0) |
| `thinking` | No | `0` | Reasoning level (0=off, or numeric for tokens) |

**Response:**

Server-Sent Events stream with `Content-Type: text/readystream`. Each line contains JSON:

```json
{"type": "content", "format": "markdown", "content": "Quantum computing uses..."}
{"type": "content", "format": "markdown", "content": " quantum mechanics..."}
{"type": "complete", "format": "markdown", "content": ""}
```

**Types:**

- `content` - Response chunk
- `error` - Error message
- `complete` - Stream finished

**Formats:**

- `markdown` - Standard text
- `mermaid` - Mermaid diagram
- `plain` - Plain text

**Example:**

```bash
curl -X POST http://localhost:8080/chat \
  -H "Content-Type: application/json" \
  -d '{
    "prompts": [{
      "userInput": "What is Fabric?",
      "vendor": "openai",
      "model": "gpt-5.2",
      "patternName": "explain"
    }]
  }'
```

### Patterns

Manage reusable AI prompts.

| Method | Endpoint | Description |
| -------- | ---------- | ------------- |
| `GET` | `/patterns/names` | List all pattern names |
| `GET` | `/patterns/:name` | Get pattern content |
| `GET` | `/patterns/exists/:name` | Check if pattern exists |
| `POST` | `/patterns/:name` | Create or update pattern |
| `DELETE` | `/patterns/:name` | Delete pattern |
| `PUT` | `/patterns/rename/:oldName/:newName` | Rename pattern |
| `POST` | `/patterns/:name/apply` | Apply pattern with variables |

Pattern endpoints resolve configured pattern names only. Filesystem pattern paths remain a CLI-only feature and are not accepted over HTTP.

**Example - Get pattern:**

```bash
curl http://localhost:8080/patterns/summarize
```

**Example - Apply pattern with variables:**

```bash
curl -X POST http://localhost:8080/patterns/translate/apply \
  -H "Content-Type: application/json" \
  -d '{
    "input": "Hello world",
    "variables": {"lang_code": "es"}
  }'
```

**Example - Create pattern:**

```bash
curl -X POST http://localhost:8080/patterns/my_custom_pattern \
  -H "Content-Type: text/plain" \
  -d "You are an expert in explaining complex topics simply..."
```

### Contexts

Manage context snippets that prepend to prompts.

| Method | Endpoint | Description |
| -------- | ---------- | ------------- |
| `GET` | `/contexts/names` | List all context names |
| `GET` | `/contexts/:name` | Get context content |
| `GET` | `/contexts/exists/:name` | Check if context exists |
| `POST` | `/contexts/:name` | Create or update context |
| `DELETE` | `/contexts/:name` | Delete context |
| `PUT` | `/contexts/rename/:oldName/:newName` | Rename context |

Context and session resource names are treated as single identifiers and cannot include path separators.

### Sessions

Manage chat conversation history.

| Method | Endpoint | Description |
| -------- | ---------- | ------------- |
| `GET` | `/sessions/names` | List all session names |
| `GET` | `/sessions/:name` | Get session messages (JSON array) |
| `GET` | `/sessions/exists/:name` | Check if session exists |
| `POST` | `/sessions/:name` | Save session messages |
| `DELETE` | `/sessions/:name` | Delete session |
| `PUT` | `/sessions/rename/:oldName/:newName` | Rename session |

### Models

List available AI models.

**Endpoint:** `GET /models/names`

**Response:**

```json
{
  "models": ["gpt-5.2", "gpt-5-mini", "claude-sonnet-4.5", "gemini-2.0-flash-exp"],
  "vendors": {
    "openai": ["gpt-5.2", "gpt-5-mini"],
    "anthropic": ["claude-sonnet-4.5", "claude-opus-4.5"],
    "gemini": ["gemini-2.0-flash-exp", "gemini-2.0-flash-thinking-exp"]
  }
}
```

### Strategies

List available prompt strategies (Chain of Thought, etc.).

**Endpoint:** `GET /strategies`

**Response:**

```json
[
  {
    "name": "chain_of_thought",
    "description": "Think step by step",
    "prompt": "Let's think through this step by step..."
  }
]
```

### YouTube Transcripts

Extract transcripts from YouTube videos.

**Endpoint:** `POST /youtube/transcript`

**Request:**

```json
{
  "url": "https://youtube.com/watch?v=dQw4w9WgXcQ",
  "timestamps": false
}
```

**Response:**

```json
{
  "videoId": "Video ID",
  "title": "Video Title",
  "description" : "Video description...",
  "transcript": "Full transcript text..."
}
```

**Example:**

```bash
curl -X POST http://localhost:8080/youtube/transcript \
  -H "Content-Type: application/json" \
  -d '{"url": "https://youtube.com/watch?v=dQw4w9WgXcQ", "timestamps": true}'
```

### Configuration

Manage API keys and environment settings.

**Get configuration:**

`GET /config`

Returns API keys and URLs for supported vendors. Secret values are masked to their last 4 characters.

**Update configuration:**

`POST /config/update`

```json
{
  "openai_api_key": "sk-...",
  "anthropic_api_key": "sk-ant-...",
  "ollama_url": "http://localhost:11434"
}
```

Updates `~/.config/fabric/.env` with merged values.

- Omitted fields are preserved.
- Masked values returned by `GET /config` are treated as unchanged and preserved.
- Sending an empty string clears that specific setting.

## Complete Workflow Examples

### Example: Summarize a YouTube Video

This example shows how to extract a YouTube transcript and summarize it using the `youtube_summary` pattern. This requires two API calls:

#### Step 1: Extract the transcript

```bash
curl -X POST http://localhost:8080/youtube/transcript \
  -H "Content-Type: application/json" \
  -d '{
    "url": "https://youtube.com/watch?v=dQw4w9WgXcQ",
    "timestamps": false
  }' > transcript.json
```

Response:

```json
{
  "videoId": "dQw4w9WgXcQ",
  "title": "Rick Astley - Never Gonna Give You Up (Official Video)",
  "description": "The official video for “Never Gonna Give You Up” by Rick Astley...",
  "transcript": "We're no strangers to love. You know the rules and so do I..."
}
```

#### Step 2: Summarize the transcript

Extract the transcript text and send it to the chat endpoint with the `youtube_summary` pattern:

```bash
# Extract transcript text from JSON
TRANSCRIPT=$(cat transcript.json | jq -r '.transcript')

# Send to chat endpoint with pattern
curl -X POST http://localhost:8080/chat \
  -H "Content-Type: application/json" \
  -d "{
    \"prompts\": [{
      \"userInput\": \"$TRANSCRIPT\",
      \"vendor\": \"openai\",
      \"model\": \"gpt-5.2\",
      \"patternName\": \"youtube_summary\"
    }]
  }"
```

#### Combined one-liner (using jq)

```bash
curl -s -X POST http://localhost:8080/youtube/transcript \
  -H "Content-Type: application/json" \
  -d '{"url": "https://youtube.com/watch?v=dQw4w9WgXcQ", "timestamps": false}' | \
jq -r '.transcript' | \
xargs -I {} curl -X POST http://localhost:8080/chat \
  -H "Content-Type: application/json" \
  -d "{\"prompts\":[{\"userInput\":\"{}\",\"vendor\":\"openai\",\"model\":\"gpt-5.2\",\"patternName\":\"youtube_summary\"}]}"
```

#### Alternative: Using a script

```bash
#!/bin/bash
YOUTUBE_URL="https://youtube.com/watch?v=dQw4w9WgXcQ"
API_BASE="http://localhost:8080"

# Step 1: Get transcript
echo "Extracting transcript..."
TRANSCRIPT=$(curl -s -X POST "$API_BASE/youtube/transcript" \
  -H "Content-Type: application/json" \
  -d "{\"url\":\"$YOUTUBE_URL\",\"timestamps\":false}" | jq -r '.transcript')

# Step 2: Summarize with pattern
echo "Generating summary..."
curl -X POST "$API_BASE/chat" \
  -H "Content-Type: application/json" \
  -d "{
    \"prompts\": [{
      \"userInput\": $(echo "$TRANSCRIPT" | jq -Rs .),
      \"vendor\": \"openai\",
      \"model\": \"gpt-5.2\",
      \"patternName\": \"youtube_summary\"
    }]
  }"
```

#### Comparison with CLI

The CLI combines these steps automatically:

```bash
# CLI version (single command)
fabric -y "https://youtube.com/watch?v=dQw4w9WgXcQ" --pattern youtube_summary
```

The API provides more flexibility by separating transcript extraction and summarization, allowing you to:

- Extract the transcript once and process it multiple ways
- Apply different patterns to the same transcript
- Store the transcript for later use
- Use different models or vendors for summarization

## Docker Usage

Run the server in Docker:

```bash
# Setup (first time)
mkdir -p $HOME/.fabric-config
docker run --rm -it \
  -v $HOME/.fabric-config:/root/.config/fabric \
  kayvan/fabric:latest --setup

# Start server
docker run --rm -it \
  -p 8080:8080 \
  -v $HOME/.fabric-config:/root/.config/fabric \
  kayvan/fabric:latest --serve --address 0.0.0.0:8080 --api-key my_secret_key

# With authentication
docker run --rm -it \
  -p 8080:8080 \
  -v $HOME/.fabric-config:/root/.config/fabric \
  kayvan/fabric:latest --serve --address 0.0.0.0:8080 --api-key my_secret_key
```

## Ollama Compatibility Mode

Fabric can emulate Ollama's API endpoints:

```bash
fabric --serveOllama --address 127.0.0.1:11434
```

Ollama compatibility mode is loopback-only because it does not expose an authentication layer.

This mode provides:

- `GET /api/tags` - Lists patterns as models
- `GET /api/version` - Server version
- `POST /api/chat` - Ollama-compatible chat endpoint

## Error Handling

All endpoints return standard HTTP status codes:

- `200 OK` - Success
- `400 Bad Request` - Invalid input
- `401 Unauthorized` - Missing or invalid API key
- `404 Not Found` - Resource not found
- `500 Internal Server Error` - Server error

Error responses include JSON with details:

```json
{
  "error": "Pattern not found: nonexistent"
}
```

## Rate Limiting

When you enable `--api-key`, Fabric applies a built-in per-client rate limit of 60 requests per minute across REST API routes. This helps absorb credential-guessing and burst abuse before requests reach the handler layer.

For internet-facing deployments, still put Fabric behind a reverse proxy (nginx, Caddy, Traefik, etc.) so you can enforce stricter quotas, IP reputation controls, and edge-layer logging.

## Request Size Limits

Fabric caps JSON and raw-body request payloads at 16 MiB. Oversized requests return `413 Request Entity Too Large`.

## CORS

Fabric does not enable permissive global CORS. The `/chat` endpoint allows the local web development origin `http://localhost:5173` and responds to preflight requests for that origin only:

```http
Access-Control-Allow-Origin: http://localhost:5173
```

For production, configure CORS through a reverse proxy.
