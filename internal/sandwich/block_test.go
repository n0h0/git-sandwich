package sandwich

import (
	"regexp"
	"testing"
)

var (
	startRe = regexp.MustCompile(`# START`)
	endRe   = regexp.MustCompile(`# END`)
)

func TestParseBlocks_Single(t *testing.T) {
	content := `line 1
# START
line 3
line 4
# END
line 6`

	blocks, err := ParseBlocks(content, startRe, endRe, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(blocks) != 1 {
		t.Fatalf("expected 1 block, got %d", len(blocks))
	}
	if blocks[0].StartLine != 2 || blocks[0].EndLine != 5 {
		t.Errorf("expected block {2,5}, got %+v", blocks[0])
	}
}

func TestParseBlocks_Multiple(t *testing.T) {
	content := `# START
content
# END
between
# START
more content
# END`

	blocks, err := ParseBlocks(content, startRe, endRe, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(blocks) != 2 {
		t.Fatalf("expected 2 blocks, got %d", len(blocks))
	}
	if blocks[0].StartLine != 1 || blocks[0].EndLine != 3 {
		t.Errorf("expected block[0] {1,3}, got %+v", blocks[0])
	}
	if blocks[1].StartLine != 5 || blocks[1].EndLine != 7 {
		t.Errorf("expected block[1] {5,7}, got %+v", blocks[1])
	}
}

func TestParseBlocks_NestedAllowed(t *testing.T) {
	content := `# START
  # START
  inner
  # END
# END`

	blocks, err := ParseBlocks(content, startRe, endRe, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(blocks) != 2 {
		t.Fatalf("expected 2 blocks, got %d", len(blocks))
	}
	// Inner block first (stack-based)
	if blocks[0].StartLine != 2 || blocks[0].EndLine != 4 {
		t.Errorf("expected inner block {2,4}, got %+v", blocks[0])
	}
	if blocks[1].StartLine != 1 || blocks[1].EndLine != 5 {
		t.Errorf("expected outer block {1,5}, got %+v", blocks[1])
	}
}

func TestParseBlocks_NestedNotAllowed(t *testing.T) {
	content := `# START
  # START
  # END
# END`

	_, err := ParseBlocks(content, startRe, endRe, false)
	if err == nil {
		t.Fatal("expected error for nested blocks")
	}
}

func TestParseBlocks_UnmatchedBegin(t *testing.T) {
	content := `# START
no end`

	_, err := ParseBlocks(content, startRe, endRe, false)
	if err == nil {
		t.Fatal("expected error for unmatched BEGIN")
	}
}

func TestParseBlocks_UnmatchedEnd(t *testing.T) {
	content := `no start
# END`

	_, err := ParseBlocks(content, startRe, endRe, false)
	if err == nil {
		t.Fatal("expected error for unmatched END")
	}
}

func TestBlock_ContainsLine(t *testing.T) {
	b := Block{StartLine: 5, EndLine: 10}

	tests := []struct {
		line   int
		expect bool
	}{
		{4, false},  // before block
		{5, false},  // START line itself
		{6, true},   // inside
		{9, true},   // inside
		{10, false}, // END line itself
		{11, false}, // after block
	}
	for _, tt := range tests {
		got := b.ContainsLine(tt.line)
		if got != tt.expect {
			t.Errorf("ContainsLine(%d) = %v, want %v", tt.line, got, tt.expect)
		}
	}
}

func TestBlock_IsBoundary(t *testing.T) {
	b := Block{StartLine: 5, EndLine: 10}

	tests := []struct {
		line   int
		expect bool
	}{
		{4, false},
		{5, true},
		{6, false},
		{10, true},
		{11, false},
	}
	for _, tt := range tests {
		got := b.IsBoundary(tt.line)
		if got != tt.expect {
			t.Errorf("IsBoundary(%d) = %v, want %v", tt.line, got, tt.expect)
		}
	}
}

func TestHasBlocks(t *testing.T) {
	withBlocks := `line 1
# START
line 3
# END`
	if !HasBlocks(withBlocks, startRe, endRe) {
		t.Error("expected HasBlocks=true")
	}

	withoutBlocks := `line 1
line 2
line 3`
	if HasBlocks(withoutBlocks, startRe, endRe) {
		t.Error("expected HasBlocks=false")
	}
}
