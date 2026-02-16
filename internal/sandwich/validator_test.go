package sandwich

import (
	"testing"

	"github.com/n0h0/git-sandwich/internal/diff"
)

func TestClassifyLine_Inside(t *testing.T) {
	blocks := []Block{{StartLine: 5, EndLine: 10}}
	if got := classifyLine(7, blocks); got != "inside" {
		t.Errorf("expected inside, got %s", got)
	}
}

func TestClassifyLine_Boundary(t *testing.T) {
	blocks := []Block{{StartLine: 5, EndLine: 10}}
	if got := classifyLine(5, blocks); got != "boundary" {
		t.Errorf("expected boundary for START, got %s", got)
	}
	if got := classifyLine(10, blocks); got != "boundary" {
		t.Errorf("expected boundary for END, got %s", got)
	}
}

func TestClassifyLine_Outside(t *testing.T) {
	blocks := []Block{{StartLine: 5, EndLine: 10}}
	if got := classifyLine(3, blocks); got != "outside" {
		t.Errorf("expected outside, got %s", got)
	}
	if got := classifyLine(12, blocks); got != "outside" {
		t.Errorf("expected outside, got %s", got)
	}
}

func TestClassifyLines(t *testing.T) {
	blocks := []Block{{StartLine: 5, EndLine: 10}}

	ranges := []diff.LineRange{{Start: 3, End: 7}}
	outside, boundary := classifyLines(ranges, blocks)

	if len(outside) != 1 || outside[0].Start != 3 || outside[0].End != 4 {
		t.Errorf("expected outside [{3,4}], got %+v", outside)
	}
	if len(boundary) != 1 || boundary[0].Start != 5 || boundary[0].End != 5 {
		t.Errorf("expected boundary [{5,5}], got %+v", boundary)
	}
}

func TestFindOutsideLines(t *testing.T) {
	blocks := []Block{{StartLine: 5, EndLine: 10}}

	ranges := []diff.LineRange{{Start: 6, End: 9}}
	outside := findOutsideLines(ranges, blocks)
	if len(outside) != 0 {
		t.Errorf("expected no outside lines, got %+v", outside)
	}

	ranges = []diff.LineRange{{Start: 3, End: 4}, {Start: 11, End: 12}}
	outside = findOutsideLines(ranges, blocks)
	if len(outside) != 2 {
		t.Fatalf("expected 2 outside ranges, got %d", len(outside))
	}
	if outside[0].Start != 3 || outside[0].End != 4 {
		t.Errorf("expected outside[0] {3,4}, got %+v", outside[0])
	}
	if outside[1].Start != 11 || outside[1].End != 12 {
		t.Errorf("expected outside[1] {11,12}, got %+v", outside[1])
	}
}

func TestAppendOrExtend_Contiguous(t *testing.T) {
	ranges := []diff.LineRange{{Start: 3, End: 4}}
	ranges = appendOrExtend(ranges, 5)
	if len(ranges) != 1 || ranges[0].End != 5 {
		t.Errorf("expected [{3,5}], got %+v", ranges)
	}
}

func TestAppendOrExtend_NonContiguous(t *testing.T) {
	ranges := []diff.LineRange{{Start: 3, End: 4}}
	ranges = appendOrExtend(ranges, 7)
	if len(ranges) != 2 {
		t.Fatalf("expected 2 ranges, got %d", len(ranges))
	}
	if ranges[1].Start != 7 || ranges[1].End != 7 {
		t.Errorf("expected [{3,4},{7,7}], got %+v", ranges)
	}
}

func TestClassifyLines_MultipleBlocks(t *testing.T) {
	blocks := []Block{
		{StartLine: 3, EndLine: 6},
		{StartLine: 10, EndLine: 15},
	}

	ranges := []diff.LineRange{{Start: 7, End: 9}}
	outside, boundary := classifyLines(ranges, blocks)
	if len(outside) != 1 || outside[0].Start != 7 || outside[0].End != 9 {
		t.Errorf("expected outside [{7,9}], got %+v", outside)
	}
	if len(boundary) != 0 {
		t.Errorf("expected no boundary, got %+v", boundary)
	}

	ranges = []diff.LineRange{{Start: 11, End: 14}}
	outside, _ = classifyLines(ranges, blocks)
	if len(outside) != 0 {
		t.Errorf("expected no outside for inside lines, got %+v", outside)
	}
}
