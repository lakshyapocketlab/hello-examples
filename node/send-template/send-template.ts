// Send an approved template. Usage: npm start -- <to> <template_name> <language> [param...]
import { Hello } from "@lakshyapocketlab/hello";

const [to, name, language, ...params] = process.argv.slice(2);
if (!to || !name || !language) {
  console.error("usage: npm start -- <to> <template_name> <language> [param...]");
  process.exit(1);
}

const hello = new Hello(); // reads HELLO_API_KEY

// Use HELLO_PHONE_NUMBER_ID, or fall back to your first connected number.
let from = process.env.HELLO_PHONE_NUMBER_ID;
if (!from) {
  const { data: numbers } = await hello.phoneNumbers.list();
  from = numbers[0]?.phone_number_id;
}
if (!from) throw new Error("No phone numbers connected to this Hello account");

const res = await hello.messages.sendTemplate(
  from,
  to,
  {
    name,
    language,
    components: params.length
      ? [{ type: "body", parameters: params.map((text) => ({ type: "text", text })) }]
      : undefined,
  },
  // Lets the SDK retry safely: a retry returns the original message instead of sending twice.
  { idempotencyKey: crypto.randomUUID() },
);

const id = res.messages[0]?.hello_id;
console.log("accepted, Hello ID:", id);
if (id) console.log("status:", (await hello.messages.get(id)).status);
