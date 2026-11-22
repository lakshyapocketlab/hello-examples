# Hello examples for the WhatsApp Business API

Sample apps for [Hello](https://lakshyapocketlab.com), the WhatsApp Business API. Each folder is self-contained, with its own README (what it does, prerequisites, environment variables, how to run it).

[![CI](https://github.com/lakshyapocketlab/hello-examples/actions/workflows/ci.yml/badge.svg)](https://github.com/lakshyapocketlab/hello-examples/actions/workflows/ci.yml)
[![OpenSSF Scorecard](https://api.securityscorecards.dev/projects/github.com/lakshyapocketlab/hello-examples/badge)](https://scorecard.dev/viewer/?uri=github.com/lakshyapocketlab/hello-examples)

## Examples

| Example | Language | What it shows |
|---|---|---|
| [`node/send-template`](node/send-template) | Node.js (TypeScript) | Send an approved template and read its status |
| [`python/send-text`](python/send-text) | Python | Send a text message inside the 24-hour window |
| [`go/send-template`](go/send-template) | Go | Send an approved template with parameters |
| [`webhooks/node-express`](webhooks/node-express) | Node.js (Express) | Receive webhooks: verification handshake and event handling |
| [`ai-agents/claude-mcp`](ai-agents/claude-mcp) | Guide | Connect Claude Code, Claude Desktop, Cursor or Codex to Hello over MCP |
| [`n8n`](n8n) | Guide + workflow | Send WhatsApp messages from n8n with the HTTP Request node |

## Getting started

1. Create an API key (`hello_live_...`) in the Hello dashboard, under **Integrations**, and export it:
   ```sh
   export HELLO_API_KEY=hello_live_xxxxxxxx
   ```
2. Pick an example and follow its README.

Messages are real and billed by WhatsApp. Templates can start a conversation; free-form messages need an open 24-hour window.

## Installing the SDKs

```sh
npm i @lakshyapocketlab/hello                # Node.js
pip install lakshyapocketlab-hello           # Python
go get github.com/lakshyapocketlab/hello-go  # Go
npx @lakshyapocketlab/hello-mcp              # MCP server for AI agents
```

Once published, the examples work with the normal commands (`npm i @lakshyapocketlab/hello`, `pip install lakshyapocketlab-hello`, `go get github.com/lakshyapocketlab/hello-go`).

## Links

- Docs: https://lakshyapocketlab.com/docs/
- SDKs: [hello-node](https://github.com/lakshyapocketlab/hello-node), [hello-python](https://github.com/lakshyapocketlab/hello-python), [hello-go](https://github.com/lakshyapocketlab/hello-go)
- MCP server: [hello-mcp](https://github.com/lakshyapocketlab/hello-mcp)

## Contributing

Examples should stay minimal and read their configuration from environment variables. Never commit API keys. CI type-checks or compiles every Node.js, Python and Go example, so check yours locally first (see `.github/workflows/ci.yml`).

## License

MIT, Copyright Lakshya Pocket Lab.
