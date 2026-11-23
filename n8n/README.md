# Send WhatsApp messages from n8n

Two ways to use Hello from [n8n](https://n8n.io):

- **The Hello community node** ([`@lakshyapocketlab/n8n-nodes-hello`](https://www.npmjs.com/package/@lakshyapocketlab/n8n-nodes-hello)): send messages and templates, and start workflows when a number receives a message. See [AI assistant for a WhatsApp Business app number](#ai-assistant-for-a-whatsapp-business-app-number) below.
- **The HTTP Request node**, with no community node: Hello's `/v1` API takes the WhatsApp Cloud API message body, so the HTTP Request node sends messages directly. The rest of this page shows how.

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

## AI assistant for a WhatsApp Business app number

[`whatsapp-business-app-ai-assistant.json`](whatsapp-business-app-ai-assistant.json) uses the Hello community node. A customer writes to your number, an AI agent replies, and when someone on your team answers from the WhatsApp Business app on the phone (coexistence), the assistant stays quiet for that customer for two hours.

**Hello Trigger, Message or echo?, Assistant allowed?, Text and not paused?, AI Agent (chat model + Simple Memory), Send reply**, plus **Person replied** on the echo branch.

1. Install the node: **Settings, Community nodes, Install a community node**, enter `@lakshyapocketlab/n8n-nodes-hello`. It installs on self-hosted n8n; n8n Cloud support follows n8n's verification review.
2. **Workflows, Import from file**, choose `whatsapp-business-app-ai-assistant.json`.
3. Select your **Hello API** credential on **Hello Trigger** and **Send reply**, and your model credential on **OpenAI Chat Model** (or swap in another chat model).
4. On **Hello Trigger**, pick your phone number. Its events are **Incoming Message** and **Coexistence Echo**.
5. On **AI Agent**, replace the `REPLACE` lines in the system message with your own hours, prices and policies.
6. **Publish** the workflow and message the number from another phone.

While the workflow is published, the number's incoming messages go to n8n instead of the Hello inbox; unpublish to give them back. The trigger needs a plan with direct webhooks and a public HTTPS n8n address.

Step-by-step guide with screenshots, in 7 languages: [Connect the WhatsApp Business app to n8n with the official WhatsApp API](https://lakshyapocketlab.com/connect-whatsapp-business-app-to-n8n/).
