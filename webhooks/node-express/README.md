# Receive WhatsApp webhooks with Express

A small Express server that answers Meta's verification handshake and receives WhatsApp events (incoming messages and delivery statuses) for a number that uses Hello's **direct webhook**.

## How authentication works here

Read this before copying the code.

- **Handshake.** Meta calls `GET <url>?hub.mode=subscribe&hub.verify_token=...&hub.challenge=...`. The server checks the verify token and replies `200` with the challenge as plain text.
- **Secret path token.** With Hello's direct webhook, Meta signs deliveries with *Hello's* app secret, which is not shared with you, so you cannot verify `X-Hub-Signature-256`. Hello's docs say to authenticate deliveries with a long random token in the URL path instead. This example serves the webhook only at `/hooks/whatsapp/<WEBHOOK_PATH_SECRET>` and returns 404 elsewhere.
- **Optional signature check.** The code also implements Meta's standard check (`X-Hub-Signature-256` = `sha256=` + hex HMAC-SHA256 of the raw body, keyed with the app secret, compared in constant time). It runs only when `WHATSAPP_APP_SECRET` is set, which is useful if the webhook is signed with an app secret you own. It is off by default because it cannot work with Hello's app secret.

See [Direct webhook](https://lakshyapocketlab.com/docs/webhooks/direct-webhook/).

## Prerequisites

- Node.js 18 or newer
- A plan with direct webhook, and a public HTTPS URL that reaches this server (for local development use a tunnel such as `cloudflared tunnel --url http://localhost:3000`)
- The server must be running **before** you set the webhook, because Meta verifies the URL immediately

## Environment variables

| Variable | Required | Description |
|---|---|---|
| `WHATSAPP_VERIFY_TOKEN` | yes | 8 to 100 printable characters, no spaces. You choose it and send the same value when setting the webhook |
| `WEBHOOK_PATH_SECRET` | yes | Long random string used as the last URL path segment, e.g. `openssl rand -hex 16` |
| `WHATSAPP_APP_SECRET` | no | Enables the `X-Hub-Signature-256` check (only if you own the signing app secret) |
| `PORT` | no | Default `3000` |

## Run

```sh
npm install
export WHATSAPP_VERIFY_TOKEN=$(openssl rand -hex 16)
export WEBHOOK_PATH_SECRET=$(openssl rand -hex 16)
npm start
```

Then set the webhook for a number, with the token you exported (URL must be HTTPS, at most 200 characters):

```sh
curl -X PUT "https://api.lakshyapocketlab.com/v1/$PHONE_NUMBER_ID/webhook-override" \
  -H "Authorization: Bearer $HELLO_API_KEY" \
  -H "Content-Type: application/json" \
  -d "{\"url\": \"https://YOUR-HOST/hooks/whatsapp/$WEBHOOK_PATH_SECRET\",
       \"verify_token\": \"$WHATSAPP_VERIFY_TOKEN\",
       \"acknowledge_no_tracking\": true}"
```

While a direct webhook is on, Hello no longer receives that number's events, so its dashboard stops tracking messages and statuses for it. Remove it with `DELETE` on the same URL to go back.
