// Command smoke-test is a live smoke test for github.com/lakshyapocketlab/hello-go
// against the real Hello API.
//
// Steps 1-4 are read-only. Steps 5-11 send real, billed WhatsApp messages and run
// only with -send (or --send). Every step prints PASS, FAIL, SKIP or DRY; the run
// continues after a failure. See ../README.md.
//
// Usage:
//
//	go run .          # dry run
//	go run . -send    # live
package main

import (
	"bytes"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	hello "github.com/lakshyapocketlab/hello-go"
)

const (
	defaultImageURL    = "https://upload.wikimedia.org/wikipedia/commons/4/47/PNG_transparency_demonstration_1.png"
	defaultDocumentURL = "https://www.w3.org/WAI/ER/tests/xhtml/testfiles/resources/pdf/dummy.pdf"
	thumbsUp           = "\U0001F44D"
	userAgent          = "hello-smoke-test/1.0"
)

var placeholderRe = regexp.MustCompile(`\{\{(\d+)\}\}`)

func env(name string) string { return strings.TrimSpace(os.Getenv(name)) }

func envOr(name, fallback string) string {
	if v := env(name); v != "" {
		return v
	}
	return fallback
}

// tally counts step outcomes for the summary.
var tally = map[string]int{"PASS": 0, "FAIL": 0, "SKIP": 0, "DRY": 0}

func line(status, name, detail string) {
	tally[status]++
	if detail != "" {
		fmt.Printf("%-4s %s: %s\n", status, name, detail)
		return
	}
	fmt.Printf("%-4s %s\n", status, name)
}

func describe(err error) string {
	var apiErr *hello.APIError
	if errors.As(err, &apiErr) {
		code := ""
		if apiErr.HelloCode != "" {
			code = " " + apiErr.HelloCode
		}
		return fmt.Sprintf("HTTP %d%s: %s", apiErr.Status, code, apiErr.Message)
	}
	return err.Error()
}

// sentMessage is a successful send; step 10 reads its status by helloID.
type sentMessage struct {
	label   string
	helloID string
}

// templateVars returns the highest {{n}} placeholder count in the BODY component.
func templateVars(t hello.Template) int {
	highest := 0
	for _, c := range t.Components {
		m, ok := c.(map[string]any)
		if !ok || m["type"] != "BODY" {
			continue
		}
		text, _ := m["text"].(string)
		for _, match := range placeholderRe.FindAllStringSubmatch(text, -1) {
			if n, err := strconv.Atoi(match[1]); err == nil && n > highest {
				highest = n
			}
		}
	}
	return highest
}

// templateUsable is true when the only placeholders are body variables.
func templateUsable(t hello.Template) bool {
	for _, c := range t.Components {
		m, ok := c.(map[string]any)
		if ok && m["type"] == "BODY" {
			continue
		}
		if strings.Contains(fmt.Sprint(c), "{{") {
			return false
		}
	}
	return true
}

func pickTemplate(templates []hello.Template, wanted, wantedLang string) *hello.Template {
	if wanted != "" {
		var named []hello.Template
		for _, t := range templates {
			if t.Name == wanted && (wantedLang == "" || t.Language == wantedLang) && templateUsable(t) {
				named = append(named, t)
			}
		}
		for i := range named {
			if named[i].Status == "APPROVED" {
				return &named[i]
			}
		}
		if len(named) > 0 {
			return &named[0]
		}
		return nil
	}
	var approved []hello.Template
	for _, t := range templates {
		if t.Status == "APPROVED" && templateUsable(t) {
			approved = append(approved, t)
		}
	}
	for i := range approved {
		if templateVars(approved[i]) == 0 {
			return &approved[i]
		}
	}
	if len(approved) > 0 {
		return &approved[0]
	}
	return nil
}

func download(ctx context.Context, url string) ([]byte, string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, "", err
	}
	req.Header.Set("User-Agent", userAgent)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("download %s returned HTTP %d", url, resp.StatusCode)
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", err
	}
	mime := strings.TrimSpace(strings.Split(resp.Header.Get("Content-Type"), ";")[0])
	return data, mime, nil
}

