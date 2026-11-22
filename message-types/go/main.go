// Command message-types sends one WhatsApp message type and follows its delivery status.
//
//	HELLO_API_KEY=hello_live_xxx HELLO_PHONE_NUMBER_ID=... HELLO_TO=+91... go run . <type>
//
// "go run . all" lists the types.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"slices"
	"strings"
	"time"

	hello "github.com/lakshyapocketlab/hello-go"
)

// sender sends one message type from phone number ID "from" to "to".
type sender func(ctx context.Context, c *hello.Client, from, to string) (*hello.SendResponse, error)

// need reads a required environment variable or exits with a clear message.
func need(name string) string {
	v := os.Getenv(name)
	if v == "" {
		fmt.Fprintf(os.Stderr, "%s is not set (see .env.example)\n", name)
		os.Exit(1)
	}
	return v
}

func language() string {
	if v := os.Getenv("HELLO_TEMPLATE_LANGUAGE"); v != "" {
		return v
	}
	return "en_US"
}

func text(value string) map[string]any { return map[string]any{"type": "text", "text": value} }

// ---------------------------------------------------------------------------
// Free-form messages: need an open 24-hour window (the recipient wrote to you
// in the last 24 hours). Bodies follow https://lakshyapocketlab.com/docs/messages/
// ---------------------------------------------------------------------------

// sendText sends text, up to 4,096 characters. The SDK helper has no preview_url
// option, so this uses the raw send to turn on the link preview.
func sendText(ctx context.Context, c *hello.Client, from, to string) (*hello.SendResponse, error) {
	return c.Messages.Send(ctx, from, map[string]any{
		"messaging_product": "whatsapp",
		"to":                to,
		"type":              "text",
		"text": map[string]any{
			"body":        "Your table for 4 is confirmed for 8 pm. Menu: https://lakshyapocketlab.com/docs/",
			"preview_url": true, // preview of the first URL in body
		},
	})
}

// sendImage sends a JPEG or PNG (up to 5 MB) by public HTTPS link, with a caption.
func sendImage(ctx context.Context, c *hello.Client, from, to string) (*hello.SendResponse, error) {
	return c.Messages.SendMedia(ctx, from, to, hello.MediaMessage{
		Type: hello.MediaImage, Link: need("HELLO_IMAGE_URL"), Caption: "The Runner 2 in red, back in stock.",
	})
}

// sendVideo sends an MP4 or 3GPP video (up to 16 MB) by public HTTPS link, with a caption.
func sendVideo(ctx context.Context, c *hello.Client, from, to string) (*hello.SendResponse, error) {
	return c.Messages.SendMedia(ctx, from, to, hello.MediaMessage{
		Type: hello.MediaVideo, Link: need("HELLO_VIDEO_URL"), Caption: "Setting up your device in 60 seconds",
	})
}

// sendAudio sends audio (up to 16 MB). WhatsApp does not allow captions on audio.
func sendAudio(ctx context.Context, c *hello.Client, from, to string) (*hello.SendResponse, error) {
	return c.Messages.SendMedia(ctx, from, to, hello.MediaMessage{Type: hello.MediaAudio, Link: need("HELLO_AUDIO_URL")})
}

// sendDocument sends a PDF, Office or text file (up to 100 MB) with the file name the recipient sees.
func sendDocument(ctx context.Context, c *hello.Client, from, to string) (*hello.SendResponse, error) {
	return c.Messages.SendMedia(ctx, from, to, hello.MediaMessage{
		Type: hello.MediaDocument, Link: need("HELLO_DOCUMENT_URL"),
		Filename: "Invoice INV-2291.pdf", Caption: "Your invoice for September",
	})
}

// sendSticker sends a static WebP (up to 100 KB). Stickers have no link option:
// upload the file first (c.Media.Upload) and pass the media ID.
func sendSticker(ctx context.Context, c *hello.Client, from, to string) (*hello.SendResponse, error) {
	return c.Messages.SendMedia(ctx, from, to, hello.MediaMessage{Type: hello.MediaSticker, ID: need("HELLO_STICKER_MEDIA_ID")})
}

