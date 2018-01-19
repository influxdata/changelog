package git

import (
	"bytes"
	"errors"
	"os/exec"
	"strings"
)

// Root returns the root of the git repository.
func Root() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	out, err := cmd.Output()
	if err != nil {
		if _, ok := err.(*exec.ExitError); ok {
			return "", errors.New(strings.TrimPrefix(string(bytes.TrimSpace(out)), "fatal: "))
		}
		return "", err
	}
	return string(bytes.TrimSpace(out)), nil
}

// LastModified finds the commit hash where the path was last modified.
// If the path does not exist or has never been modified, this returns an empty string.
func LastModified(path string) (string, error) {
	cmd := exec.Command("git", "log", "-1", "-m", "--first-parent", "--pretty=format:%H", "--", path)
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(bytes.TrimSpace(out)), nil
}