func main() {
	send := flag.Bool("send", false, "send real, billed messages (default: dry run)")
	flag.Parse()

	apiKey := env("HELLO_API_KEY")
	pnid := env("HELLO_PHONE_NUMBER_ID")
	to := env("HELLO_TO")
	switch {
	case apiKey == "":
		fmt.Println("HELLO_API_KEY is not set. Copy smoke-test/.env.example to .env, fill it in and load it. Nothing was run.")
		os.Exit(1)
	case pnid == "":
		fmt.Println("HELLO_PHONE_NUMBER_ID is not set (see smoke-test/.env.example). Nothing was run.")
		os.Exit(1)
	case *send && to == "":
		fmt.Println("HELLO_TO is not set. It is needed for -send. Nothing was run.")
		os.Exit(1)
	}
	if !*send {
		to = envOr("HELLO_TO", "<HELLO_TO>")
	}

	runID := fmt.Sprintf("smoke-%d", time.Now().UnixMilli())
	mode := "dry run, nothing is sent (add -send to send)"
	if *send {
		mode = "LIVE, real messages will be sent"
	}
	fmt.Printf("Hello smoke test, run %s: %s\n", runID, mode)

	opts := []hello.Option{hello.WithAPIKey(apiKey)}
	if base := env("HELLO_BASE_URL"); base != "" {
		opts = append(opts, hello.WithBaseURL(base))
	}
	client, err := hello.NewClient(opts...)
	if err != nil {
		fmt.Println("could not create client:", describe(err))
		os.Exit(1)
	}

	ctx := context.Background()
	templateName := env("HELLO_TEMPLATE_NAME")
	templateLang := env("HELLO_TEMPLATE_LANGUAGE")
	imageURL := envOr("HELLO_MEDIA_IMAGE_URL", defaultImageURL)
	documentURL := envOr("HELLO_MEDIA_DOCUMENT_URL", defaultDocumentURL)
	videoURL := env("HELLO_MEDIA_VIDEO_URL")
	audioURL := env("HELLO_MEDIA_AUDIO_URL")
	inboundID := env("HELLO_INBOUND_MESSAGE_ID")

	// step runs a read-only step and reports its outcome.
	step := func(name string, fn func() (string, error)) {
		detail, err := fn()
		if err != nil {
			line("FAIL", name, describe(err))
			return
		}
		line("PASS", name, detail)
	}
	// record keeps a send's hello_id for step 10 and returns a detail line.
	var sent []sentMessage
	record := func(label string, res *hello.SendResponse) string {
		m := res.Messages[0]
		sent = append(sent, sentMessage{label: label, helloID: m.HelloID})
		return "hello_id " + m.HelloID
	}
	// sendStep runs a send or write step only with -send.
	sendStep := func(name, wouldDo string, fn func() (string, error)) {
		if !*send {
			line("DRY", name, "would "+wouldDo)
			return
		}
		step(name, fn)
	}

	// 1. ping
	step("1 ping", func() (string, error) {
		res, err := client.Ping(ctx)
		if err != nil {
			return "", err
		}
		if res.Status != "ok" {
			return "", fmt.Errorf("unexpected status %q", res.Status)
		}
		return "status " + res.Status, nil
	})

	// 2. phone number and WABA
	var waba string
	step("2 find phone number", func() (string, error) {
		numbers, err := client.PhoneNumbers.List(ctx)
		if err != nil {
			return "", err
		}
		for _, n := range numbers {
			if n.PhoneNumberID == pnid {
				waba = n.WabaID
				return fmt.Sprintf("waba_id %s, can_send %t", n.WabaID, n.CanSend), nil
			}
		}
		return "", fmt.Errorf("no number with phone_number_id %s among %d on this key", pnid, len(numbers))
	})

	// 3. templates on the WABA, then pick one
	var picked *hello.Template
	step("3 list templates and pick one", func() (string, error) {
		if waba == "" {
			return "", errors.New("no WABA id from step 2")
		}
		list, err := client.Templates.List(ctx, waba, &hello.ListTemplatesParams{Limit: 100})
		if err != nil {
			return "", err
		}
		for _, t := range list.Data {
			fmt.Printf("     %s  %s  %s  %s  variables=%d\n", t.Name, t.Language, t.Status, t.Category, templateVars(t))
		}
		picked = pickTemplate(list.Data, templateName, templateLang)
		if picked == nil {
			if templateName != "" {
				return "", fmt.Errorf("no usable template named %s", templateName)
			}
			return "", errors.New("no APPROVED template this smoke test can fill in")
		}
		return fmt.Sprintf("using %s (%s, %s, %d variables)", picked.Name, picked.Language, picked.Status, templateVars(*picked)), nil
	})

	// 4. get template by id
	step("4 get template by id", func() (string, error) {
		if picked == nil {
			return "", errors.New("no template from step 3")
		}
		got, err := client.Templates.Get(ctx, picked.ID)
		if err != nil {
			return "", err
		}
		if got.ID != picked.ID {
			return "", fmt.Errorf("asked for %s, got %s", picked.ID, got.ID)
		}
		return fmt.Sprintf("%s %s %s", got.Name, got.Language, got.Status), nil
	})

	// 5-11 send or write; each one is gated by -send.

	// Sample body parameters for the template's variables.
	templateComponents := func(t hello.Template) []any {
		n := templateVars(t)
		if n == 0 {
			return nil
		}
		values := make([]string, n)
		for i := range values {
			values[i] = fmt.Sprintf("Sample %d", i+1)
		}
		return []any{hello.BodyParameters(values...)}
	}
	templateLabel := "?"
	if picked != nil {
		templateLabel = picked.Name
	}

	// 5. template
	sendStep("5 send template", "send template "+templateLabel+" to "+to, func() (string, error) {
		if picked == nil {
			return "", errors.New("no template from step 3")
		}
		res, err := client.Messages.SendTemplate(ctx, pnid, to, hello.TemplateMessage{
			Name:       picked.Name,
			Language:   picked.Language,
			Components: templateComponents(*picked),
		}, hello.WithIdempotencyKey(runID+"-template"))
		if err != nil {
			return "", err
		}
		return record("template", res), nil
	})

	// 6. text, synchronous so the reply carries the wamid the reaction needs
	var textWamid string
	sendStep("6 send text", "send a text message to "+to, func() (string, error) {
		res, err := client.Messages.SendText(ctx, pnid, to, "Hello smoke test "+runID+": text message.",
			hello.WithSync(), hello.WithIdempotencyKey(runID+"-text"))
		if err != nil {
			return "", err
		}
		textWamid = res.Messages[0].ID
		return record("text", res), nil
	})

	// 7. media by link
	sendStep("7 send image by link", "send the image at "+imageURL, func() (string, error) {
		res, err := client.Messages.SendMedia(ctx, pnid, to, hello.MediaMessage{
			Type: hello.MediaImage, Link: imageURL, Caption: "Smoke test image " + runID,
		}, hello.WithIdempotencyKey(runID+"-image-link"))
		if err != nil {
			return "", err
		}
		return record("image-link", res), nil
	})
	sendStep("7 send document by link", "send the document at "+documentURL, func() (string, error) {
		res, err := client.Messages.SendMedia(ctx, pnid, to, hello.MediaMessage{
			Type: hello.MediaDocument, Link: documentURL, Filename: "smoke-test.pdf", Caption: "Smoke test document " + runID,
		}, hello.WithIdempotencyKey(runID+"-document-link"))
		if err != nil {
			return "", err
		}
		return record("document-link", res), nil
	})
	if videoURL != "" {
		sendStep("7 send video by link", "send the video at "+videoURL, func() (string, error) {
			res, err := client.Messages.SendMedia(ctx, pnid, to, hello.MediaMessage{
				Type: hello.MediaVideo, Link: videoURL, Caption: "Smoke test video " + runID,
			}, hello.WithIdempotencyKey(runID+"-video-link"))
			if err != nil {
				return "", err
			}
			return record("video-link", res), nil
		})
	} else {
		line("SKIP", "7 send video by link", "HELLO_MEDIA_VIDEO_URL not set")
	}
	if audioURL != "" {
		sendStep("7 send audio by link", "send the audio at "+audioURL, func() (string, error) {
			res, err := client.Messages.SendMedia(ctx, pnid, to, hello.MediaMessage{
				Type: hello.MediaAudio, Link: audioURL,
			}, hello.WithIdempotencyKey(runID+"-audio-link"))
			if err != nil {
				return "", err
			}
			return record("audio-link", res), nil
		})
	} else {
		line("SKIP", "7 send audio by link", "HELLO_MEDIA_AUDIO_URL not set")
	}

	// 8. download, upload, send by media id
	sendStep("8 upload image and send by media id", "download "+imageURL+", upload it and send it by id", func() (string, error) {
		data, mime, err := download(ctx, imageURL)
		if err != nil {
			return "", err
		}
		if mime == "" {
			mime = "image/png"
		}
		up, err := client.Media.Upload(ctx, pnid, bytes.NewReader(data), "smoke-test.png", mime)
		if err != nil {
			return "", err
		}
		res, err := client.Messages.SendMedia(ctx, pnid, to, hello.MediaMessage{
			Type: hello.MediaImage, ID: up.ID, Caption: "Smoke test uploaded image " + runID,
		}, hello.WithIdempotencyKey(runID+"-image-id"))
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("media id %s, %s", up.ID, record("image-id", res)), nil
	})

	// 9. other types through the raw send
	rawSend := func(label string, body map[string]any) func() (string, error) {
		return func() (string, error) {
			res, err := client.Messages.Send(ctx, pnid, body, hello.WithIdempotencyKey(runID+"-"+label))
			if err != nil {
				return "", err
			}
			return record(label, res), nil
		}
	}
	sendStep("9 send location", "send a location to "+to, rawSend("location", map[string]any{
		"messaging_product": "whatsapp",
		"to":                to,
		"type":              "location",
		"location": map[string]any{
			"latitude": 28.6139, "longitude": 77.209, "name": "Smoke test location", "address": "New Delhi, India",
		},
	}))
	sendStep("9 send contacts", "send a contact card to "+to, rawSend("contacts", map[string]any{
		"messaging_product": "whatsapp",
		"to":                to,
		"type":              "contacts",
		"contacts": []any{map[string]any{
			"name":   map[string]any{"formatted_name": "Smoke Test", "first_name": "Smoke", "last_name": "Test"},
			"phones": []any{map[string]any{"phone": "+15555550100", "type": "CELL", "wa_id": "15555550100"}},
		}},
	}))
	sendStep("9 send reply buttons", "send reply buttons to "+to, rawSend("buttons", map[string]any{
		"messaging_product": "whatsapp",
		"to":                to,
		"type":              "interactive",
		"interactive": map[string]any{
			"type": "button",
			"body": map[string]any{"text": "Smoke test " + runID + ": pick one"},
			"action": map[string]any{"buttons": []any{
				map[string]any{"type": "reply", "reply": map[string]any{"id": "yes", "title": "Yes"}},
				map[string]any{"type": "reply", "reply": map[string]any{"id": "no", "title": "No"}},
			}},
		},
	}))
	sendStep("9 send list", "send an interactive list to "+to, rawSend("list", map[string]any{
		"messaging_product": "whatsapp",
		"to":                to,
		"type":              "interactive",
		"interactive": map[string]any{
			"type": "list",
			"body": map[string]any{"text": "Smoke test " + runID + ": pick a row"},
			"action": map[string]any{
				"button": "Choose",
				"sections": []any{map[string]any{
					"title": "Smoke test",
					"rows": []any{
						map[string]any{"id": "a", "title": "Option A"},
						map[string]any{"id": "b", "title": "Option B"},
					},
				}},
			},
		},
	}))
	if !*send {
		line("DRY", "9 react to text message", "would react "+thumbsUp+" to the step 6 message")
	} else if textWamid == "" {
		line("SKIP", "9 react to text message", "step 6 returned no wamid (send failed or was queued)")
	} else {
		step("9 react to text message", rawSend("reaction", map[string]any{
			"messaging_product": "whatsapp",
			"to":                to,
			"type":              "reaction",
			"reaction":          map[string]any{"message_id": textWamid, "emoji": thumbsUp},
		}))
	}

	// 10. status of each sent message
	switch {
	case !*send:
		line("DRY", "10 message status", "would read the status of each sent message after 5 s")
	case len(sent) == 0:
		line("SKIP", "10 message status", "nothing was sent")
	default:
		time.Sleep(5 * time.Second)
		for _, m := range sent {
			if m.helloID == "" {
				line("FAIL", "10 status "+m.label, "no hello_id returned")
				continue
			}
			step("10 status "+m.label, func() (string, error) {
				msg, err := client.Messages.Get(ctx, m.helloID)
				if err != nil {
					return "", err
				}
				if msg.Status == "failed" {
					return "", fmt.Errorf("failed: %s %s", msg.ErrorCode, msg.ErrorMessage)
				}
				return "status " + msg.Status, nil
			})
		}
	}

	// 11. mark an inbound message as read
	switch {
	case inboundID == "":
		line("SKIP", "11 mark as read", "HELLO_INBOUND_MESSAGE_ID not set")
	case !*send:
		line("DRY", "11 mark as read", "would mark "+inboundID+" as read")
	default:
		step("11 mark as read", func() (string, error) {
			if err := client.Messages.MarkAsRead(ctx, pnid, inboundID); err != nil {
				return "", err
			}
			return "marked read", nil
		})
	}

	fmt.Printf("\nSummary: %d passed, %d failed, %d skipped, %d dry-run only\n",
		tally["PASS"], tally["FAIL"], tally["SKIP"], tally["DRY"])
	if !*send {
		fmt.Println("Dry run: nothing was sent. Add -send to send real, billed messages.")
	}
	if tally["FAIL"] > 0 {
		os.Exit(1)
	}
}