// sendLocation shares a map pin. Latitude and longitude are numbers here
// (strings in a template header).
func sendLocation(ctx context.Context, c *hello.Client, from, to string) (*hello.SendResponse, error) {
	return c.Messages.Send(ctx, from, map[string]any{
		"messaging_product": "whatsapp",
		"to":                to,
		"type":              "location",
		"location": map[string]any{
			"latitude":  12.9716,
			"longitude": 77.5946,
			"name":      "Indiranagar store",
			"address":   "100 Feet Rd, Indiranagar, Bengaluru 560038",
		},
	})
}

// sendContacts shares a contact card. "contacts" is an array at the top level
// of the request, not an object.
func sendContacts(ctx context.Context, c *hello.Client, from, to string) (*hello.SendResponse, error) {
	return c.Messages.Send(ctx, from, map[string]any{
		"messaging_product": "whatsapp",
		"to":                to,
		"type":              "contacts",
		"contacts": []any{
			map[string]any{
				"name":   map[string]any{"formatted_name": "Priya Sharma", "first_name": "Priya", "last_name": "Sharma"},
				"phones": []any{map[string]any{"phone": "+919812345678", "type": "WORK"}},
			},
		},
	})
}

// sendButtons sends interactive reply buttons: up to 3, titles up to 20 characters.
// A tap arrives as an inbound message carrying the button's id and title.
func sendButtons(ctx context.Context, c *hello.Client, from, to string) (*hello.SendResponse, error) {
	return c.Messages.Send(ctx, from, map[string]any{
		"messaging_product": "whatsapp",
		"to":                to,
		"type":              "interactive",
		"interactive": map[string]any{
			"type":   "button",
			"header": map[string]any{"type": "text", "text": "Booking"},
			"body":   map[string]any{"text": "Confirm your booking for tomorrow 10:00?"},
			"footer": map[string]any{"text": "Reply anytime"},
			"action": map[string]any{
				"buttons": []any{
					map[string]any{"type": "reply", "reply": map[string]any{"id": "yes", "title": "Yes"}},
					map[string]any{"type": "reply", "reply": map[string]any{"id": "no", "title": "No"}},
				},
			},
		},
	})
}

// sendList sends an interactive list: one button opens up to 10 rows in total,
// grouped in sections.
func sendList(ctx context.Context, c *hello.Client, from, to string) (*hello.SendResponse, error) {
	return c.Messages.Send(ctx, from, map[string]any{
		"messaging_product": "whatsapp",
		"to":                to,
		"type":              "interactive",
		"interactive": map[string]any{
			"type":   "list",
			"header": map[string]any{"type": "text", "text": "Appointments"},
			"body":   map[string]any{"text": "Pick a slot"},
			"footer": map[string]any{"text": "Times are IST"},
			"action": map[string]any{
				"button": "View slots", // up to 20 characters
				"sections": []any{
					map[string]any{
						"title": "Monday",
						"rows": []any{
							map[string]any{"id": "mon-9", "title": "09:00", "description": "Morning"},
							map[string]any{"id": "mon-14", "title": "14:00", "description": "Afternoon"},
						},
					},
					map[string]any{
						"title": "Tuesday",
						"rows":  []any{map[string]any{"id": "tue-11", "title": "11:00"}},
					},
				},
			},
		},
	})
}

// sendCtaURL sends an interactive call-to-action URL: one button that opens a
// link. Unlike a template URL button it needs no pre-approval.
func sendCtaURL(ctx context.Context, c *hello.Client, from, to string) (*hello.SendResponse, error) {
	return c.Messages.Send(ctx, from, map[string]any{
		"messaging_product": "whatsapp",
		"to":                to,
		"type":              "interactive",
		"interactive": map[string]any{
			"type":   "cta_url",
			"header": map[string]any{"type": "text", "text": "Your order"},
			"body":   map[string]any{"text": "Track your order in real time"},
			"footer": map[string]any{"text": "Thanks for shopping"},
			"action": map[string]any{
				"name":       "cta_url",
				"parameters": map[string]any{"display_text": "Track order", "url": "https://example.com/track/10432"},
			},
		},
	})
}

