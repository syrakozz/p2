package common

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// MarshalIndent adds HTML escaping to the JSON encoder.
func MarshalIndent(a any) string {
	buf := &bytes.Buffer{}
	enc := json.NewEncoder(buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	enc.Encode(a)
	return buf.String()
}

// P is a convenient function to print the output of MarshalIndent.
func P(a any) {
	fmt.Println(MarshalIndent(a))
}
