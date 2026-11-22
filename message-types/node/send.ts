// Send one WhatsApp message type and follow its delivery status.
// Usage: npm start -- <type>     (npm start -- all lists the types)
import { Hello } from "@lakshyapocketlab/hello";

type Send = (hello: Hello, from: string, to: string) => ReturnType<Hello["messages"]["send"]>;

/** Reads a required environment variable or exits with a clear message. */
function need(name: string): string {
  const value = process.env[name];
  if (!value) {
    console.error(`${name} is not set (see .env.example)`);
    process.exit(1);
  }
  return value;
}

const language = () => process.env.HELLO_TEMPLATE_LANGUAGE || "en_US";
const text = (value: string) => ({ type: "text", text: value });

// ---------------------------------------------------------------------------
// Free-form messages: need an open 24-hour window (the recipient wrote to you
// in the last 24 hours). Bodies follow https://lakshyapocketlab.com/docs/messages/
// ---------------------------------------------------------------------------

// Text, up to 4,096 characters. preview_url shows a preview of the first URL.
const sendText: Send = (hello, from, to) =>
  hello.messages.sendText(
    from,
    to,
    "Your table for 4 is confirmed for 8 pm. Menu: https://lakshyapocketlab.com/docs/",
    { previewUrl: true },
  );

// Image (JPEG or PNG, up to 5 MB) by public HTTPS link, with a caption.
const sendImage: Send = (hello, from, to) =>
  hello.messages.sendMedia(from, to, {
    type: "image",
    link: need("HELLO_IMAGE_URL"),
    caption: "The Runner 2 in red, back in stock.",
  });

// Video (MP4 or 3GPP, up to 16 MB) by public HTTPS link, with a caption.
const sendVideo: Send = (hello, from, to) =>
  hello.messages.sendMedia(from, to, {
    type: "video",
    link: need("HELLO_VIDEO_URL"),
    caption: "Setting up your device in 60 seconds",
  });

// Audio (up to 16 MB). WhatsApp does not allow captions on audio.
const sendAudio: Send = (hello, from, to) =>
  hello.messages.sendMedia(from, to, { type: "audio", link: need("HELLO_AUDIO_URL") });

// Document (PDF, Office or text, up to 100 MB) with the file name the recipient sees.
const sendDocument: Send = (hello, from, to) =>
  hello.messages.sendMedia(from, to, {
    type: "document",
    link: need("HELLO_DOCUMENT_URL"),
    filename: "Invoice INV-2291.pdf",
    caption: "Your invoice for September",
  });

// Sticker: static WebP up to 100 KB. Stickers have no link option: upload the
// file first (hello.media.upload) and pass the media id.
const sendSticker: Send = (hello, from, to) =>
  hello.messages.sendMedia(from, to, { type: "sticker", id: need("HELLO_STICKER_MEDIA_ID") });

// Location: latitude and longitude are numbers here (strings in a template header).
const sendLocation: Send = (hello, from, to) =>
  hello.messages.send(from, {
    messaging_product: "whatsapp",
    to,
    type: "location",
    location: {
      latitude: 12.9716,
      longitude: 77.5946,
      name: "Indiranagar store",
      address: "100 Feet Rd, Indiranagar, Bengaluru 560038",
    },
  });

// Contacts: `contacts` is an array at the top level of the request, not an object.
const sendContacts: Send = (hello, from, to) =>
  hello.messages.send(from, {
    messaging_product: "whatsapp",
    to,
    type: "contacts",
    contacts: [
      {
        name: { formatted_name: "Priya Sharma", first_name: "Priya", last_name: "Sharma" },
        phones: [{ phone: "+919812345678", type: "WORK" }],
      },
    ],
  });

// Interactive reply buttons: up to 3 buttons, titles up to 20 characters. A tap
// arrives as an inbound message carrying the button's id and title.
const sendButtons: Send = (hello, from, to) =>
  hello.messages.send(from, {
    messaging_product: "whatsapp",
    to,
    type: "interactive",
    interactive: {
      type: "button",
      header: { type: "text", text: "Booking" },
      body: { text: "Confirm your booking for tomorrow 10:00?" },
      footer: { text: "Reply anytime" },
      action: {
        buttons: [
          { type: "reply", reply: { id: "yes", title: "Yes" } },
          { type: "reply", reply: { id: "no", title: "No" } },
        ],
      },
    },
  });

// Interactive list: one button opens up to 10 rows in total, grouped in sections.
const sendList: Send = (hello, from, to) =>
  hello.messages.send(from, {
    messaging_product: "whatsapp",
    to,
    type: "interactive",
    interactive: {
      type: "list",
      header: { type: "text", text: "Appointments" },
      body: { text: "Pick a slot" },
      footer: { text: "Times are IST" },
      action: {
        button: "View slots", // up to 20 characters
        sections: [
          {
            title: "Monday",
            rows: [
              { id: "mon-9", title: "09:00", description: "Morning" },
              { id: "mon-14", title: "14:00", description: "Afternoon" },
            ],
          },
          { title: "Tuesday", rows: [{ id: "tue-11", title: "11:00" }] },
        ],
      },
    },
  });

// Interactive call-to-action URL: one button that opens a link. Unlike a
// template URL button it needs no pre-approval.
const sendCtaUrl: Send = (hello, from, to) =>
  hello.messages.send(from, {
    messaging_product: "whatsapp",
    to,
    type: "interactive",
    interactive: {
      type: "cta_url",
      header: { type: "text", text: "Your order" },
      body: { text: "Track your order in real time" },
      footer: { text: "Thanks for shopping" },
      action: {
        name: "cta_url",
        parameters: { display_text: "Track order", url: "https://example.com/track/10432" },
      },
    },
  });