// sendReaction reacts with one emoji to a message the customer sent (its
// wamid). An empty emoji removes the reaction.
func sendReaction(ctx context.Context, c *hello.Client, from, to string) (*hello.SendResponse, error) {
	return c.Messages.Send(ctx, from, map[string]any{
		"messaging_product": "whatsapp",
		"to":                to,
		"type":              "reaction",
		"reaction":          map[string]any{"message_id": need("HELLO_REPLY_TO_MESSAGE_ID"), "emoji": "👍"},
	})
}

// sendReply quotes an earlier message: any message type accepts a top-level
// context.message_id next to type. The Go helpers have no reply option, so this
// uses the raw send.
func sendReply(ctx context.Context, c *hello.Client, from, to string) (*hello.SendResponse, error) {
	return c.Messages.Send(ctx, from, map[string]any{
		"messaging_product": "whatsapp",
		"to":                to,
		"context":           map[string]any{"message_id": need("HELLO_REPLY_TO_MESSAGE_ID")},
		"type":              "text",
		"text":              map[string]any{"body": "Yes, that one is back in stock."},
	})
}

// ---------------------------------------------------------------------------
// Template messages: the only type that can start a conversation (no 24-hour
// window needed). The template must exist on the number's WABA and be APPROVED.
// ---------------------------------------------------------------------------

// sendTemplate sends a simple template. Body variables come from
// HELLO_TEMPLATE_PARAMS, comma separated: positional values ("Asha,#10432") or
// named ones ("first_name=Asha,otp=482913"). The default, hello_world, has none.
func sendTemplate(ctx context.Context, c *hello.Client, from, to string) (*hello.SendResponse, error) {
	name := os.Getenv("HELLO_TEMPLATE_NAME")
	if name == "" {
		name = "hello_world"
	}
	var params []any
	for _, v := range strings.Split(os.Getenv("HELLO_TEMPLATE_PARAMS"), ",") {
		if v == "" {
			continue
		}
		if k, val, ok := strings.Cut(v, "="); ok {
			p := text(val)
			p["parameter_name"] = k
			params = append(params, p)
		} else {
			params = append(params, text(v))
		}
	}
	tpl := hello.TemplateMessage{Name: name, Language: language()}
	if len(params) > 0 {
		tpl.Components = []any{map[string]any{"type": "body", "parameters": params}}
	}
	return c.Messages.SendTemplate(ctx, from, to, tpl)
}

// sendTemplateHeaderFooter sends a template with a text header variable and a
// footer. The footer is fixed in the approved template, so only the header and
// body variables are sent.
func sendTemplateHeaderFooter(ctx context.Context, c *hello.Client, from, to string) (*hello.SendResponse, error) {
	return c.Messages.SendTemplate(ctx, from, to, hello.TemplateMessage{
		Name:     need("HELLO_TEMPLATE_HEADER_FOOTER_NAME"),
		Language: language(),
		Components: []any{
			map[string]any{"type": "header", "parameters": []any{text("Asha")}},
			hello.BodyParameters("Dr. Rao", "3 October", "10:30 am"),
		},
	})
}

// sendTemplateButtons sends a template with a dynamic URL button. "index" is the
// button's position in the template's BUTTONS array, as a string. Quick-reply,
// fixed-URL and call buttons need nothing at send time: send only the body.
func sendTemplateButtons(ctx context.Context, c *hello.Client, from, to string) (*hello.SendResponse, error) {
	return c.Messages.SendTemplate(ctx, from, to, hello.TemplateMessage{
		Name:     need("HELLO_TEMPLATE_BUTTONS_NAME"),
		Language: language(),
		Components: []any{
			hello.BodyParameters("Asha"),
			map[string]any{"type": "button", "sub_type": "url", "index": "0", "parameters": []any{text("ORD-10432")}},
		},
	})
}

