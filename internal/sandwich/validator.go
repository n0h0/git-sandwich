package sandwich

import (
	"fmt"

	"github.com/n0h0/git-sandwich/internal/diff"
	"github.com/n0h0/git-sandwich/internal/git"
)

// Validate performs the sandwich validation based on the given config.
func Validate(cfg *Config) (*Result, error) {
	diffBytes, err := git.GetDiff(cfg.BaseRef, cfg.HeadRef, cfg.Paths)
	if err != nil {
		return nil, fmt.Errorf("failed to get diff: %w", err)
	}

	if len(diffBytes) == 0 {
		return &Result{Success: true}, nil
	}

	fileDiffs, err := diff.Parse(diffBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse diff: %w", err)
	}

	result := &Result{Success: true}
	for _, fd := range fileDiffs {
		fr := validateFile(cfg, &fd)
		result.Files = append(result.Files, fr)
		if !fr.Success {
			result.Success = false
		}
	}
	return result, nil
}

func validateFile(cfg *Config, fd *diff.FileDiff) FileResult {
	path := fd.NewPath
	if fd.IsDeleted {
		path = fd.OldPath
	}

	fr := FileResult{Path: path, Success: true}

	// New file: skip (only validate block structure in head)
	if fd.IsNew {
		content, exists, err := git.GetFileContent(cfg.HeadRef, fd.NewPath)
		if err != nil || !exists {
			fr.SkipReason = "new file"
			return fr
		}
		if HasBlocks(content, cfg.StartMarkerRegex, cfg.EndMarkerRegex) {
			_, blockErr := ParseBlocks(content, cfg.StartMarkerRegex, cfg.EndMarkerRegex, cfg.AllowNesting)
			if blockErr != nil {
				fr.Success = false
				fr.BlockError = blockErr.Error()
				return fr
			}
		}
		fr.SkipReason = "new file"
		return fr
	}

	// Get base content
	baseContent, baseExists, err := git.GetFileContent(cfg.BaseRef, fd.OldPath)
	if err != nil {
		fr.Success = false
		fr.BlockError = fmt.Sprintf("failed to read base file: %v", err)
		return fr
	}

	// If base file doesn't exist or has no blocks, skip
	if !baseExists || !HasBlocks(baseContent, cfg.StartMarkerRegex, cfg.EndMarkerRegex) {
		fr.SkipReason = "no blocks in base"
		return fr
	}

	// Parse base blocks
	baseBlocks, baseBlockErr := ParseBlocks(baseContent, cfg.StartMarkerRegex, cfg.EndMarkerRegex, cfg.AllowNesting)
	if baseBlockErr != nil {
		fr.Success = false
		fr.BlockError = fmt.Sprintf("base: %v", baseBlockErr)
		return fr
	}

	// Deleted file: check all deleted ranges against base blocks
	if fd.IsDeleted {
		outsideBase := findOutsideLines(fd.OldRanges, baseBlocks)
		if len(outsideBase) > 0 {
			fr.Success = false
			fr.OutsideBase = outsideBase
		}
		return fr
	}

	// Normal file: get head content and parse blocks
	headContent, headExists, err := git.GetFileContent(cfg.HeadRef, fd.NewPath)
	if err != nil || !headExists {
		fr.Success = false
		fr.BlockError = "failed to read head file"
		return fr
	}

	headBlocks, headBlockErr := ParseBlocks(headContent, cfg.StartMarkerRegex, cfg.EndMarkerRegex, cfg.AllowNesting)
	if headBlockErr != nil {
		fr.Success = false
		fr.BlockError = fmt.Sprintf("head: %v", headBlockErr)
		return fr
	}

	// Classify changes
	outsideBase, boundaryBase := classifyLines(fd.OldRanges, baseBlocks)
	outsideHead, boundaryHead := classifyLines(fd.NewRanges, headBlocks)

	fr.OutsideBase = outsideBase
	fr.OutsideHead = outsideHead
	fr.BoundaryChanged = len(boundaryBase) > 0 || len(boundaryHead) > 0

	hasOutside := len(outsideBase) > 0 || len(outsideHead) > 0

	if hasOutside {
		if fr.BoundaryChanged && cfg.AllowBoundaryWithOutside {
			fr.Success = true
		} else {
			fr.Success = false
		}
	}

	return fr
}

// classifyLines categorizes each line in the ranges as outside or boundary.
// Lines that are inside blocks are simply ignored (they're OK).
func classifyLines(ranges []diff.LineRange, blocks []Block) (outside []diff.LineRange, boundary []diff.LineRange) {
	for _, r := range ranges {
		for line := r.Start; line <= r.End; line++ {
			classification := classifyLine(line, blocks)
			switch classification {
			case "outside":
				outside = appendOrExtend(outside, line)
			case "boundary":
				boundary = appendOrExtend(boundary, line)
			}
		}
	}
	return
}

// classifyLine classifies a single line number against the blocks.
func classifyLine(line int, blocks []Block) string {
	for _, b := range blocks {
		if b.ContainsLine(line) {
			return "inside"
		}
		if b.IsBoundary(line) {
			return "boundary"
		}
	}
	return "outside"
}

// findOutsideLines returns ranges of lines that are outside all blocks.
func findOutsideLines(ranges []diff.LineRange, blocks []Block) []diff.LineRange {
	var outside []diff.LineRange
	for _, r := range ranges {
		for line := r.Start; line <= r.End; line++ {
			isInside := false
			for _, b := range blocks {
				if b.ContainsLine(line) || b.IsBoundary(line) {
					isInside = true
					break
				}
			}
			if !isInside {
				outside = appendOrExtend(outside, line)
			}
		}
	}
	return outside
}

// appendOrExtend appends a line to the ranges, extending the last range if contiguous.
func appendOrExtend(ranges []diff.LineRange, line int) []diff.LineRange {
	if len(ranges) > 0 && ranges[len(ranges)-1].End == line-1 {
		ranges[len(ranges)-1].End = line
		return ranges
	}
	return append(ranges, diff.LineRange{Start: line, End: line})
}
