"""Live smoke test for the lakshyapocketlab-hello SDK against the real Hello API.

Steps 1-4 are read-only. Steps 5-11 send real, billed WhatsApp messages and run
only with --send. Every step prints PASS, FAIL, SKIP or DRY; the run continues
after a failure. See ../README.md.

Usage:
    python smoke_test.py          # dry run
    python smoke_test.py --send   # live
"""

import argparse
import os
import re
import sys
import time
import urllib.request
from typing import Any, Callable, Dict, List, Optional

from lakshyapocketlab_hello import APIError, Hello

DEFAULT_IMAGE_URL = "https://upload.wikimedia.org/wikipedia/commons/4/47/PNG_transparency_demonstration_1.png"
DEFAULT_DOCUMENT_URL = "https://www.w3.org/WAI/ER/tests/xhtml/testfiles/resources/pdf/dummy.pdf"
THUMBS_UP = "\U0001F44D"

Template = Dict[str, Any]
tally: Dict[str, int] = {"PASS": 0, "FAIL": 0, "SKIP": 0, "DRY": 0}
sent: List[Dict[str, Optional[str]]] = []
state: Dict[str, Any] = {}


def env(name: str) -> Optional[str]:
    value = os.environ.get(name, "").strip()
    return value or None


def line(status: str, name: str, detail: str = "") -> None:
    tally[status] += 1
    print(f"{status:<4} {name}" + (f": {detail}" if detail else ""))


def describe(err: Exception) -> str:
    if isinstance(err, APIError):
        code = f" {err.hello_code}" if err.hello_code else ""
        return f"HTTP {err.status}{code}: {err.message}"
    return str(err)


def step(name: str, fn: Callable[[], str]) -> None:
    try:
        line("PASS", name, fn())
    except Exception as err:  # keep going after a failure
        line("FAIL", name, describe(err))


def send_step(name: str, would_do: str, fn: Callable[[], str], send: bool) -> None:
    if not send:
        line("DRY", name, f"would {would_do}")
        return
    step(name, fn)


def record(label: str, res: Dict[str, Any]) -> str:
    message = res["messages"][0]
    hello_id = message.get("hello_id")
    sent.append({"label": label, "hello_id": hello_id, "wamid": message.get("id") or None})
    return f"hello_id {hello_id}"


def body_text(t: Template) -> str:
    for component in t.get("components") or []:
        if component.get("type") == "BODY":
            return component.get("text") or ""
    return ""


def variable_count(t: Template) -> int:
    indexes = [int(n) for n in re.findall(r"\{\{(\d+)\}\}", body_text(t))]
    return max(indexes) if indexes else 0


def usable(t: Template) -> bool:
    """Usable here when the only placeholders are body variables."""
    others = [c for c in t.get("components") or [] if c.get("type") != "BODY"]
    return not any("{{" in str(c) for c in others)


def pick_template(templates: List[Template], wanted: Optional[str], wanted_language: Optional[str]) -> Optional[Template]:
    if wanted:
        named = [
            t
            for t in templates
            if t.get("name") == wanted and (not wanted_language or t.get("language") == wanted_language) and usable(t)
        ]
        approved = [t for t in named if t.get("status") == "APPROVED"]
        return (approved or named or [None])[0]
    approved_all = [t for t in templates if t.get("status") == "APPROVED" and usable(t)]
    no_variables = [t for t in approved_all if variable_count(t) == 0]
    return (no_variables or approved_all or [None])[0]


def body_components(t: Template) -> Optional[List[Dict[str, Any]]]:
    n = variable_count(t)
    if n == 0:
        return None
    parameters = [{"type": "text", "text": f"Sample {i}"} for i in range(1, n + 1)]
    return [{"type": "body", "parameters": parameters}]


def download(url: str) -> "tuple[bytes, str]":
    request = urllib.request.Request(url, headers={"User-Agent": "hello-smoke-test/1.0"})
    with urllib.request.urlopen(request, timeout=60) as resp:
        return resp.read(), resp.headers.get_content_type()


