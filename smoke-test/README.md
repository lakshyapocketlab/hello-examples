# Live smoke test

One program per SDK (Node.js, Python, Go) that runs against the real Hello API and walks
through the main surface: ping, phone numbers, templates, every message type, media upload,
message status and mark-as-read. Use it to confirm that a key, a number and the published SDK
work end to end.

> **Cost warning.** With `--send` (`-send` in Go) the program sends **real, billed WhatsApp
> messages** to `HELLO_TO`. Each send is charged by Hello and WhatsApp. Run it only against a
> number you control.
>
> **The 24-hour window.** A template is the only message type that can start a conversation,
> so step 5 works at any time. Every other send (text, media, location, contacts, interactive,
> reaction) is free-form and only goes out if `HELLO_TO` has messaged the sending number in the
> last 24 hours. Outside that window those steps fail with `OUTSIDE_24H_WINDOW`. Send a message
> from `HELLO_TO` to the number first, or expect steps 6 to 10 to fail.

## What it does

Each program runs these steps in order, printing `PASS`, `FAIL`, `SKIP` or `DRY` per step. A
failure does not stop the run. It ends with a summary and exits 1 if anything failed.

| Step | Read-only? | What |
|---|---|---|
| 1 | yes | `ping` |
| 2 | yes | List phone numbers, find `HELLO_PHONE_NUMBER_ID`, read its WABA id |
| 3 | yes | List the WABA's templates (name, language, status, category, variables), pick one |
| 4 | yes | Get that template by id |
| 5 | **send** | Send the template |
| 6 | **send** | Send a text (synchronous, so the reply has the wamid) |
| 7 | **send** | Image by link, document by link (with filename and caption). Video and audio only if their URL is set, otherwise `SKIP` |
| 8 | **send** | Download the image, upload it with the SDK, send it by media id |
| 9 | **send** | Location, contacts, reply buttons, interactive list, and a 👍 reaction to the step 6 message |
| 10 | yes | Read the status of each sent message after 5 s |
| 11 | **write** | Mark `HELLO_INBOUND_MESSAGE_ID` as read, or `SKIP` if unset |

**Template choice (step 3).** If `HELLO_TEMPLATE_NAME` is set, the program uses that template
(preferring `APPROVED`, and `HELLO_TEMPLATE_LANGUAGE` when set). Otherwise it uses the first
`APPROVED` template with no variables, or else the first `APPROVED` one with body variables.
Body variables `{{1}}`..`{{n}}` are filled with `Sample 1`..`Sample n`. Templates whose header or
button text has variables are not picked, because the program does not fill those in.
Only the first 100 templates on the WABA are listed.

**Dry run and `--send`.** Without `--send` (`-send` in Go) steps 1 to 4 run for real, and every
step that would send or write prints `DRY ... would ...` instead. Nothing is sent. Steps 5 to 11
are gated by the flag.

**Idempotency.** Each send uses an idempotency key built from the run id, so a retry of the same
send within one run cannot deliver a second copy. A new run is a new set of sends.

**Secrets.** The API key is read from the environment and never printed.

## Setup

1. Copy the example env file and fill it in. `.env` is git-ignored.
   ```sh
   cp smoke-test/.env.example smoke-test/.env
   ```
   Required: `HELLO_API_KEY`, `HELLO_PHONE_NUMBER_ID`. Required with `--send`: `HELLO_TO`.
   The others are optional (see the comments in `.env.example`).
2. Load it in the shell you run from:
   ```sh
   set -a; source smoke-test/.env; set +a
   ```
   None of the programs read `.env` themselves. Without the variables set, each program exits
   with a message that says which one is missing.

Both `HELLO_TO` and the other phone and key values stay in your `.env`. Do not paste them into
issues, logs or commits.

## Run

Each language has a dry run (no `--send`) and a live run (with `--send`). Run the dry run first.

### Node.js (TypeScript, tsx)

```sh
cd smoke-test/node
npm install
npm start              # dry run (after loading .env)
npm start -- --send    # live: sends real, billed messages
```

Type-check: `npm run typecheck`. Requires Node.js 18 or newer.

### Python

```sh
cd smoke-test/python
python3 -m venv .venv && . .venv/bin/activate
pip install -r requirements.txt
python smoke_test.py          # dry run (after loading .env)
python smoke_test.py --send   # live: sends real, billed messages
```

Requires Python 3.9 or newer.

### Go

```sh
cd smoke-test/go
go run .         # dry run (after loading .env)
go run . -send  # live: sends real, billed messages
```

Requires Go 1.22 or newer. `-send` and `--send` both work.

## Expected output

A dry run with a placeholder key, which is what CI-style checks see, looks like this. The 401
comes from the real API, so the read-only steps fail and the rest report what they would do:

```text
Hello smoke test, run smoke-1795340018674: dry run, nothing is sent (add --send to send)
FAIL 1 ping: HTTP 401 UNAUTHORIZED: Invalid or disabled API key
FAIL 2 find phone number: HTTP 401 UNAUTHORIZED: Invalid or disabled API key
FAIL 3 list templates and pick one: no WABA id from step 2
FAIL 4 get template by id: no template from step 3
DRY  5 send template: would send template ? to <HELLO_TO>
DRY  6 send text: would send a text message to <HELLO_TO>
...
SKIP 7 send video by link: HELLO_MEDIA_VIDEO_URL not set
SKIP 7 send audio by link: HELLO_MEDIA_AUDIO_URL not set
...
SKIP 11 mark as read: HELLO_INBOUND_MESSAGE_ID not set

Summary: 0 passed, 4 failed, 3 skipped, 11 dry-run only
Dry run: nothing was sent. Add --send to send real, billed messages.
```

A real dry run with a valid key shows `PASS` for steps 1 to 4, followed by the `DRY` lines and a
summary. A live run (`--send`) shows `PASS` or `FAIL` for each send, then the status of each
sent message in step 10. Example:

```text
PASS 1 ping: status ok
PASS 2 find phone number: waba_id 1234567890, can_send true
...
PASS 5 send template: hello_id 0a1b2c3d-...
PASS 10 status text: status sent
...
Summary: 14 passed, 0 failed, 2 skipped, 0 dry-run only
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
- `TEMPLATE_NOT_APPROVED` or `no APPROVED template`: approve a template first, or set
  `HELLO_TEMPLATE_NAME`.
- `OUTSIDE_24H_WINDOW` on steps 6 to 9: see the 24-hour note above.
- `step 6 returned no wamid` (reaction skipped): the text was queued rather than accepted within
  25 seconds, or it failed. Check step 10.
- `TEMPLATE_LANGUAGE_REQUIRED`: several languages share the name. Set `HELLO_TEMPLATE_LANGUAGE`.

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
