// Package mbox parses an mbox file.
package mbox

import (
	"fmt"
	"mime"
	"regexp"
	"strconv"
	"strings"
)

var (
	cleanID = strings.NewReplacer(
		"<", "",
		">", "",
		" ", "",
	)

	cleanEqualChars = strings.NewReplacer(
		"=09", "",
		"=20", "",
		"=E2", "",
		"=80", "",
		"=94", "",
	)

	multipleNewlinesRex          = regexp.MustCompile(`(\s*\n){3,}`)
	endingNewlineContinuationRex = regexp.MustCompile(`=\n`)
)

func decodeHeader(subject string) string {
	if !strings.HasPrefix(strings.ToLower(subject), "=?utf-8?") {
		return subject
	}

	dec := &mime.WordDecoder{}
	decoded, err := dec.DecodeHeader(subject)
	if err != nil {
		return subject
	}

	unquoted, err := strconv.Unquote(fmt.Sprintf("%q", decoded))
	if err != nil {
		return decoded
	}

	return unquoted
}

func cleanString(dirty string) string {
	s := cleanEqualChars.Replace(dirty)
	s = endingNewlineContinuationRex.ReplaceAllString(s, "")
	s = strings.ReplaceAll(s, "&nbsp;", " ")
	s = multipleNewlinesRex.ReplaceAllString(s, "\n\n")
	return s
}