def run(send: bool, run_id: str, client: Hello) -> None:
    to = env("HELLO_TO") or "<HELLO_TO>"
    pnid = env("HELLO_PHONE_NUMBER_ID") or ""
    template_name = env("HELLO_TEMPLATE_NAME")
    template_language = env("HELLO_TEMPLATE_LANGUAGE")
    image_url = env("HELLO_MEDIA_IMAGE_URL") or DEFAULT_IMAGE_URL
    document_url = env("HELLO_MEDIA_DOCUMENT_URL") or DEFAULT_DOCUMENT_URL
    video_url = env("HELLO_MEDIA_VIDEO_URL")
    audio_url = env("HELLO_MEDIA_AUDIO_URL")
    inbound_id = env("HELLO_INBOUND_MESSAGE_ID")

    # 1. ping
    def ping() -> str:
        res = client.ping()
        if res.get("status") != "ok":
            raise RuntimeError(f"unexpected status {res.get('status')}")
        return f"status {res['status']}"

    step("1 ping", ping)

    # 2. phone number and WABA
    def find_number() -> str:
        numbers = client.phone_numbers.list()["data"]
        match = next((n for n in numbers if n.get("phone_number_id") == pnid), None)
        if match is None:
            raise RuntimeError(f"no number with phone_number_id {pnid} among {len(numbers)} on this key")
        state["waba"] = match["waba_id"]
        return f"waba_id {match['waba_id']}, can_send {match.get('can_send')}"

    step("2 find phone number", find_number)

    # 3. list templates and pick one
    def list_and_pick() -> str:
        if not state.get("waba"):
            raise RuntimeError("no WABA id from step 2")
        templates = client.templates.list(state["waba"], limit=100)["data"]
        for t in templates:
            print(f"     {t.get('name')}  {t.get('language')}  {t.get('status')}  {t.get('category')}  "
                  f"variables={variable_count(t)}")
        picked = pick_template(templates, template_name, template_language)
        if picked is None:
            if template_name:
                suffix = f" in {template_language}" if template_language else ""
                raise RuntimeError(f"no usable template named {template_name}{suffix}")
            raise RuntimeError("no APPROVED template this smoke test can fill in")
        state["template"] = picked
        return f"using {picked['name']} ({picked['language']}, {picked['status']}, {variable_count(picked)} variables)"

    step("3 list templates and pick one", list_and_pick)

    # 4. get template by id
    def get_template() -> str:
        t = state.get("template")
        if t is None:
            raise RuntimeError("no template from step 3")
        got = client.templates.get(t["id"])
        if got.get("id") != t["id"]:
            raise RuntimeError(f"asked for {t['id']}, got {got.get('id')}")
        return f"{got['name']} {got['language']} {got['status']}"

    step("4 get template by id", get_template)

    # 5-11 send or write; each one is gated by --send.

    def template_name_for_dry() -> str:
        t = state.get("template")
        return t["name"] if t else "?"

    def send_template() -> str:
        t = state.get("template")
        if t is None:
            raise RuntimeError("no template from step 3")
        res = client.messages.send_template(
            pnid,
            to,
            t["name"],
            t["language"],
            components=body_components(t),
            idempotency_key=f"{run_id}-template",
        )
        return record("template", res)

    send_step("5 send template", f"send template {template_name_for_dry()} to {to}", send_template, send)

    def send_text() -> str:
        res = client.messages.send_text(
            pnid,
            to,
            f"Hello smoke test {run_id}: text message.",
            sync=True,
            idempotency_key=f"{run_id}-text",
        )
        state["text_wamid"] = res["messages"][0].get("id") or None
        return record("text", res)

    send_step("6 send text", f"send a text message to {to}", send_text, send)

    send_step(
        "7 send image by link",
        f"send the image at {image_url}",
        lambda: record(
            "image-link",
            client.messages.send_media(
                pnid,
                to,
                "image",
                link=image_url,
                caption=f"Smoke test image {run_id}",
                idempotency_key=f"{run_id}-image-link",
            ),
        ),
        send,
    )
    send_step(
        "7 send document by link",
        f"send the document at {document_url}",
        lambda: record(
            "document-link",
            client.messages.send_media(
                pnid,
                to,
                "document",
                link=document_url,
                filename="smoke-test.pdf",
                caption=f"Smoke test document {run_id}",
                idempotency_key=f"{run_id}-document-link",
            ),
        ),
        send,
    )
    if video_url:
        send_step(
            "7 send video by link",
            f"send the video at {video_url}",
            lambda: record(
                "video-link",
                client.messages.send_media(
                    pnid,
                    to,
                    "video",
                    link=video_url,
                    caption=f"Smoke test video {run_id}",
                    idempotency_key=f"{run_id}-video-link",
                ),
            ),
            send,
        )
    else:
        line("SKIP", "7 send video by link", "HELLO_MEDIA_VIDEO_URL not set")
    if audio_url:
        send_step(
            "7 send audio by link",
            f"send the audio at {audio_url}",
            lambda: record(
                "audio-link",
                client.messages.send_media(
                    pnid,
                    to,
                    "audio",
                    link=audio_url,
                    idempotency_key=f"{run_id}-audio-link",
                ),
            ),
            send,
        )
    else:
        line("SKIP", "7 send audio by link", "HELLO_MEDIA_AUDIO_URL not set")

    def upload_and_send() -> str:
        data, mime = download(image_url)
        uploaded = client.media.upload(pnid, data, mime or "image/png", "smoke-test.png")
        res = client.messages.send_media(
            pnid,
            to,
            "image",
            media_id=uploaded["id"],
            caption=f"Smoke test uploaded image {run_id}",
            idempotency_key=f"{run_id}-image-id",
        )
        return f"media id {uploaded['id']}, {record('image-id', res)}"

    send_step(
        "8 upload image and send by media id",
        f"download {image_url}, upload it and send it by id",
        upload_and_send,
        send,
    )

    send_step(
        "9 send location",
        f"send a location to {to}",
        lambda: record(
            "location",
            client.messages.send(
                pnid,
                {
                    "messaging_product": "whatsapp",
                    "to": to,
                    "type": "location",
                    "location": {
                        "latitude": 28.6139,
                        "longitude": 77.209,
                        "name": "Smoke test location",
                        "address": "New Delhi, India",
                    },
                },
                idempotency_key=f"{run_id}-location",
            ),
        ),
        send,
    )
    send_step(
        "9 send contacts",
        f"send a contact card to {to}",
        lambda: record(
            "contacts",
            client.messages.send(
                pnid,
                {
                    "messaging_product": "whatsapp",
                    "to": to,
                    "type": "contacts",
                    "contacts": [
                        {
                            "name": {"formatted_name": "Smoke Test", "first_name": "Smoke", "last_name": "Test"},
                            "phones": [{"phone": "+15555550100", "type": "CELL", "wa_id": "15555550100"}],
                        }
                    ],
                },
                idempotency_key=f"{run_id}-contacts",
            ),
        ),
        send,
    )
    send_step(
        "9 send reply buttons",
        f"send reply buttons to {to}",
        lambda: record(
            "buttons",
            client.messages.send(
                pnid,
                {
                    "messaging_product": "whatsapp",
                    "to": to,
                    "type": "interactive",
                    "interactive": {
                        "type": "button",
                        "body": {"text": f"Smoke test {run_id}: pick one"},
                        "action": {
                            "buttons": [
                                {"type": "reply", "reply": {"id": "yes", "title": "Yes"}},
                                {"type": "reply", "reply": {"id": "no", "title": "No"}},
                            ]
                        },
                    },
                },
                idempotency_key=f"{run_id}-buttons",
            ),
        ),
        send,
    )
    send_step(
        "9 send list",
        f"send an interactive list to {to}",
        lambda: record(
            "list",
            client.messages.send(
                pnid,
                {
                    "messaging_product": "whatsapp",
                    "to": to,
                    "type": "interactive",
                    "interactive": {
                        "type": "list",
                        "body": {"text": f"Smoke test {run_id}: pick a row"},
                        "action": {
                            "button": "Choose",
                            "sections": [
                                {
                                    "title": "Smoke test",
                                    "rows": [
                                        {"id": "a", "title": "Option A"},
                                        {"id": "b", "title": "Option B"},
                                    ],
                                }
                            ],
                        },
                    },
                },
                idempotency_key=f"{run_id}-list",
            ),
        ),
        send,
    )

    def react() -> str:
        return record(
            "reaction",
            client.messages.send(
                pnid,
                {
                    "messaging_product": "whatsapp",
                    "to": to,
                    "type": "reaction",
                    "reaction": {"message_id": state["text_wamid"], "emoji": THUMBS_UP},
                },
                idempotency_key=f"{run_id}-reaction",
            ),
        )

    if not send:
        line("DRY", "9 react to text message", f"would react {THUMBS_UP} to the step 6 message")
    elif not state.get("text_wamid"):
        line("SKIP", "9 react to text message", "step 6 returned no wamid (send failed or was queued)")
    else:
        step("9 react to text message", react)

    # 10. status of each sent message
    if not send:
        line("DRY", "10 message status", "would read the status of each sent message after 5 s")
    elif not sent:
        line("SKIP", "10 message status", "nothing was sent")
    else:
        time.sleep(5)
        for item in sent:
            hello_id = item["hello_id"]
            if not hello_id:
                line("FAIL", f"10 status {item['label']}", "no hello_id returned")
                continue

            def status(hid: Optional[str] = hello_id) -> str:
                msg = client.messages.get(hid or "")
                if msg.get("status") == "failed":
                    raise RuntimeError(f"failed: {msg.get('error_code') or ''} {msg.get('error_message') or ''}".strip())
                return f"status {msg.get('status')}"

            step(f"10 status {item['label']}", status)

    # 11. mark an inbound message as read
    if not inbound_id:
        line("SKIP", "11 mark as read", "HELLO_INBOUND_MESSAGE_ID not set")
    elif not send:
        line("DRY", "11 mark as read", f"would mark {inbound_id} as read")
    else:

        def mark_read() -> str:
            res = client.messages.mark_as_read(pnid, inbound_id)
            if not res.get("success"):
                raise RuntimeError("API did not return success")
            return "marked read"

        step("11 mark as read", mark_read)


