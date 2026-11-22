"""Send one WhatsApp message type and follow its delivery status.

Usage: python send.py <type>     (python send.py all lists the types)
"""

from __future__ import annotations

import os
import sys
import time
from typing import Any, Callable, Dict

from lakshyapocketlab_hello import Hello


def need(name: str) -> str:
    """Read a required environment variable or exit with a clear message."""
    value = os.environ.get(name)
    if not value:
        raise SystemExit(f"{name} is not set (see .env.example)")
    return value


def language() -> str:
    return os.environ.get("HELLO_TEMPLATE_LANGUAGE") or "en_US"


def text(value: str) -> Dict[str, str]:
    return {"type": "text", "text": value}


# ---------------------------------------------------------------------------
# Free-form messages: need an open 24-hour window (the recipient wrote to you
# in the last 24 hours). Bodies follow https://lakshyapocketlab.com/docs/messages/
# ---------------------------------------------------------------------------


def send_text(client: Hello, sender: str, to: str) -> Dict[str, Any]:
    # Text, up to 4,096 characters. preview_url shows a preview of the first URL.
    return client.messages.send_text(
        sender,
        to,
        "Your table for 4 is confirmed for 8 pm. Menu: https://lakshyapocketlab.com/docs/",
        preview_url=True,
    )


def send_image(client: Hello, sender: str, to: str) -> Dict[str, Any]:
    # Image (JPEG or PNG, up to 5 MB) by public HTTPS link, with a caption.
    return client.messages.send_media(
        sender,
        to,
        "image",
        link=need("HELLO_IMAGE_URL"),
        caption="The Runner 2 in red, back in stock.",
    )


def send_video(client: Hello, sender: str, to: str) -> Dict[str, Any]:
    # Video (MP4 or 3GPP, up to 16 MB) by public HTTPS link, with a caption.
    return client.messages.send_media(
        sender,
        to,
        "video",
        link=need("HELLO_VIDEO_URL"),
        caption="Setting up your device in 60 seconds",
    )


def send_audio(client: Hello, sender: str, to: str) -> Dict[str, Any]:
    # Audio (up to 16 MB). WhatsApp does not allow captions on audio.
    return client.messages.send_media(sender, to, "audio", link=need("HELLO_AUDIO_URL"))


def send_document(client: Hello, sender: str, to: str) -> Dict[str, Any]:
    # Document (PDF, Office or text, up to 100 MB) with the file name the recipient sees.
    return client.messages.send_media(
        sender,
        to,
        "document",
        link=need("HELLO_DOCUMENT_URL"),
        filename="Invoice INV-2291.pdf",
        caption="Your invoice for September",
    )


def send_sticker(client: Hello, sender: str, to: str) -> Dict[str, Any]:
    # Sticker: static WebP up to 100 KB. Stickers have no link option: upload the
    # file first (client.media.upload) and pass the media id.
    return client.messages.send_media(sender, to, "sticker", media_id=need("HELLO_STICKER_MEDIA_ID"))


def send_location(client: Hello, sender: str, to: str) -> Dict[str, Any]:
    # Location: latitude and longitude are numbers here (strings in a template header).
    return client.messages.send(
        sender,
        {
            "messaging_product": "whatsapp",
            "to": to,
            "type": "location",
            "location": {
                "latitude": 12.9716,
                "longitude": 77.5946,
                "name": "Indiranagar store",
                "address": "100 Feet Rd, Indiranagar, Bengaluru 560038",
            },
        },
    )


def send_contacts(client: Hello, sender: str, to: str) -> Dict[str, Any]:
    # Contacts: `contacts` is an array at the top level of the request, not an object.
    return client.messages.send(
        sender,
        {
            "messaging_product": "whatsapp",
            "to": to,
            "type": "contacts",
            "contacts": [
                {
                    "name": {
                        "formatted_name": "Priya Sharma",
                        "first_name": "Priya",
                        "last_name": "Sharma",
                    },
                    "phones": [{"phone": "+919812345678", "type": "WORK"}],
                }
            ],
        },
    )


