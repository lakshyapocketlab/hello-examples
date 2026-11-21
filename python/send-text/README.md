# Send a text message with Python

Sends a WhatsApp text message through the [Hello Python SDK](https://github.com/lakshyapocketlab/hello-python).

## Prerequisites

- Python 3.9 or newer
- A Hello API key (`hello_live_...`) and a connected WhatsApp number
- An open 24-hour window with the recipient (they messaged you in the last 24 hours). Otherwise send a template instead.

## Environment variables

| Variable | Required | Description |
|---|---|---|
| `HELLO_API_KEY` | yes | Your Hello API key |
| `HELLO_PHONE_NUMBER_ID` | no | Number to send from. Defaults to the first connected number |

## Run

```sh
python -m venv .venv && source .venv/bin/activate
pip install -r requirements.txt
HELLO_API_KEY=hello_live_xxx python send_text.py +919999999999 "Hello!"
```

The SDK is not on PyPI yet. Until it is, install it from GitHub, see the [root README](../../README.md#installing-the-sdks-before-they-are-published).