def main() -> int:
    parser = argparse.ArgumentParser(description="Live smoke test for the Hello Python SDK.")
    parser.add_argument("--send", action="store_true", help="send real, billed messages (default: dry run)")
    args = parser.parse_args()

    api_key = env("HELLO_API_KEY")
    if not api_key:
        print("HELLO_API_KEY is not set. Copy smoke-test/.env.example to .env, fill it in and load it. Nothing was run.")
        return 1
    if not env("HELLO_PHONE_NUMBER_ID"):
        print("HELLO_PHONE_NUMBER_ID is not set (see smoke-test/.env.example). Nothing was run.")
        return 1
    if args.send and not env("HELLO_TO"):
        print("HELLO_TO is not set. It is needed for --send. Nothing was run.")
        return 1

    run_id = f"smoke-{int(time.time() * 1000)}"
    mode = "LIVE, real messages will be sent" if args.send else "dry run, nothing is sent (add --send to send)"
    print(f"Hello smoke test, run {run_id}: {mode}")

    with Hello(api_key=api_key, base_url=env("HELLO_BASE_URL")) as client:
        run(args.send, run_id, client)

    print(
        f"\nSummary: {tally['PASS']} passed, {tally['FAIL']} failed, {tally['SKIP']} skipped, "
        f"{tally['DRY']} dry-run only"
    )
    if not args.send:
        print("Dry run: nothing was sent. Add --send to send real, billed messages.")
    return 1 if tally["FAIL"] else 0


if __name__ == "__main__":
    sys.exit(main())
