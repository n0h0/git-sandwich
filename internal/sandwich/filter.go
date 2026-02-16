package sandwich

import (
	"strings"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/n0h0/git-sandwich/internal/diff"
)

// filterFiles returns only the files that pass the include/exclude filters.
// If no patterns are specified, all files are returned.
func filterFiles(fileDiffs []diff.FileDiff, includes, excludes []string) []diff.FileDiff {
	if len(includes) == 0 && len(excludes) == 0 {
		return fileDiffs
	}

	var result []diff.FileDiff
	for _, fd := range fileDiffs {
		path := fd.NewPath
		if fd.IsDeleted {
			path = fd.OldPath
		}
		if shouldIncludeFile(path, includes, excludes) {
			result = append(result, fd)
		}
	}
	return result
}

// shouldIncludeFile checks whether a file path passes the include/exclude filters.
// Logic: include first (if specified), then exclude.
func shouldIncludeFile(path string, includes, excludes []string) bool {
	if len(includes) > 0 {
		matched := false
		for _, pattern := range includes {
			if matchesPattern(path, pattern) {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}

	for _, pattern := range excludes {
		if matchesPattern(path, pattern) {
			return false
		}
	}

	return true
}

// matchesPattern checks if a path matches a glob pattern.
// If the pattern looks like a plain directory name (no glob chars, no slash suffix,
// and a directory with that name could exist), it is automatically expanded to "pattern/**".
func matchesPattern(path, pattern string) bool {
	// Try the pattern as-is first
	if matched, _ := doublestar.PathMatch(pattern, path); matched {
		return true
	}

	// Auto-expand directory-like patterns: if the pattern has no glob metacharacters
	// and doesn't contain a path separator after the first segment, try pattern/**
	if !containsGlobMeta(pattern) && !strings.HasSuffix(pattern, "/") {
		expanded := pattern + "/**"
		if matched, _ := doublestar.PathMatch(expanded, path); matched {
			return true
		}
	}

	return false
}

// containsGlobMeta returns true if the string contains glob metacharacters.
func containsGlobMeta(s string) bool {
	return strings.ContainsAny(s, "*?[{")
}

