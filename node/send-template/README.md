# Send a template with Node.js

Sends an approved WhatsApp template through the [Hello Node.js SDK](https://github.com/lakshyapocketlab/hello-node), then reads the message status back.

## Prerequisites

- Node.js 18 or newer
- A Hello API key (`hello_live_...`) and a connected WhatsApp number
- An **approved** template (templates can start a conversation; free-form text needs an open 24-hour window)

## Environment variables

| Variable | Required | Description |
|---|---|---|
| `HELLO_API_KEY` | yes | Your Hello API key |
| `HELLO_PHONE_NUMBER_ID` | no | Number to send from. Defaults to the first connected number |

## Run

```sh
npm install
HELLO_API_KEY=hello_live_xxx npm start -- +919999999999 order_shipped en Asha
```

Arguments: recipient, template name, language, then the body parameters in order.

The SDK is not on npm yet. Until it is, install it from GitHub, see the [root README](../../README.md#installing-the-sdks-before-they-are-published).
