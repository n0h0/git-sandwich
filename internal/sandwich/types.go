package sandwich

import (
	"regexp"

	"github.com/n0h0/git-sandwich/internal/diff"
)

// Config holds the configuration for sandwich validation.
type Config struct {
	StartMarkerRegex         *regexp.Regexp
	EndMarkerRegex           *regexp.Regexp
	BaseRef                  string
	HeadRef                  string
	AllowNesting             bool
	AllowBoundaryWithOutside bool
	Paths                    []string
	IncludePatterns          []string
	ExcludePatterns          []string
}

// Block represents a BEGIN/END sandwich block.
// StartLine is the line number of the BEGIN marker.
// EndLine is the line number of the END marker.
type Block struct {
	StartLine int
	EndLine   int
}

// ContainsLine returns true if the given line is inside the block
// (between START and END, exclusive of the marker lines themselves).
func (b Block) ContainsLine(line int) bool {
	return line > b.StartLine && line < b.EndLine
}

// IsBoundary returns true if the given line is a marker line (START or END).
func (b Block) IsBoundary(line int) bool {
	return line == b.StartLine || line == b.EndLine
}

// FileResult represents the validation result for a single file.
type FileResult struct {
	Path            string           `json:"path"`
	Success         bool             `json:"success"`
	OutsideBase     []diff.LineRange `json:"outside_base,omitempty"`
	OutsideHead     []diff.LineRange `json:"outside_head,omitempty"`
	BoundaryChanged bool             `json:"boundary_changed,omitempty"`
	BlockError      string           `json:"block_error,omitempty"`
	SkipReason      string           `json:"skip_reason,omitempty"`
}

// Result represents the overall validation result.
type Result struct {
	Success bool         `json:"success"`
	Files   []FileResult `json:"files"`
}
