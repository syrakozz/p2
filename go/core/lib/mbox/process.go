package mbox

import (
	"bufio"
	"bytes"
	"context"
	"encoding/base64"
	"encoding/csv"
	"html"
	"io"
	"log/slog"
	"mime"
	"mime/multipart"
	"net/mail"
	"os"
	"regexp"
	"strings"

	"github.com/microcosm-cc/bluemonday"

	"github.com/korylprince/mbox"
)

// ProcessFunc is called for every part and multipart section.
// return true to continue
type ProcessFunc func(ctx context.Context, logCtx *slog.Logger, namespace string, mediaType string, header mail.Header, body string) bool

// ProcessResponse is the Process API response structure.
type ProcessResponse struct {
	Total     int
	Processed int
}

var (
	spacesRex   = regexp.MustCompile(` +`)
	newlinesRex = regexp.MustCompile(`\n{3,}`)
	bm          = bluemonday.StrictPolicy()
)

// Process mbox file
func Process(ctx context.Context, logCtx *slog.Logger, filename, namespace string, processFunc ProcessFunc, continueFlag bool) (ProcessResponse, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	logCtx = logCtx.With("fid", "mbox.Process")

	seen := map[string]struct{}{}

	if continueFlag {
		func() {
			seenF, err := os.Open(filename + ".csv")
			if err != nil {
				return
			}
			defer seenF.Close()

			r := csv.NewReader(seenF)

			for {
				row, err := r.Read()
				if err == io.EOF {
					break
				}
				if err != nil {
					logCtx.Error("unable to read mbox.csv", "error", err)
					return
				}
				seen[row[0]] = struct{}{}
			}

		}()
	}

	f, err := os.Open(filename)
	if err != nil {
		logCtx.Error("unable to open file", "error", err)
		return ProcessResponse{}, err
	}
	defer f.Close()

	var flog *os.File

	if continueFlag {
		flog, err = os.OpenFile(filename+".csv", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0664)
	} else {
		flog, err = os.Create(filename + ".csv")
	}
	if err != nil {
		logCtx.Error("unable to create mbox log file", "error", err)
		return ProcessResponse{}, err
	}
	defer flog.Close()

	csvLog := csv.NewWriter(flog)
	defer csvLog.Flush()

	if !continueFlag {
		csvLog.Write([]string{"id", "subject"})
	}

	m := mbox.NewScanner(f)
	m.MaxTokenSize = 1024 * 1024

	totalCount := 0
	processedCount := 0
	c := false

	for m.Scan() {
		select {
		case <-ctx.Done():
			logCtx.Info("canceled")
			return ProcessResponse{}, context.Canceled
		default:
		}

		totalCount++
		b := m.Bytes()
		r := bufio.NewReader(bytes.NewReader(b))

		// read mbox separator line
		_, _, err = r.ReadLine()
		if err != nil {
			logCtx.Warn("unable to read separator line", "error", err)
			continue
		}

		msg, err := mail.ReadMessage(r)
		if err != nil {
			logCtx.Warn("unable to read message", "error", err)
			continue
		}

		if msg == nil {
			logCtx.Warn("invalid message")
			continue
		}

		header := msg.Header

		messageID := header.Get("Message-ID")
		messageIDClean := cleanID.Replace(messageID)
		logCtx = logCtx.With("id", messageIDClean)

		if continueFlag {
			if _, ok := seen[messageIDClean]; ok {
				continue
			}
		}

		subject := decodeHeader(header.Get("Subject"))

		mediaType, params, err := mime.ParseMediaType(msg.Header.Get("Content-Type"))
		if err != nil {
			logCtx.Error("unable to parse media type", "error", err)
			continue
		}

		if strings.HasPrefix(mediaType, "multipart/") {
			mr := multipart.NewReader(msg.Body, params["boundary"])

			textPlain := ""
			textHTML := ""

			for {
				p, err := mr.NextPart()
				if err == io.EOF {
					break
				}

				if err != nil {
					logCtx.Error("unable to advance to next part", "error", err)
					return ProcessResponse{}, err
				}

				mediaType, _, err := mime.ParseMediaType(p.Header.Get("Content-Type"))
				if err != nil {
					logCtx.Error("unable to parse multipart media type", "error", err)
					return ProcessResponse{}, err
				}

				if mediaType == "text/plain" {
					textPlain, err = readPart(logCtx, p)
					if err != nil {
						return ProcessResponse{}, err
					}
				} else if mediaType == "text/html" {
					textHTML, err = readPart(logCtx, p)
					if err != nil {
						return ProcessResponse{}, err
					}
					textHTML = cleanString(textHTML)
				}
			}

			if textPlain == "" && textHTML == "" {
				logCtx.Warn("not processed", "subject", subject)
				continue
			}

			if textPlain != "" {
				c = processFunc(ctx, logCtx, namespace, "text/plain", header, textPlain)
			} else if textHTML != "" {
				c = processFunc(ctx, logCtx, namespace, "text/html", header, textHTML)
			}

			csvLog.Write([]string{messageIDClean, subject})
			processedCount++
		} else {
			body, err := readPart(logCtx, msg.Body)
			if err != nil {
				return ProcessResponse{}, err
			}
			body = cleanString(body)

			c = processFunc(ctx, logCtx, namespace, mediaType, header, body)

			csvLog.Write([]string{messageIDClean, subject})
			processedCount++
		}

		if !c {
			cancel()
			return ProcessResponse{}, context.Canceled
		}

		if err := m.Err(); err != nil {
			logCtx.Error("unable to scan mbox", "error", err)
			return ProcessResponse{}, err
		}
	}

	res := ProcessResponse{
		Total:     totalCount,
		Processed: processedCount,
	}

	logCtx.Info("ingested", "total", totalCount, "processed", processedCount)
	return res, nil
}

func readPart(logCtx *slog.Logger, r io.Reader) (string, error) {
	b, err := io.ReadAll(r)
	if err != nil {
		logCtx.Error("unable to read multipart", "error", err)
		return "", err
	}

	b2, err := base64.StdEncoding.DecodeString(string(b))
	if err == nil {
		b = b2
	}

	b = bm.SanitizeBytes(b)
	b = bytes.TrimSpace(b)

	// Replace spaces at the beginning or end of each line with nothing
	lines := bytes.Split(b, []byte{'\n'})
	for i, line := range lines {
		line = spacesRex.ReplaceAll(line, []byte{' '})
		lines[i] = bytes.TrimSpace(line)
	}
	b = bytes.Join(lines, []byte{'\n'})

	b = newlinesRex.ReplaceAll(b, []byte{'\n'})
	return html.UnescapeString(string(b)), nil
}
