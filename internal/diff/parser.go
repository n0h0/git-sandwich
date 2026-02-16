package diff

import (
	godiff "github.com/sourcegraph/go-diff/diff"
)

// LineRange represents a range of lines (1-indexed, inclusive).
type LineRange struct {
	Start int
	End   int
}

// FileDiff represents the diff information for a single file.
type FileDiff struct {
	OldPath   string
	NewPath   string
	IsNew     bool
	IsDeleted bool
	OldRanges []LineRange
	NewRanges []LineRange
}

// Parse parses a unified diff (produced with -U0) and returns per-file diff info.
func Parse(diffBytes []byte) ([]FileDiff, error) {
	fileDiffs, err := godiff.ParseMultiFileDiff(diffBytes)
	if err != nil {
		return nil, err
	}

	var result []FileDiff
	for _, fd := range fileDiffs {
		fileDiff := FileDiff{
			OldPath:   cleanPath(fd.OrigName),
			NewPath:   cleanPath(fd.NewName),
			IsNew:     fd.OrigName == "/dev/null",
			IsDeleted: fd.NewName == "/dev/null",
		}

		for _, hunk := range fd.Hunks {
			// Old side (deletions)
			if hunk.OrigLines > 0 {
				fileDiff.OldRanges = append(fileDiff.OldRanges, LineRange{
					Start: int(hunk.OrigStartLine),
					End:   int(hunk.OrigStartLine + hunk.OrigLines - 1),
				})
			}
			// New side (additions)
			if hunk.NewLines > 0 {
				fileDiff.NewRanges = append(fileDiff.NewRanges, LineRange{
					Start: int(hunk.NewStartLine),
					End:   int(hunk.NewStartLine + hunk.NewLines - 1),
				})
			}
		}

		result = append(result, fileDiff)
	}
	return result, nil
}

// cleanPath removes the a/ or b/ prefix from diff paths.
func cleanPath(path string) string {
	if len(path) > 2 && (path[:2] == "a/" || path[:2] == "b/") {
		return path[2:]
	}
	return path
}
