# Live smoke test

One program per SDK (Node.js, Python, Go) that runs against the real Hello API and walks
through the main surface: ping, phone numbers, templates, message sends, media upload,
message status and mark-as-read. Use it to confirm that a key, a number and the published SDK
work end to end.

> **Cost warning.** With `--send` (`-send` in Go) the program sends **real, billed WhatsApp
> messages** to `HELLO_TO`. Each send is charged by Hello and WhatsApp. Run it only against a
> number you control.
>
> **Default sends and the 24-hour window.** By default `--send` sends two messages: the OTP
> template (step 5) and a static reply-buttons message (printed as step 6). The OTP template is
> an AUTHENTICATION template, so it can start a conversation and works at any time. The
> reply-buttons message is free-form too, so it also needs an open 24-hour window: it only goes
> out if `HELLO_TO` has messaged the sending number in the last 24 hours. Outside that window it
> fails with `OUTSIDE_24H_WINDOW`. Send a message from `HELLO_TO` to the number first.
>
> **`--all`.** `--all` (`-all` in Go) also sends a text, image, document, video, audio,
> uploaded image, location, contacts, list and reaction. Those are free-form, so they need the
> same open 24-hour window.

## What it does

Each program runs these steps in order, printing `PASS`, `FAIL`, `SKIP` or `DRY` per step. A
failure does not stop the run. It ends with a summary and exits 1 if anything failed.

| Step | Read-only? | What |
|---|---|---|
| 1 | yes | `ping` |
| 2 | yes | List phone numbers, find `HELLO_PHONE_NUMBER_ID`, read its WABA id |
| 3 | yes | List the WABA's templates (name, language, status, category), find `HELLO_TEMPLATE_NAME` in `HELLO_TEMPLATE_LANGUAGE` |
| 4 | yes | Get that template by id |
| 5 | **send** | Send the OTP template. Sent by default |
| 6 | **send** | Static reply buttons. Sent by default. With `--all`, a text message comes first (synchronous, so the reply carries the wamid) |
| 7 | **send, `--all`** | Image and document by link. Video and audio only if their URL is set, otherwise `SKIP` |
| 8 | **send, `--all`** | Download the image, upload it with the SDK, send it by media id |
| 9 | **send, `--all`** | Location, contacts, interactive list, and a thumbs-up reaction to the step 6 text message |
| 10 | yes | Read the status of each sent message after 5 s |
| 11 | **write** | Mark `HELLO_INBOUND_MESSAGE_ID` as read, or `SKIP` if unset |

Without `--all`, the text (step 6), steps 7 to 8, the location, contacts, list and reaction
(step 9) print `SKIP ... needs --all`.

**Template (steps 3 and 5).** `HELLO_TEMPLATE_NAME` is required with `--send` and must name an
**APPROVED AUTHENTICATION (OTP) template** on the number's WABA. The program looks for exactly
that name in `HELLO_TEMPLATE_LANGUAGE` (default `en_US`). It does not pick a template for you.
The OTP (`HELLO_OTP`, default `123456`) is the template's only parameter: it fills body `{{1}}`
and, if the template has a BUTTONS component, the copy-code button. A template in any other
category is refused. Only the first 100 templates on the WABA are listed.

**Dry run and `--send`.** Without `--send` (`-send` in Go) steps 1 to 4 run for real, and every
step that would send or write prints `DRY ... would ...` instead. Nothing is sent. Steps 5 to 11
are gated by the flag. `--all` without `--send` only changes the `SKIP` lines to `DRY` lines.

**Idempotency.** Each send uses an idempotency key built from the run id, so a retry of the same
send within one run cannot deliver a second copy. A new run is a new set of sends.

**Secrets.** The API key is read from the environment and never printed.

## Setup

1. Copy the example env file and fill it in. `.env` is git-ignored.
   ```sh
   cp smoke-test/.env.example smoke-test/.env
   ```
   Required: `HELLO_API_KEY`, `HELLO_PHONE_NUMBER_ID`. Required with `--send`: `HELLO_TEMPLATE_NAME`
   (an approved AUTHENTICATION/OTP template) and `HELLO_TO`. Optional: `HELLO_TEMPLATE_LANGUAGE`
   (default `en_US`), `HELLO_OTP` (default `123456`), and the media URLs (see the comments in
   `.env.example`).
2. Load it in the shell you run from:
   ```sh
   set -a; source smoke-test/.env; set +a
   ```
   None of the programs read `.env` themselves. Without the variables set, each program exits
   with a message that says which one is missing.

Both `HELLO_TO` and the other phone and key values stay in your `.env`. Do not paste them into
issues, logs or commits.

## Run

Each language has a dry run (no `--send`), a live run (`--send`: OTP template and reply buttons)
and a live run with every message type (`--send --all`). Run the dry run first.

### Node.js (TypeScript, tsx)

