// Receives WhatsApp Cloud API webhooks for a number that uses Hello's "direct webhook".
//
// How Hello's direct webhook works (https://lakshyapocketlab.com/docs/webhooks/direct-webhook/):
//   1. You set a URL and a verify token for one phone number. Meta calls the URL once with
//      GET ?hub.mode=subscribe&hub.verify_token=...&hub.challenge=... and expects the challenge
//      echoed back as plain text.
//   2. Meta then POSTs events (messages, statuses) straight to your URL.
//   3. Meta signs POSTs with X-Hub-Signature-256 using HELLO's app secret, which is not shared
//      with you. You therefore cannot verify that header. Hello's documented way to authenticate
//      deliveries is a long random token in the URL path, which this example checks.
import { createHmac, timingSafeEqual } from "node:crypto";
import express from "express";

const PORT = Number(process.env.PORT ?? 3000);
const VERIFY_TOKEN = required("WHATSAPP_VERIFY_TOKEN");
const PATH_SECRET = required("WEBHOOK_PATH_SECRET");
// Optional. Only set it if the number's webhook is signed with an app secret YOU own (for
// example your own Meta app). With Hello's direct webhook, leave it unset.
const APP_SECRET = process.env.WHATSAPP_APP_SECRET;

function required(name: string): string {
  const value = process.env[name];
  if (!value) {
    console.error(`${name} is required`);
    process.exit(1);
  }
  return value;
}

function safeEqual(a: string, b: string): boolean {
  const x = Buffer.from(a);
  const y = Buffer.from(b);
  return x.length === y.length && timingSafeEqual(x, y);
}

/** Meta's scheme: header is "sha256=" + hex HMAC-SHA256 of the raw body keyed with the app secret. */
export function validSignature(rawBody: Buffer, header: string | undefined, secret: string): boolean {
  if (!header?.startsWith("sha256=")) return false;
  const expected = createHmac("sha256", secret).update(rawBody).digest("hex");
  return safeEqual(header.slice("sha256=".length), expected);
}

const app = express();
const route = `/hooks/whatsapp/${PATH_SECRET}`;

// 1. Handshake: echo hub.challenge if the verify token matches.
app.get(route, (req, res) => {
  const { "hub.mode": mode, "hub.verify_token": token, "hub.challenge": challenge } = req.query;
  if (mode === "subscribe" && typeof token === "string" && safeEqual(token, VERIFY_TOKEN)) {
    res.status(200).type("text/plain").send(String(challenge));
    return;
  }
  res.sendStatus(403);
});

// 2. Events. express.raw keeps the exact bytes, which a signature check needs.
app.post(route, express.raw({ type: "application/json", limit: "4mb" }), (req, res) => {
  const raw = req.body as Buffer;
  if (APP_SECRET && !validSignature(raw, req.header("x-hub-signature-256"), APP_SECRET)) {
    res.sendStatus(401);
    return;
  }
  res.sendStatus(200); // answer quickly; Meta retries slow or failed deliveries

  const payload = JSON.parse(raw.toString("utf8"));
  for (const entry of payload.entry ?? []) {
    for (const change of entry.changes ?? []) {
      const value = change.value ?? {};
      for (const msg of value.messages ?? []) {
        console.log("incoming message", msg.from, msg.type, msg.text?.body ?? "");
      }
      for (const status of value.statuses ?? []) {
        console.log("status", status.id, status.status);
      }
    }
  }
});

// Anything else, including a wrong path secret, is a plain 404.
app.listen(PORT, () => console.log(`listening on :${PORT}, webhook path ${route.replace(PATH_SECRET, "<secret>")}`));
