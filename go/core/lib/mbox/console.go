package mbox

import (
	"context"
	"fmt"
	"log/slog"
	"net/mail"
	"time"
)

// ConsoleFunc prints each mbox part to the console.
// Return true to continue processing the next email.
func ConsoleFunc(ctx context.Context, logCtx *slog.Logger, namespace string, mediaType string, header mail.Header, body string) bool {
	// Not used for console output
	_ = ctx
	_ = logCtx
	_ = namespace

	var fromStr string

	from, _ := header.AddressList("From")
	if len(from) > 0 {
		fromStr = from[0].String()
	}

	to, _ := header.AddressList("To")
	cc, _ := header.AddressList("Cc")
	bcc, _ := header.AddressList("Bcc")

	date, _ := header.Date()
	datetime := date.UTC().Format(time.DateTime)

	subject := decodeHeader(header.Get("Subject"))
	messageID := header.Get("Message-ID")
	threadIndex := header.Get("Thread-Index")
	threadTopic := header.Get("Thread-Topic")

	fmt.Println()
	fmt.Println("#####################################################################")
	fmt.Println("#####################################################################")
	fmt.Println()

	fmt.Println("From:", fromStr)
	fmt.Println("To:", to)
	fmt.Println("CC:", cc)
	fmt.Println("BCC:", bcc)
	fmt.Println("Subject:", subject)
	fmt.Println("DateTime UTC:", datetime)
	fmt.Println("Message-ID", messageID)
	fmt.Println("Thread-Index:", threadIndex)
	fmt.Println("Thread-Topic:", threadTopic)
	fmt.Println("MediaType", mediaType)

	fmt.Println()
	fmt.Println(body)

	return true

}
