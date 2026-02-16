package sandwich

import (
	"fmt"
	"regexp"
	"strings"
)

// ParseBlocks scans the content line by line and returns matched BEGIN/END block pairs.
func ParseBlocks(content string, startRe, endRe *regexp.Regexp, allowNesting bool) ([]Block, error) {
	lines := strings.Split(content, "\n")
	var blocks []Block
	var stack []int // stack of BEGIN line numbers (1-indexed)

	for i, line := range lines {
		lineNum := i + 1
		isStart := startRe.MatchString(line)
		isEnd := endRe.MatchString(line)

		if isStart {
			if !allowNesting && len(stack) > 0 {
				return nil, fmt.Errorf("nested BEGIN at line %d (nesting not allowed)", lineNum)
			}
			stack = append(stack, lineNum)
		} else if isEnd {
			if len(stack) == 0 {
				return nil, fmt.Errorf("END without matching BEGIN at line %d", lineNum)
			}
			startLine := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			blocks = append(blocks, Block{StartLine: startLine, EndLine: lineNum})
		}
	}

	if len(stack) > 0 {
		return nil, fmt.Errorf("BEGIN without matching END at line %d", stack[len(stack)-1])
	}

	return blocks, nil
}

// HasBlocks returns true if the content contains any BEGIN or END markers.
func HasBlocks(content string, startRe, endRe *regexp.Regexp) bool {
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		if startRe.MatchString(line) || endRe.MatchString(line) {
			return true
		}
	}
	return false
}
