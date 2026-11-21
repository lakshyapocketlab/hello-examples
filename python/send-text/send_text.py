"""Send a text message. Usage: python send_text.py <to> [text]

Free-form text needs an open 24-hour window: the recipient must have messaged you in the last
24 hours. Use a template to start a conversation (see ../../node/send-template).
"""

from __future__ import annotations

import os
import sys

from lakshyapocketlab_hello import Hello


def main(to: str, text: str) -> None:
    with Hello() as client:  # reads HELLO_API_KEY
        phone_number_id = os.environ.get("HELLO_PHONE_NUMBER_ID")
        if not phone_number_id:
            numbers = client.phone_numbers.list()["data"]
            if not numbers:
                raise SystemExit("No phone numbers connected to this Hello account")
            phone_number_id = numbers[0]["phone_number_id"]

        sent = client.messages.send_text(phone_number_id, to, text, sync=True)
        print("accepted, message id:", sent["messages"][0]["id"])


if __name__ == "__main__":
    if len(sys.argv) < 2:
        raise SystemExit("usage: python send_text.py <to> [text]")
    main(sys.argv[1], sys.argv[2] if len(sys.argv) > 2 else "Hello from the Hello Python SDK!")
