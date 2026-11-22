/**
 * Live smoke test for @lakshyapocketlab/hello against the real Hello API.
 *
 * Steps 1-4 are read-only. Steps 5-11 send real, billed WhatsApp messages and run
 * only with --send. Every step prints PASS, FAIL, SKIP or DRY; the run continues
 * after a failure. See ../README.md.
 */
import { APIError, Hello, type Template, type TemplateComponent } from "@lakshyapocketlab/hello";

const SEND = process.argv.includes("--send");
const RUN_ID = `smoke-${Date.now()}`;

const DEFAULT_IMAGE_URL = "https://upload.wikimedia.org/wikipedia/commons/4/47/PNG_transparency_demonstration_1.png";
const DEFAULT_DOCUMENT_URL = "https://www.w3.org/WAI/ER/tests/xhtml/testfiles/resources/pdf/dummy.pdf";
const THUMBS_UP = "\u{1F44D}";

function env(name: string): string | undefined {
  const value = process.env[name]?.trim();
  return value ? value : undefined;
}

const API_KEY = env("HELLO_API_KEY");
const PHONE_NUMBER_ID = env("HELLO_PHONE_NUMBER_ID");
const TO = env("HELLO_TO");
const TEMPLATE_NAME = env("HELLO_TEMPLATE_NAME");
const TEMPLATE_LANGUAGE = env("HELLO_TEMPLATE_LANGUAGE") ?? "en_US";
const OTP = env("HELLO_OTP") ?? "123456";
/** --all also sends text, media, location, contacts, list and reaction (needs an open 24h window). */
const ALL = process.argv.includes("--all");
const IMAGE_URL = env("HELLO_MEDIA_IMAGE_URL") ?? DEFAULT_IMAGE_URL;
const DOCUMENT_URL = env("HELLO_MEDIA_DOCUMENT_URL") ?? DEFAULT_DOCUMENT_URL;
const VIDEO_URL = env("HELLO_MEDIA_VIDEO_URL");
const AUDIO_URL = env("HELLO_MEDIA_AUDIO_URL");
const INBOUND_ID = env("HELLO_INBOUND_MESSAGE_ID");
const BASE_URL = env("HELLO_BASE_URL");

if (!API_KEY) {
  console.error("HELLO_API_KEY is not set. Copy smoke-test/.env.example to .env, fill it in and load it. Nothing was run.");
  process.exit(1);
}
if (!PHONE_NUMBER_ID) {
  console.error("HELLO_PHONE_NUMBER_ID is not set (see smoke-test/.env.example). Nothing was run.");
  process.exit(1);
}
if (SEND && !TEMPLATE_NAME) {
  console.error("HELLO_TEMPLATE_NAME is not set. Name an AUTHENTICATION (OTP) template for --send. Nothing was run.");
  process.exit(1);
}
if (SEND && !TO) {
  console.error("HELLO_TO is not set. It is needed for --send. Nothing was run.");
  process.exit(1);
}

const hello = new Hello({ apiKey: API_KEY, baseUrl: BASE_URL });
const dest = TO ?? "<HELLO_TO>";

type Status = "PASS" | "FAIL" | "SKIP" | "DRY";
const tally: Record<Status, number> = { PASS: 0, FAIL: 0, SKIP: 0, DRY: 0 };
const sent: { label: string; helloId?: string; wamid?: string }[] = [];
const state: { waba?: string; template?: Template; textWamid?: string } = {};

function line(status: Status, name: string, detail = ""): void {
  tally[status] += 1;
  console.log(`${status.padEnd(4)} ${name}${detail ? `: ${detail}` : ""}`);
}

function describe(err: unknown): string {
  if (err instanceof APIError) return `HTTP ${err.status}${err.helloCode ? ` ${err.helloCode}` : ""}: ${err.message}`;
  return err instanceof Error ? err.message : String(err);
}

async function step(name: string, fn: () => Promise<string>): Promise<void> {
  try {
    line("PASS", name, await fn());
  } catch (err) {
    line("FAIL", name, describe(err));
  }
}

/** A step that sends or writes. Without --send it only reports what it would do. */
async function sendStep(name: string, wouldDo: string, fn: () => Promise<string>): Promise<void> {
  if (!SEND) return line("DRY", name, `would ${wouldDo}`);
  await step(name, fn);
}

