package convert

import (
	"fmt"
	"os"

	"disruptive/lib/convert"
)

// TextPath extracts text from a file.
func TextPath(path, output string, meta bool) error {
	s, m, err := convert.TextPath(path)
	if err != nil {
		return err
	}

	if output == "" {
		fmt.Println(s)
	} else {
		f, err := os.Create(output)
		if err != nil {
			return err
		}
		defer f.Close()

		if _, err := f.WriteString(s); err != nil {
			return err
		}
	}

	if output == "" {
		fmt.Println()
	}

	if meta {
		fmt.Println("Metadata")
		for k, v := range m {
			fmt.Printf("    %s: %s\n", k, v)
		}
	}

	return nil
}
