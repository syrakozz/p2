// Package convert processing
package convert

import (
	"code.sajari.com/docconv"
)

// TextPath extracts text from a file.
func TextPath(path string) (string, map[string]string, error) {
	res, err := docconv.ConvertPath(path)
	if err != nil {
		return "", nil, err
	}

	return res.Body, res.Meta, err
}