/** Extra message types: only with --all (they need the recipient's 24h window to be open). */
async function extraStep(name: string, wouldDo: string, fn: () => Promise<string>): Promise<void> {
  if (!ALL) return line("SKIP", name, "needs --all");
  await sendStep(name, wouldDo, fn);
}

function recipient(): string {
  if (!TO) throw new Error("HELLO_TO is not set");
  return TO;
}

function record(label: string, res: { messages: { id: string; hello_id?: string }[] }): string {
  const m = res.messages[0];
  sent.push({ label, helloId: m?.hello_id, wamid: m?.id || undefined });
  return `hello_id ${m?.hello_id ?? "?"}`;
}

/**
 * Components for an AUTHENTICATION (OTP) template: the code fills the body {{1}} and the
 * copy-code / one-tap button. That is the only template shape this smoke test sends, so it
 * works the same on every account: name the template in HELLO_TEMPLATE_NAME.
 */
function otpComponents(t: Template): TemplateComponent[] {
  if (t.category !== "AUTHENTICATION") {
    throw new Error(`${t.name} is ${t.category}; HELLO_TEMPLATE_NAME must be an AUTHENTICATION (OTP) template`);
  }
  const components: TemplateComponent[] = [{ type: "body", parameters: [{ type: "text", text: OTP }] }];
  const buttons = t.components?.find((c) => String(c.type).toUpperCase() === "BUTTONS")?.buttons;
  if (Array.isArray(buttons) && buttons.length > 0) {
    components.push({ type: "button", sub_type: "url", index: "0", parameters: [{ type: "text", text: OTP }] });
  }
  return components;
}

async function download(url: string): Promise<{ bytes: Uint8Array; mime: string }> {
  const res = await fetch(url, { headers: { "User-Agent": "hello-smoke-test/1.0" } });
  if (!res.ok) throw new Error(`download ${url} returned HTTP ${res.status}`);
  const mime = (res.headers.get("content-type") ?? "application/octet-stream").split(";")[0]?.trim() ?? "";
  return { bytes: new Uint8Array(await res.arrayBuffer()), mime };
}

const sleep = (ms: number) => new Promise((resolve) => setTimeout(resolve, ms));

