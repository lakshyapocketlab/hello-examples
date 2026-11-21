# Send a template with Go

Sends an approved WhatsApp template through the [Hello Go SDK](https://github.com/lakshyapocketlab/hello-go).

## Prerequisites

- Go 1.22 or newer
- A Hello API key (`hello_live_...`), the `phone_number_id` of a connected number, and an approved template

## Environment variables

| Variable | Required | Description |
|---|---|---|
| `HELLO_API_KEY` | yes | Your Hello API key |

## Run

```sh
HELLO_API_KEY=hello_live_xxx go run . <phone_number_id> +919999999999 order_shipped en Asha
```

Arguments: `phone_number_id`, recipient, template name, language, then the body parameters in order.

The SDK has no release tag yet. Until it does, point the module at a local checkout, see the [root README](../../README.md#installing-the-sdks-before-they-are-published):

```sh
git clone https://github.com/lakshyapocketlab/hello-go ../../../hello-go
go mod edit -replace github.com/lakshyapocketlab/hello-go=../../../hello-go
```
