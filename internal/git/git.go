package git

import (
	"os/exec"
	"strings"
)

// GetDiff runs git diff -U0 base...head -- [paths] and returns the raw diff output.
func GetDiff(baseRef, headRef string, paths []string) ([]byte, error) {
	args := []string{"diff", "-U0", baseRef + "..." + headRef, "--"}
	if len(paths) > 0 {
		args = append(args, paths...)
	}
	cmd := exec.Command("git", args...)
	return cmd.Output()
}

// GetFileContent retrieves the content of a file at a given ref using git show.
// Returns the content, whether the file exists, and any error.
func GetFileContent(ref, path string) (string, bool, error) {
	cmd := exec.Command("git", "show", ref+":"+path)
	out, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			stderr := string(exitErr.Stderr)
			if strings.Contains(stderr, "does not exist") ||
				strings.Contains(stderr, "not exist in") ||
				strings.Contains(stderr, "fatal: path") {
				return "", false, nil
			}
		}
		return "", false, err
	}
	return string(out), true, nil
}