async function main(): Promise<void> {
  console.log(
    `Hello smoke test, run ${RUN_ID}: ${SEND ? "LIVE, real messages will be sent" : "dry run, nothing is sent (add --send to send)"}`,
  );

  // 1. ping
  await step("1 ping", async () => {
    const res = await hello.ping();
    if (res.status !== "ok") throw new Error(`unexpected status ${res.status}`);
    return `status ${res.status}`;
  });

  // 2. phone number and WABA
  await step("2 find phone number", async () => {
    const { data } = await hello.phoneNumbers.list();
    const match = data.find((n) => n.phone_number_id === PHONE_NUMBER_ID);
    if (!match) throw new Error(`no number with phone_number_id ${PHONE_NUMBER_ID} among ${data.length} on this key`);
    state.waba = match.waba_id;
    return `waba_id ${match.waba_id}, can_send ${match.can_send}`;
  });

  // 3. templates on the WABA, then find HELLO_TEMPLATE_NAME
  await step("3 list templates and find HELLO_TEMPLATE_NAME", async () => {
    if (!state.waba) throw new Error("no WABA id from step 2");
    const { data } = await hello.templates.list(state.waba, { limit: 100 });
    for (const t of data) console.log(`     ${t.name}  ${t.language}  ${t.status}  ${t.category}`);
    if (!TEMPLATE_NAME) return `${data.length} templates (set HELLO_TEMPLATE_NAME to pick one)`;
    const found = data.find((t) => t.name === TEMPLATE_NAME && t.language === TEMPLATE_LANGUAGE);
    if (!found) throw new Error(`no template ${TEMPLATE_NAME} in ${TEMPLATE_LANGUAGE} on WABA ${state.waba}`);
    state.template = found;
    return `using ${found.name} (${found.language}, ${found.category}, ${found.status})`;
  });

  // 4. get that template by id
  await step("4 get template by id", async () => {
    const t = state.template;
    if (!t) throw new Error("no template from step 3");
    const got = await hello.templates.get(t.id);
    if (got.id !== t.id) throw new Error(`asked for ${t.id}, got ${got.id}`);
    return `${got.name} ${got.language} ${got.status}`;
  });

  // Steps 5-11 send or write. Each one is gated by --send.

  // 5. template send
  await sendStep("5 send OTP template", `send template ${TEMPLATE_NAME ?? "<HELLO_TEMPLATE_NAME>"} with code ${OTP} to ${dest}`, async () => {
    const t = state.template;
    if (!t) throw new Error("no template from step 3");
    const res = await hello.messages.sendTemplate(
      PHONE_NUMBER_ID!,
      recipient(),
      { name: t.name, language: t.language, components: otpComponents(t) },
      { idempotencyKey: `${RUN_ID}-template` },
    );
    return record("template", res);
  });

  // 6. text send, synchronous so the reply carries the wamid needed for the reaction
  await extraStep("6 send text", `send a text message to ${dest}`, async () => {
    const res = await hello.messages.sendText(PHONE_NUMBER_ID!, recipient(), `Hello smoke test ${RUN_ID}: text message.`, {
      sync: true,
      idempotencyKey: `${RUN_ID}-text`,
    });
    state.textWamid = res.messages[0]?.id || undefined;
    return record("text", res);
  });

  // 7. media by link
  await extraStep("7 send image by link", `send the image at ${IMAGE_URL}`, async () =>
    record(
      "image-link",
      await hello.messages.sendMedia(
        PHONE_NUMBER_ID!,
        recipient(),
        { type: "image", link: IMAGE_URL, caption: `Smoke test image ${RUN_ID}` },
        { idempotencyKey: `${RUN_ID}-image-link` },
      ),
    ),
  );
  await extraStep("7 send document by link", `send the document at ${DOCUMENT_URL}`, async () =>
    record(
      "document-link",
      await hello.messages.sendMedia(
        PHONE_NUMBER_ID!,
        recipient(),
        { type: "document", link: DOCUMENT_URL, filename: "smoke-test.pdf", caption: `Smoke test document ${RUN_ID}` },
        { idempotencyKey: `${RUN_ID}-document-link` },
      ),
    ),
  );
  if (VIDEO_URL) {
    await extraStep("7 send video by link", `send the video at ${VIDEO_URL}`, async () =>
      record(
        "video-link",
        await hello.messages.sendMedia(
          PHONE_NUMBER_ID!,
          recipient(),
          { type: "video", link: VIDEO_URL, caption: `Smoke test video ${RUN_ID}` },
          { idempotencyKey: `${RUN_ID}-video-link` },
        ),
      ),
    );
  } else {
    line("SKIP", "7 send video by link", "HELLO_MEDIA_VIDEO_URL not set");
  }
  if (AUDIO_URL) {
    await extraStep("7 send audio by link", `send the audio at ${AUDIO_URL}`, async () =>
      record(
        "audio-link",
        await hello.messages.sendMedia(
          PHONE_NUMBER_ID!,
          recipient(),
          { type: "audio", link: AUDIO_URL },
          { idempotencyKey: `${RUN_ID}-audio-link` },
        ),
      ),
    );
  } else {
    line("SKIP", "7 send audio by link", "HELLO_MEDIA_AUDIO_URL not set");
  }

  // 8. download, upload, send by media id
  await extraStep("8 upload image and send by media id", `download ${IMAGE_URL}, upload it and send it by id`, async () => {
    const { bytes, mime } = await download(IMAGE_URL);
    const uploaded = await hello.media.upload(PHONE_NUMBER_ID!, bytes, mime || "image/png", "smoke-test.png");
    const res = await hello.messages.sendMedia(
      PHONE_NUMBER_ID!,
      recipient(),
      { type: "image", id: uploaded.id, caption: `Smoke test uploaded image ${RUN_ID}` },
      { idempotencyKey: `${RUN_ID}-image-id` },
    );
    return `media id ${uploaded.id}, ${record("image-id", res)}`;
  });

  // 9. other message types through the raw send
  await extraStep("9 send location", `send a location to ${dest}`, async () =>
    record(
      "location",
      await hello.messages.send(
        PHONE_NUMBER_ID!,
        {
          messaging_product: "whatsapp",
          to: recipient(),
          type: "location",
          location: { latitude: 28.6139, longitude: 77.209, name: "Smoke test location", address: "New Delhi, India" },
        },
        { idempotencyKey: `${RUN_ID}-location` },
      ),
    ),
  );
  await extraStep("9 send contacts", `send a contact card to ${dest}`, async () =>
    record(
      "contacts",
      await hello.messages.send(
        PHONE_NUMBER_ID!,
        {
          messaging_product: "whatsapp",
          to: recipient(),
          type: "contacts",
          contacts: [
            {
              name: { formatted_name: "Smoke Test", first_name: "Smoke", last_name: "Test" },
              phones: [{ phone: "+15555550100", type: "CELL", wa_id: "15555550100" }],
            },
          ],
        },
        { idempotencyKey: `${RUN_ID}-contacts` },
      ),
    ),
  );
  await sendStep("6 send interactive reply buttons", `send a static reply-buttons message to ${dest}`, async () =>
    record(
      "buttons",
      await hello.messages.send(
        PHONE_NUMBER_ID!,
        {
          messaging_product: "whatsapp",
          to: recipient(),
          type: "interactive",
          interactive: {
            type: "button",
            body: { text: `Smoke test ${RUN_ID}: pick one` },
            action: {
              buttons: [
                { type: "reply", reply: { id: "yes", title: "Yes" } },
                { type: "reply", reply: { id: "no", title: "No" } },
              ],
            },
          },
        },
        { idempotencyKey: `${RUN_ID}-buttons` },
      ),
    ),
  );
  await extraStep("9 send list", `send an interactive list to ${dest}`, async () =>
    record(
      "list",
      await hello.messages.send(
        PHONE_NUMBER_ID!,
        {
          messaging_product: "whatsapp",
          to: recipient(),
          type: "interactive",
          interactive: {
            type: "list",
            body: { text: `Smoke test ${RUN_ID}: pick a row` },
            action: {
              button: "Choose",
              sections: [
                {
                  title: "Smoke test",
                  rows: [
                    { id: "a", title: "Option A" },
                    { id: "b", title: "Option B" },
                  ],
                },
              ],
            },
          },
        },
        { idempotencyKey: `${RUN_ID}-list` },
      ),
    ),
  );
  if (!ALL) {
    line("SKIP", "9 react to text message", "needs --all");
  } else if (!SEND) {
    line("DRY", "9 react to text message", `would react ${THUMBS_UP} to the step 6 message`);
  } else if (!state.textWamid) {
    line("SKIP", "9 react to text message", "step 6 returned no wamid (send failed or was queued)");
  } else {
    const wamid = state.textWamid;
    await step("9 react to text message", async () =>
      record(
        "reaction",
        await hello.messages.send(
          PHONE_NUMBER_ID!,
          {
            messaging_product: "whatsapp",
            to: recipient(),
            type: "reaction",
            reaction: { message_id: wamid, emoji: THUMBS_UP },
          },
          { idempotencyKey: `${RUN_ID}-reaction` },
        ),
      ),
    );
  }

  // 10. status of each sent message
  if (!SEND) {
    line("DRY", "10 message status", "would read the status of each sent message after 5 s");
  } else if (sent.length === 0) {
    line("SKIP", "10 message status", "nothing was sent");
  } else {
    await sleep(5000);
    for (const m of sent) {
      const helloId = m.helloId;
      if (!helloId) {
        line("FAIL", `10 status ${m.label}`, "no hello_id returned");
        continue;
      }
      await step(`10 status ${m.label}`, async () => {
        const msg = await hello.messages.get(helloId);
        if (msg.status === "failed") throw new Error(`failed: ${msg.error_code ?? ""} ${msg.error_message ?? ""}`.trim());
        return `status ${msg.status}`;
      });
    }
  }

  // 11. mark an inbound message as read
  if (!INBOUND_ID) {
    line("SKIP", "11 mark as read", "HELLO_INBOUND_MESSAGE_ID not set");
  } else if (!SEND) {
    line("DRY", "11 mark as read", `would mark ${INBOUND_ID} as read`);
  } else {
    await step("11 mark as read", async () => {
      const res = await hello.messages.markAsRead(PHONE_NUMBER_ID!, INBOUND_ID);
      if (!res.success) throw new Error("API did not return success");
      return "marked read";
    });
  }

  console.log(
    `\nSummary: ${tally.PASS} passed, ${tally.FAIL} failed, ${tally.SKIP} skipped, ${tally.DRY} dry-run only`,
  );
  if (!SEND) console.log("Dry run: nothing was sent. Add --send to send real, billed messages.");
  process.exitCode = tally.FAIL > 0 ? 1 : 0;
}

await main();