// sendTemplateMedia sends a template with a media header (here an image by
// public link) plus body variables. Every send must supply the actual media;
// the footer and buttons come from the template.
func sendTemplateMedia(ctx context.Context, c *hello.Client, from, to string) (*hello.SendResponse, error) {
	return c.Messages.SendTemplate(ctx, from, to, hello.TemplateMessage{
		Name:     need("HELLO_TEMPLATE_MEDIA_NAME"),
		Language: language(),
		Components: []any{
			map[string]any{"type": "header", "parameters": []any{
				map[string]any{"type": "image", "image": map[string]any{"link": need("HELLO_IMAGE_URL")}},
			}},
			hello.BodyParameters("Asha", "Runner 2"),
		},
	})
}

var typeNames = []string{
	"text", "image", "video", "audio", "document", "sticker", "location", "contacts",
	"buttons", "list", "cta-url", "reaction", "reply",
	"template", "template-header-footer", "template-buttons", "template-media",
}

var types = map[string]sender{
	"text":                   sendText,
	"image":                  sendImage,
	"video":                  sendVideo,
	"audio":                  sendAudio,
	"document":               sendDocument,
	"sticker":                sendSticker,
	"location":               sendLocation,
	"contacts":               sendContacts,
	"buttons":                sendButtons,
	"list":                   sendList,
	"cta-url":                sendCtaURL,
	"reaction":               sendReaction,
	"reply":                  sendReply,
	"template":               sendTemplate,
	"template-header-footer": sendTemplateHeaderFooter,
	"template-buttons":       sendTemplateButtons,
	"template-media":         sendTemplateMedia,
}

// ---------------------------------------------------------------------------
// Delivery tracking: print the Hello ID, then poll Messages.Get for ~30 s and
// print every status change (queued, sent, delivered, read or failed).
// ---------------------------------------------------------------------------

func track(ctx context.Context, c *hello.Client, helloID string) error {
	deadline := time.Now().Add(30 * time.Second)
	last := ""
	for time.Now().Before(deadline) {
		msg, err := c.Messages.Get(ctx, helloID)
		if err != nil {
			return err
		}
		if msg.Status != last {
			last = msg.Status
			if msg.Status == "failed" {
				fmt.Printf("status: failed (%s %s)\n", msg.ErrorCode, msg.ErrorMessage)
			} else {
				fmt.Println("status:", msg.Status)
			}
		}
		if slices.Contains([]string{"delivered", "read", "failed"}, msg.Status) {
			return nil
		}
		time.Sleep(2 * time.Second)
	}
	fmt.Printf("still %s after 30 s; check later with Messages.Get(%q)\n", last, helloID)
	return nil
}

func main() {
	kind := ""
	if len(os.Args) > 1 {
		kind = os.Args[1]
	}
	if kind == "all" {
		fmt.Println("message types:\n  " + strings.Join(typeNames, "\n  "))
		return
	}
	if os.Getenv("HELLO_API_KEY") == "" {
		fmt.Fprintln(os.Stderr, "HELLO_API_KEY is not set (see .env.example)")
		os.Exit(1)
	}
	send, ok := types[kind]
	if !ok {
		fmt.Fprintf(os.Stderr, "usage: go run . <type>\ntypes: %s\n", strings.Join(typeNames, ", "))
		os.Exit(1)
	}
	from, to := need("HELLO_PHONE_NUMBER_ID"), need("HELLO_TO")

	client, err := hello.NewClient() // reads HELLO_API_KEY
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()
	res, err := send(ctx, client, from, to)
	if err != nil {
		log.Fatal(err)
	}
	helloID := res.Messages[0].HelloID
	fmt.Println("accepted, Hello ID:", helloID)
	if helloID != "" {
		if err := track(ctx, client, helloID); err != nil {
			log.Fatal(err)
		}
	}
}
