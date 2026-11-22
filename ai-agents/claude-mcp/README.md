# Connect Claude, Cursor or Codex to Hello (MCP)

Hello runs a [Model Context Protocol](https://modelcontextprotocol.io) server at
`https://helloapi.lakshyapocketlab.com/mcp`. Connect an AI assistant to it and ask in plain language:
list your numbers, send a template, check delivery, create templates. Every tool call is the
matching `/v1` request made with your API key, so plan limits and billing apply.

Full reference: https://lakshyapocketlab.com/docs/ai/integrate-with-ai/

## Prerequisites

- A Hello API key (`hello_live_...`) from **Integrations** in the dashboard. Make one key per assistant.
- Keep the key out of prompts and repositories. Messages sent this way are real and billed.

## Option 1 (recommended): connect to the hosted server

### Claude Code

```sh
claude mcp add --transport http hello https://helloapi.lakshyapocketlab.com/mcp \
  --header "Authorization: Bearer $HELLO_API_KEY"
claude mcp list
```

### Cursor

`.cursor/mcp.json` (Cursor reads `HELLO_API_KEY` from your environment):

```json
{
  "mcpServers": {
    "hello": {
      "url": "https://helloapi.lakshyapocketlab.com/mcp",
      "headers": { "Authorization": "Bearer ${env:HELLO_API_KEY}" }
    }
  }
}
```

### Codex

`~/.codex/config.toml`:

```toml
[mcp_servers.hello]
url = "https://helloapi.lakshyapocketlab.com/mcp"
bearer_token_env_var = "HELLO_API_KEY"
```

## Option 2: stdio, with hello-mcp

For clients that only speak stdio, such as Claude Desktop, use
[`@lakshyapocketlab/hello-mcp`](https://github.com/lakshyapocketlab/hello-mcp). It proxies to the same hosted server.

`claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "hello": {
      "command": "npx",
      "args": ["-y", "@lakshyapocketlab/hello-mcp"],
      "env": { "HELLO_API_KEY": "hello_live_..." }
    }
  }
}
```

Claude Desktop's custom connectors can also use the remote URL with the key as an
`?Authentication=` query parameter, but a key in a URL leaks more easily. Prefer the stdio config above.

## Try it

Restart the assistant, then ask:

- "Ping Hello." (checks the key)
- "List my WhatsApp numbers and the approved templates on each."
- "Send the `order_shipped` template in `en` to +91 99999 99999 with the name Asha, then tell me if it was delivered."
- "Write a Node.js function that sends a WhatsApp text through Hello and test it on my number."

A good assistant asks before it sends, deletes a template or changes a webhook. If yours does not, ask it to confirm first.