```sh
cd smoke-test/node
npm install
npm start                    # dry run (after loading .env)
npm start -- --send          # live: OTP template and reply buttons
npm start -- --send --all    # live: every message type
```

Type-check: `npm run typecheck`. Requires Node.js 18 or newer.

### Python

```sh
cd smoke-test/python
python3 -m venv .venv && . .venv/bin/activate
pip install -r requirements.txt
python smoke_test.py                # dry run (after loading .env)
python smoke_test.py --send         # live: OTP template and reply buttons
python smoke_test.py --send --all   # live: every message type
```

Requires Python 3.9 or newer.

### Go

```sh
cd smoke-test/go
go run .              # dry run (after loading .env)
go run . -send        # live: OTP template and reply buttons
go run . -send -all   # live: every message type
```

Requires Go 1.22 or newer. `-send` and `--send` both work, and so do `-all` and `--all`.

## Expected output

A dry run with a placeholder key, which is what CI-style checks see, looks like this. The 401
comes from the real API, so the read-only steps fail and the rest report what they would do:

```text
Hello smoke test, run smoke-1795340018674: dry run, nothing is sent (add --send to send)
FAIL 1 ping: HTTP 401 UNAUTHORIZED: Invalid or disabled API key
FAIL 2 find phone number: HTTP 401 UNAUTHORIZED: Invalid or disabled API key
FAIL 3 list templates and find HELLO_TEMPLATE_NAME: no WABA id from step 2
FAIL 4 get template by id: no template from step 3
DRY  5 send OTP template: would send template <HELLO_TEMPLATE_NAME> with code 123456 to <HELLO_TO>
SKIP 6 send text: needs --all
SKIP 7 send image by link: needs --all
...
DRY  6 send interactive reply buttons: would send a static reply-buttons message to <HELLO_TO>
...
DRY  10 message status: would read the status of each sent message after 5 s
SKIP 11 mark as read: HELLO_INBOUND_MESSAGE_ID not set

Summary: 0 passed, 4 failed, 11 skipped, 3 dry-run only
Dry run: nothing was sent. Add --send to send real, billed messages.
```

A real dry run with a valid key shows `PASS` for steps 1 to 4, followed by the `DRY` lines and a
summary. A live run (`--send`) shows `PASS` or `FAIL` for the OTP template and the reply
buttons, then the status of each sent message in step 10. Example:

```text
PASS 1 ping: status ok
PASS 2 find phone number: waba_id 1234567890, can_send true
PASS 3 list templates and find HELLO_TEMPLATE_NAME: using my_otp_template (en_US, AUTHENTICATION, APPROVED)
PASS 4 get template by id: my_otp_template en_US APPROVED
PASS 5 send OTP template: hello_id 0a1b2c3d-...
PASS 6 send interactive reply buttons: hello_id 4e5f6a7b-...
SKIP 6 send text: needs --all
...
PASS 10 status template: status sent
PASS 10 status buttons: status sent
...
Summary: 8 passed, 0 failed, 11 skipped, 0 dry-run only
```

If no key is set, nothing runs:

```text
HELLO_API_KEY is not set. Copy smoke-test/.env.example to .env, fill it in and load it. Nothing was run.
```

## Troubleshooting

- `401 UNAUTHORIZED`: the key is wrong, disabled, or a test key. Create a live key in the
  dashboard under Integrations.
- `ACCOUNT_NOT_FOUND` or no matching number in step 2: `HELLO_PHONE_NUMBER_ID` does not belong to
  this key's organization.
- `HELLO_TEMPLATE_NAME is not set` (before any send): set it to an APPROVED AUTHENTICATION (OTP)
  template on the WABA.
- `no template <name> in <language>` in step 3: the name or `HELLO_TEMPLATE_LANGUAGE` does not
  match a template on the WABA. Check the list printed above it.
- `... must be an AUTHENTICATION (OTP) template` in step 5: `HELLO_TEMPLATE_NAME` names a
  marketing or utility template. Use an OTP template instead.
- `OUTSIDE_24H_WINDOW` on step 6 (reply buttons), or on steps 6 to 9 with `--all`: see the
  24-hour note above.
- `step 6 returned no wamid` (reaction skipped, `--all` only): the text was queued rather than
  accepted within 25 seconds, or it failed. Check step 10.

## Files

- `node/smoke-test.ts`: Node.js program (`@lakshyapocketlab/hello@^1.0.0`)
- `python/smoke_test.py`: Python program (`lakshyapocketlab-hello>=1.0.0,<2`)
- `go/main.go`: Go program (`github.com/lakshyapocketlab/hello-go v1.0.0`)
- `.env.example`: every variable the programs read

These programs are checked in CI by type-checking or compiling only. CI never calls the Hello
API.

## Links

- Docs: https://lakshyapocketlab.com/docs/
- SDKs: [hello-node](https://github.com/lakshyapocketlab/hello-node), [hello-python](https://github.com/lakshyapocketlab/hello-python), [hello-go](https://github.com/lakshyapocketlab/hello-go)
