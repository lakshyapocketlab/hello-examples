# Send WhatsApp messages from n8n

[n8n](https://n8n.io) has no Hello node, but you do not need one: Hello's `/v1` API takes the
WhatsApp Cloud API message body, so the **HTTP Request** node sends messages directly.

## Prerequisites

- An n8n instance (cloud or self-hosted)
- A Hello API key (`hello_live_...`), the `phone_number_id` of a connected number, and an approved template (a template is the only message type that can start a conversation)

## 1. Create the credential

In n8n, **Credentials, Create credential, Bearer Auth**. Name it `Hello API key` and paste your key into **Token**. n8n stores it encrypted, so the key never appears in the workflow.

## 2. Configure the HTTP Request node

| Field | Value |
|---|---|
| Method | `POST` |
| URL | `https://helloapi.lakshyapocketlab.com/v1/{phone_number_id}/messages` |
| Authentication | Generic Credential Type, Bearer Auth, your `Hello API key` |
| Send Headers | `Idempotency-Key`: a unique value per message (optional, but it makes retries safe) |
| Send Body | JSON |

Body for a template with one body variable:

```json
{
  "messaging_product": "whatsapp",
  "to": "+919999999999",
  "type": "template",
  "template": {
    "name": "order_shipped",
    "language": { "code": "en" },
    "components": [
      { "type": "body", "parameters": [{ "type": "text", "text": "Asha" }] }
    ]
  }
}
```

A text message (needs an open 24-hour window) is `{"to": "+919999999999", "type": "text", "text": {"body": "Hello"}}`.

The response contains the message `id` and `hello_id`. Follow a message with `GET https://helloapi.lakshyapocketlab.com/v1/messages/{hello_id}` to see queued, sent, delivered, read or failed.

## Importable workflow

[`send-template-workflow.json`](send-template-workflow.json) is a ready workflow: **Manual Trigger, Settings, Send template (Hello)**.

1. In n8n, **Workflows, Import from file**, choose `send-template-workflow.json`.
2. Open **Send template (Hello)** and select your `Hello API key` credential (the file only contains a placeholder).
3. Open **Settings** and set `phone_number_id`, `to`, `template_name`, `language` and `name`.
4. **Execute workflow**.

Messages are real and billed. To send from another trigger (a webhook, a spreadsheet row, a CRM event), replace the Manual Trigger and map its fields in the **Settings** node.

Docs: [Send a message](https://lakshyapocketlab.com/docs/api/send-message/), [Authentication](https://lakshyapocketlab.com/docs/getting-started/authentication/).
