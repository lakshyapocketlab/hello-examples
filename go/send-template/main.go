// Command send-template sends an approved template with body parameters.
//
//	HELLO_API_KEY=hello_live_xxx go run . <phone_number_id> <to> <template> <language> [param...]
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	hello "github.com/lakshyapocketlab/hello-go"
)

func main() {
	if len(os.Args) < 5 {
		log.Fatal("usage: send-template <phone_number_id> <to> <template> <language> [param...]")
	}
	client, err := hello.NewClient() // reads HELLO_API_KEY
	if err != nil {
		log.Fatal(err)
	}
	tpl := hello.TemplateMessage{Name: os.Args[3], Language: os.Args[4]}
	if len(os.Args) > 5 {
		tpl.Components = []any{hello.BodyParameters(os.Args[5:]...)}
	}
	// The idempotency key lets the SDK retry the send safely on 5xx or network errors.
	resp, err := client.Messages.SendTemplate(context.Background(), os.Args[1], os.Args[2], tpl,
		hello.WithIdempotencyKey(fmt.Sprintf("example-%s-%s", os.Args[2], os.Args[3])))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Hello ID:", resp.Messages[0].HelloID)
}