def send_buttons(client: Hello, sender: str, to: str) -> Dict[str, Any]:
    # Interactive reply buttons: up to 3 buttons, titles up to 20 characters. A tap
    # arrives as an inbound message carrying the button's id and title.
    return client.messages.send(
        sender,
        {
            "messaging_product": "whatsapp",
            "to": to,
            "type": "interactive",
            "interactive": {
                "type": "button",
                "header": {"type": "text", "text": "Booking"},
                "body": {"text": "Confirm your booking for tomorrow 10:00?"},
                "footer": {"text": "Reply anytime"},
                "action": {
                    "buttons": [
                        {"type": "reply", "reply": {"id": "yes", "title": "Yes"}},
                        {"type": "reply", "reply": {"id": "no", "title": "No"}},
                    ]
                },
            },
        },
    )


def send_list(client: Hello, sender: str, to: str) -> Dict[str, Any]:
    # Interactive list: one button opens up to 10 rows in total, grouped in sections.
    return client.messages.send(
        sender,
        {
            "messaging_product": "whatsapp",
            "to": to,
            "type": "interactive",
            "interactive": {
                "type": "list",
                "header": {"type": "text", "text": "Appointments"},
                "body": {"text": "Pick a slot"},
                "footer": {"text": "Times are IST"},
                "action": {
                    "button": "View slots",  # up to 20 characters
                    "sections": [
                        {
                            "title": "Monday",
                            "rows": [
                                {"id": "mon-9", "title": "09:00", "description": "Morning"},
                                {"id": "mon-14", "title": "14:00", "description": "Afternoon"},
                            ],
                        },
                        {"title": "Tuesday", "rows": [{"id": "tue-11", "title": "11:00"}]},
                    ],
                },
            },
        },
    )


def send_cta_url(client: Hello, sender: str, to: str) -> Dict[str, Any]:
    # Interactive call-to-action URL: one button that opens a link. Unlike a
    # template URL button it needs no pre-approval.
    return client.messages.send(
        sender,
        {
            "messaging_product": "whatsapp",
            "to": to,
            "type": "interactive",
            "interactive": {
                "type": "cta_url",
                "header": {"type": "text", "text": "Your order"},
                "body": {"text": "Track your order in real time"},
                "footer": {"text": "Thanks for shopping"},
                "action": {
                    "name": "cta_url",
                    "parameters": {
                        "display_text": "Track order",
                        "url": "https://example.com/track/10432",
                    },
                },
            },
        },
    )


def send_reaction(client: Hello, sender: str, to: str) -> Dict[str, Any]:
    # Reaction: one emoji on a message the customer sent (its wamid). An empty
    # emoji removes the reaction.
    return client.messages.send(
        sender,
        {
            "messaging_product": "whatsapp",
            "to": to,
            "type": "reaction",
            "reaction": {"message_id": need("HELLO_REPLY_TO_MESSAGE_ID"), "emoji": "👍"},
        },
    )


def send_reply(client: Hello, sender: str, to: str) -> Dict[str, Any]:
    # Reply: any message type can quote an earlier message by adding
    # context.message_id. The helpers take it as reply_to.
    return client.messages.send_text(
        sender,
        to,
        "Yes, that one is back in stock.",
        reply_to=need("HELLO_REPLY_TO_MESSAGE_ID"),
    )


# ---------------------------------------------------------------------------
# Template messages: the only type that can start a conversation (no 24-hour
# window needed). The template must exist on the number's WABA and be APPROVED.
# ---------------------------------------------------------------------------


def send_template(client: Hello, sender: str, to: str) -> Dict[str, Any]:
    # Simple template. Body variables come from HELLO_TEMPLATE_PARAMS, comma separated:
    # positional values ("Asha,#10432") or named ones ("first_name=Asha,otp=482913").
    # The default, hello_world, has no variables.
    parameters = []
    for value in filter(None, os.environ.get("HELLO_TEMPLATE_PARAMS", "").split(",")):
        name, sep, rest = value.partition("=")
        parameters.append({**text(rest), "parameter_name": name} if sep else text(value))
    return client.messages.send_template(
        sender,
        to,
        os.environ.get("HELLO_TEMPLATE_NAME") or "hello_world",
        language(),
        [{"type": "body", "parameters": parameters}] if parameters else None,
    )


