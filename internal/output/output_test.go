package output

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/n0h0/git-sandwich/internal/diff"
	"github.com/n0h0/git-sandwich/internal/sandwich"
)

func TestFormatText_Success(t *testing.T) {
	result := &sandwich.Result{Success: true}
	var buf bytes.Buffer
	FormatText(&buf, result)
	if strings.TrimSpace(buf.String()) != "OK" {
		t.Errorf("expected 'OK', got %q", buf.String())
	}
}

func TestFormatText_Failure(t *testing.T) {
	result := &sandwich.Result{
		Success: false,
		Files: []sandwich.FileResult{
			{
				Path:    "config/application.rb",
				Success: false,
				OutsideBase: []diff.LineRange{
					{Start: 10, End: 12},
				},
				OutsideHead: []diff.LineRange{
					{Start: 25, End: 25},
				},
				BoundaryChanged: true,
			},
		},
	}
	var buf bytes.Buffer
	FormatText(&buf, result)
	output := buf.String()

	if !strings.Contains(output, "FAIL config/application.rb") {
		t.Errorf("expected FAIL line, got %q", output)
	}
	if !strings.Contains(output, "outside(base): lines 10-12") {
		t.Errorf("expected outside(base), got %q", output)
	}
	if !strings.Contains(output, "outside(head): lines 25") {
		t.Errorf("expected outside(head), got %q", output)
	}
	if !strings.Contains(output, "note: boundary changed") {
		t.Errorf("expected boundary note, got %q", output)
	}
}

func TestFormatText_BlockError(t *testing.T) {
	result := &sandwich.Result{
		Success: false,
		Files: []sandwich.FileResult{
			{
				Path:       "broken.rb",
				Success:    false,
				BlockError: "BEGIN without matching END at line 5",
			},
		},
	}
	var buf bytes.Buffer
	FormatText(&buf, result)
	output := buf.String()

	if !strings.Contains(output, "FAIL broken.rb") {
		t.Errorf("expected FAIL line, got %q", output)
	}
	if !strings.Contains(output, "error: BEGIN without matching END") {
		t.Errorf("expected error message, got %q", output)
	}
}

func TestFormatJSON(t *testing.T) {
	result := &sandwich.Result{
		Success: false,
		Files: []sandwich.FileResult{
			{
				Path:    "config/application.rb",
				Success: false,
				OutsideBase: []diff.LineRange{
					{Start: 10, End: 12},
				},
				BoundaryChanged: true,
			},
		},
	}
	var buf bytes.Buffer
	err := FormatJSON(&buf, result)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var parsed sandwich.Result
	if err := json.Unmarshal(buf.Bytes(), &parsed); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}
	if parsed.Success {
		t.Error("expected success=false")
	}
	if len(parsed.Files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(parsed.Files))
	}
	if parsed.Files[0].Path != "config/application.rb" {
		t.Errorf("expected path config/application.rb, got %s", parsed.Files[0].Path)
	}
}
