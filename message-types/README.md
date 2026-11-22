# Message types

One small, copyable function per WhatsApp message type, in Node.js (TypeScript), Python and Go. Each program sends a single message chosen by a command-line argument, prints the Hello ID, then polls the message status for up to 30 seconds and prints every change (`queued`, `sent`, `delivered`, `read` or `failed`). Request bodies follow the [Hello message docs](https://lakshyapocketlab.com/docs/messages/) exactly, which mirror the WhatsApp Cloud API.

Messages are real and billed by WhatsApp. Run these against your own number.

## Run

Set the variables from [`.env.example`](.env.example) (the programs read the process environment and do not load `.env` files), then:

```sh
# Node.js (Node 20+)
cd node && npm install && npm start -- text

# Python
cd python && pip install -r requirements.txt && python send.py text

# Go
cd go && go run . text
```

`all` lists the types instead of sending: `npm start -- all`, `python send.py all`, `go run . all`. With `HELLO_API_KEY` unset, every program exits with status 1 and `HELLO_API_KEY is not set`.

## Types

Free-form types need an open [24-hour window](https://lakshyapocketlab.com/docs/getting-started/how-sending-works/#the-24-hour-customer-service-window): the recipient must have messaged you in the last 24 hours. Templates are the only way to start a conversation.

| Type (argument) | What it is for | 24h window? | Docs | Node | Python | Go |
|---|---|---|---|---|---|---|
| `text` | Plain text up to 4,096 characters, with an optional link preview | Yes | [Text](https://lakshyapocketlab.com/docs/messages/text/) | `sendText` | `send_text` | `sendText` |
| `image` | JPEG or PNG up to 5 MB with a caption | Yes | [Image](https://lakshyapocketlab.com/docs/messages/image/) | `sendImage` | `send_image` | `sendImage` |
| `video` | MP4 or 3GPP up to 16 MB with a caption | Yes | [Video](https://lakshyapocketlab.com/docs/messages/video/) | `sendVideo` | `send_video` | `sendVideo` |
| `audio` | Audio file or voice note up to 16 MB, no caption | Yes | [Audio](https://lakshyapocketlab.com/docs/messages/audio/) | `sendAudio` | `send_audio` | `sendAudio` |
| `document` | PDF, Office or text file up to 100 MB with a file name and caption | Yes | [Document](https://lakshyapocketlab.com/docs/messages/document/) | `sendDocument` | `send_document` | `sendDocument` |
| `sticker` | Static WebP up to 100 KB, by uploaded media ID only | Yes | [Sticker](https://lakshyapocketlab.com/docs/messages/sticker/) | `sendSticker` | `send_sticker` | `sendSticker` |
| `location` | A map pin with name and address | Yes | [Location](https://lakshyapocketlab.com/docs/messages/location/) | `sendLocation` | `send_location` | `sendLocation` |
| `contacts` | One or more contact cards | Yes | [Contacts](https://lakshyapocketlab.com/docs/messages/contacts/) | `sendContacts` | `send_contacts` | `sendContacts` |
| `buttons` | Interactive reply buttons (up to 3) | Yes | [Interactive: reply buttons](https://lakshyapocketlab.com/docs/messages/interactive/#reply-buttons) | `sendButtons` | `send_buttons` | `sendButtons` |
| `list` | Interactive list: one button opening up to 10 rows | Yes | [Interactive: list](https://lakshyapocketlab.com/docs/messages/interactive/#list) | `sendList` | `send_list` | `sendList` |
| `cta-url` | Interactive button that opens a URL, no pre-approval | Yes | [Interactive: call-to-action URL](https://lakshyapocketlab.com/docs/messages/interactive/#call-to-action-url) | `sendCtaUrl` | `send_cta_url` | `sendCtaURL` |
| `reaction` | One emoji on a message the customer sent | Yes | [Reaction](https://lakshyapocketlab.com/docs/messages/reaction/) | `sendReaction` | `send_reaction` | `sendReaction` |
| `reply` | Quote an earlier message with `context` (works with every type) | Yes | [Reply to a message](https://lakshyapocketlab.com/docs/messages/reply-context/) | `sendReply` | `send_reply` | `sendReply` |
| `template` | Approved template, positional or named body variables | No | [Template](https://lakshyapocketlab.com/docs/messages/template/) | `sendTemplate` | `send_template` | `sendTemplate` |
| `template-header-footer` | Template with a text header variable and a footer | No | [Header & footer](https://lakshyapocketlab.com/docs/messages/template-header-footer/) | `sendTemplateHeaderFooter` | `send_template_header_footer` | `sendTemplateHeaderFooter` |
| `template-buttons` | Template with a dynamic URL button | No | [Buttons](https://lakshyapocketlab.com/docs/messages/template-buttons/) | `sendTemplateButtons` | `send_template_buttons` | `sendTemplateButtons` |
| `template-media` | Template with an image header, body variables, footer and buttons | No | [Media with buttons](https://lakshyapocketlab.com/docs/messages/template-media-buttons/) | `sendTemplateMedia` | `send_template_media` | `sendTemplateMedia` |

The SDKs have helpers for text, templates and media (`sendText`, `sendTemplate`, `sendMedia` in Node; `send_text`, `send_template`, `send_media` in Python; `SendText`, `SendTemplate`, `SendMedia` in Go). Everything else (location, contacts, interactive, reaction) uses the raw send with the Cloud API body, as the docs show. The Go text helper has no `preview_url` option, and no Go helper takes a reply context, so those two examples also use the raw send.

## Notes per type

- **Media by link.** `image`, `video`, `audio` and `document` take a public HTTPS URL (`HELLO_IMAGE_URL`, `HELLO_VIDEO_URL`, `HELLO_AUDIO_URL`, `HELLO_DOCUMENT_URL`). Pass `id` instead of `link` for a file you uploaded with the media endpoint; `id` wins if both are set. See [supported media](https://lakshyapocketlab.com/docs/reference/supported-media/).
- **Sticker.** Stickers have no `link` option. Upload a static WebP first (`hello.media.upload`, `client.media.upload`, `Media.Upload`) and put the returned ID in `HELLO_STICKER_MEDIA_ID`.
- **Location.** `latitude` and `longitude` are numbers in a location message but strings in a template location header.
- **Contacts.** `contacts` is an array at the top level of the request, not an object.
- **Reaction and reply.** Both need the `wamid` of a message the customer sent you (`HELLO_REPLY_TO_MESSAGE_ID`), which you get from an inbound webhook. An empty emoji removes a reaction. An unknown `message_id` on a reply is sent without the quote instead of failing.
- **Interactive.** Reply button titles and list button labels are limited to 20 characters; lists hold up to 10 rows in total. Taps come back as inbound messages carrying the button or row `id`. Location request, address request and flow kinds are documented on the [interactive page](https://lakshyapocketlab.com/docs/messages/interactive/) and use the same raw send.

## Templates

Templates take their name and language from the environment (`HELLO_TEMPLATE_LANGUAGE`, default `en_US`, applies to all four). The header type, footer text and buttons are fixed when the template is created and approved; a send only supplies the variable parts. Each example sends the sample values below, so create the template to match (or change the values in the program).

| Argument | Env var | Template shape it sends to |
|---|---|---|
| `template` | `HELLO_TEMPLATE_NAME` (default `hello_world`), `HELLO_TEMPLATE_PARAMS` | Any body. With no params it sends no components. Positional: `Asha,#10432` for `Hi {{1}}, your order {{2}} has shipped.` Named: `first_name=Asha,otp=482913` for `Hi {{first_name}}, your OTP is {{otp}}.` (named parameters add `parameter_name`; order does not matter) |
| `template-header-footer` | `HELLO_TEMPLATE_HEADER_FOOTER_NAME` | Header `TEXT` `Reminder for {{1}}`, body `Your appointment with {{1}} is on {{2}} at {{3}}.`, footer `Reply STOP to unsubscribe` (sends `Asha`; `Dr. Rao`, `3 October`, `10:30 am`) |
| `template-buttons` | `HELLO_TEMPLATE_BUTTONS_NAME` | Body with one variable plus a `URL` button whose URL ends in `{{1}}` (sends `Asha` and the suffix `ORD-10432` as button `index` `"0"`). Quick-reply, fixed-URL and call buttons need no `button` component: send only the body |
| `template-media` | `HELLO_TEMPLATE_MEDIA_NAME`, `HELLO_IMAGE_URL` | Header `IMAGE`, body `Hi {{1}}, the {{2}} just landed. Get 15% off this week only.`, footer and buttons as approved (sends the image link, `Asha`, `Runner 2`) |

Create templates with [`POST /v1/{waba_id}/message_templates`](https://lakshyapocketlab.com/docs/api/create-template/) and send only once the status is `APPROVED`. A send that does not match the approved template (wrong parameter count, missing header media) is rejected by WhatsApp, and the message then shows `failed` with an error message in the tracking output. For a video or document header, swap the header parameter as the [header & footer page](https://lakshyapocketlab.com/docs/messages/template-header-footer/#send) shows.

## Delivery tracking

A send returns at once with `messages[0].hello_id` (the key's default send mode is asynchronous). The programs then call `messages.get` with that ID every 2 seconds for up to 30 seconds, printing each status change and stopping at `delivered`, `read` or `failed`. Statuses only move forward; `delivered` can be final for recipients with read receipts off. For production, receive the same updates by [webhook](../webhooks/node-express) instead of polling.

## Why not one file per type?

Each type is one self-contained function in a single program per language, so you can copy a function without the rest. The shared pieces are only `need` (read a required env var), the dispatch table and the status poller.