def send_template_header_footer(client: Hello, sender: str, to: str) -> Dict[str, Any]:
    # Template with a text header variable and a footer. The footer is fixed in the
    # approved template, so only the header and body variables are sent.
    return client.messages.send_template(
        sender,
        to,
        need("HELLO_TEMPLATE_HEADER_FOOTER_NAME"),
        language(),
        [
            {"type": "header", "parameters": [text("Asha")]},
            {
                "type": "body",
                "parameters": [text("Dr. Rao"), text("3 October"), text("10:30 am")],
            },
        ],
    )


def send_template_buttons(client: Hello, sender: str, to: str) -> Dict[str, Any]:
    # Template with a dynamic URL button. `index` is the button's position in the
    # template's BUTTONS array, as a string. Quick-reply, fixed-URL and call buttons
    # need nothing at send time: send only the body.
    return client.messages.send_template(
        sender,
        to,
        need("HELLO_TEMPLATE_BUTTONS_NAME"),
        language(),
        [
            {"type": "body", "parameters": [text("Asha")]},
            {"type": "button", "sub_type": "url", "index": "0", "parameters": [text("ORD-10432")]},
        ],
    )


def send_template_media(client: Hello, sender: str, to: str) -> Dict[str, Any]:
    # Template with a media header (here an image by public link) plus body variables.
    # Every send must supply the actual media; the footer and buttons come from the template.
    return client.messages.send_template(
        sender,
        to,
        need("HELLO_TEMPLATE_MEDIA_NAME"),
        language(),
        [
            {
                "type": "header",
                "parameters": [{"type": "image", "image": {"link": need("HELLO_IMAGE_URL")}}],
            },
            {"type": "body", "parameters": [text("Asha"), text("Runner 2")]},
        ],
    )


TYPES: Dict[str, Callable[[Hello, str, str], Dict[str, Any]]] = {
    "text": send_text,
    "image": send_image,
    "video": send_video,
    "audio": send_audio,
    "document": send_document,
    "sticker": send_sticker,
    "location": send_location,
    "contacts": send_contacts,
    "buttons": send_buttons,
    "list": send_list,
    "cta-url": send_cta_url,
    "reaction": send_reaction,
    "reply": send_reply,
    "template": send_template,
    "template-header-footer": send_template_header_footer,
    "template-buttons": send_template_buttons,
    "template-media": send_template_media,
}


# ---------------------------------------------------------------------------
# Delivery tracking: print the Hello ID, then poll messages.get for ~30 s and
# print every status change (queued, sent, delivered, read or failed).
# ---------------------------------------------------------------------------


def track(client: Hello, hello_id: str) -> None:
    deadline = time.monotonic() + 30
    last = ""
    while time.monotonic() < deadline:
        msg = client.messages.get(hello_id)
        status = msg["status"]
        if status != last:
            last = status
            reason = f" ({msg.get('error_code', '')} {msg.get('error_message', '')})" if status == "failed" else ""
            print(f"status: {status}{reason}")
        if status in ("delivered", "read", "failed"):
            return
        time.sleep(2)
    print(f'still {last} after 30 s; check later with messages.get("{hello_id}")')


def main() -> None:
    kind = sys.argv[1] if len(sys.argv) > 1 else ""
    if kind == "all":
        print("message types:\n  " + "\n  ".join(TYPES))
        return
    if not os.environ.get("HELLO_API_KEY"):
        raise SystemExit("HELLO_API_KEY is not set (see .env.example)")
    send = TYPES.get(kind)
    if send is None:
        raise SystemExit(f"usage: python send.py <type>\ntypes: {', '.join(TYPES)}")
    sender = need("HELLO_PHONE_NUMBER_ID")
    to = need("HELLO_TO")

    with Hello() as client:  # reads HELLO_API_KEY
        res = send(client, sender, to)
        hello_id = res["messages"][0].get("hello_id")
        print("accepted, Hello ID:", hello_id)
        if hello_id:
            track(client, hello_id)


if __name__ == "__main__":
    main()
