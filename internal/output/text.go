package output

import (
	"fmt"
	"io"
	"strings"

	"github.com/n0h0/git-sandwich/internal/diff"
	"github.com/n0h0/git-sandwich/internal/sandwich"
)

// FormatText writes the validation result in human-readable text format.
func FormatText(w io.Writer, result *sandwich.Result) {
	if result.Success && len(result.Files) == 0 {
		fmt.Fprintln(w, "OK")
		return
	}

	allSkipped := true
	for _, f := range result.Files {
		if f.SkipReason == "" {
			allSkipped = false
			break
		}
	}
	if allSkipped && result.Success {
		fmt.Fprintln(w, "OK")
		return
	}

	for _, f := range result.Files {
		if f.SkipReason != "" {
			continue
		}

		if f.Success {
			fmt.Fprintf(w, "OK %s\n", f.Path)
			continue
		}

		fmt.Fprintf(w, "FAIL %s\n", f.Path)

		if f.BlockError != "" {
			fmt.Fprintf(w, "  error: %s\n", f.BlockError)
			continue
		}

		if len(f.OutsideBase) > 0 {
			fmt.Fprintf(w, "  outside(base): %s\n", formatRanges(f.OutsideBase))
		}
		if len(f.OutsideHead) > 0 {
			fmt.Fprintf(w, "  outside(head): %s\n", formatRanges(f.OutsideHead))
		}
		if f.BoundaryChanged {
			fmt.Fprintln(w, "  note: boundary changed")
		}
	}

	if result.Success {
		fmt.Fprintln(w, "OK")
	}
}

func formatRanges(ranges []diff.LineRange) string {
	var parts []string
	for _, r := range ranges {
		if r.Start == r.End {
			parts = append(parts, fmt.Sprintf("lines %d", r.Start))
		} else {
			parts = append(parts, fmt.Sprintf("lines %d-%d", r.Start, r.End))
		}
	}
	return strings.Join(parts, ", ")
}
