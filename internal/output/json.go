package output

import (
	"encoding/json"
	"io"

	"github.com/n0h0/git-sandwich/internal/sandwich"
)

// FormatJSON writes the validation result in JSON format.
func FormatJSON(w io.Writer, result *sandwich.Result) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(result)
}