// Reaction: one emoji on a message the customer sent (its wamid). An empty
// emoji removes the reaction.
const sendReaction: Send = (hello, from, to) =>
  hello.messages.send(from, {
    messaging_product: "whatsapp",
    to,
    type: "reaction",
    reaction: { message_id: need("HELLO_REPLY_TO_MESSAGE_ID"), emoji: "👍" },
  });

// Reply: any message type can quote an earlier message by adding
// context.message_id. The helpers take it as the replyTo option.
const sendReply: Send = (hello, from, to) =>
  hello.messages.sendText(from, to, "Yes, that one is back in stock.", {
    replyTo: need("HELLO_REPLY_TO_MESSAGE_ID"),
  });

// ---------------------------------------------------------------------------
// Template messages: the only type that can start a conversation (no 24-hour
// window needed). The template must exist on the number's WABA and be APPROVED.
// ---------------------------------------------------------------------------

// Simple template. Body variables come from HELLO_TEMPLATE_PARAMS, comma separated:
// positional values ("Asha,#10432") or named ones ("first_name=Asha,otp=482913").
// The default, hello_world, has no variables.
const sendTemplate: Send = (hello, from, to) => {
  const values = (process.env.HELLO_TEMPLATE_PARAMS ?? "").split(",").filter(Boolean);
  const parameters = values.map((value) => {
    const [name, ...rest] = value.split("=");
    return rest.length ? { ...text(rest.join("=")), parameter_name: name } : text(value);
  });
  return hello.messages.sendTemplate(from, to, {
    name: process.env.HELLO_TEMPLATE_NAME || "hello_world",
    language: language(),
    components: parameters.length ? [{ type: "body", parameters }] : undefined,
  });
};

// Template with a text header variable and a footer. The footer is fixed in the
// approved template, so only the header and body variables are sent.
const sendTemplateHeaderFooter: Send = (hello, from, to) =>
  hello.messages.sendTemplate(from, to, {
    name: need("HELLO_TEMPLATE_HEADER_FOOTER_NAME"),
    language: language(),
    components: [
      { type: "header", parameters: [text("Asha")] },
      { type: "body", parameters: [text("Dr. Rao"), text("3 October"), text("10:30 am")] },
    ],
  });

// Template with a dynamic URL button. `index` is the button's position in the
// template's BUTTONS array, as a string. Quick-reply, fixed-URL and call buttons
// need nothing at send time: send only the body.
const sendTemplateButtons: Send = (hello, from, to) =>
  hello.messages.sendTemplate(from, to, {
    name: need("HELLO_TEMPLATE_BUTTONS_NAME"),
    language: language(),
    components: [
      { type: "body", parameters: [text("Asha")] },
      { type: "button", sub_type: "url", index: "0", parameters: [text("ORD-10432")] },
    ],
  });

// Template with a media header (here an image by public link) plus body variables.
// Every send must supply the actual media; the footer and buttons come from the template.
const sendTemplateMedia: Send = (hello, from, to) =>
  hello.messages.sendTemplate(from, to, {
    name: need("HELLO_TEMPLATE_MEDIA_NAME"),
    language: language(),
    components: [
      { type: "header", parameters: [{ type: "image", image: { link: need("HELLO_IMAGE_URL") } }] },
      { type: "body", parameters: [text("Asha"), text("Runner 2")] },
    ],
  });

const types: Record<string, Send> = {
  text: sendText,
  image: sendImage,
  video: sendVideo,
  audio: sendAudio,
  document: sendDocument,
  sticker: sendSticker,
  location: sendLocation,
  contacts: sendContacts,
  buttons: sendButtons,
  list: sendList,
  "cta-url": sendCtaUrl,
  reaction: sendReaction,
  reply: sendReply,
  template: sendTemplate,
  "template-header-footer": sendTemplateHeaderFooter,
  "template-buttons": sendTemplateButtons,
  "template-media": sendTemplateMedia,
};

// ---------------------------------------------------------------------------
// Delivery tracking: print the Hello ID, then poll messages.get for ~30 s and
// print every status change (queued, sent, delivered, read or failed).
// ---------------------------------------------------------------------------

async function track(hello: Hello, helloId: string): Promise<void> {
  const deadline = Date.now() + 30_000;
  let last = "";
  while (Date.now() < deadline) {
    const msg = await hello.messages.get(helloId);
    if (msg.status !== last) {
      last = msg.status;
      const reason = msg.status === "failed" ? ` (${msg.error_code ?? ""} ${msg.error_message ?? ""})` : "";
      console.log(`status: ${msg.status}${reason}`);
    }
    if (["delivered", "read", "failed"].includes(msg.status)) return;
    await new Promise((resolve) => setTimeout(resolve, 2000));
  }
  console.log(`still ${last} after 30 s; check later with messages.get("${helloId}")`);
}

async function main(): Promise<void> {
  const type = process.argv[2];
  if (type === "all") {
    console.log("message types:\n  " + Object.keys(types).join("\n  "));
    return;
  }
  if (!process.env.HELLO_API_KEY) {
    console.error("HELLO_API_KEY is not set (see .env.example)");
    process.exit(1);
  }
  const send = type ? types[type] : undefined;
  if (!send) {
    console.error(`usage: npm start -- <type>\ntypes: ${Object.keys(types).join(", ")}`);
    process.exit(1);
  }
  const from = need("HELLO_PHONE_NUMBER_ID");
  const to = need("HELLO_TO");

  const hello = new Hello(); // reads HELLO_API_KEY
  const res = await send(hello, from, to);
  const helloId = res.messages[0]?.hello_id;
  console.log("accepted, Hello ID:", helloId);
  if (helloId) await track(hello, helloId);
}

await main();
