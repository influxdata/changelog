package git

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

const HEAD = "HEAD"

// Root returns the root of the git repository.
func Root() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	out, err := cmd.CombinedOutput()
	if err != nil {
		if _, ok := err.(*exec.ExitError); ok {
			return "", errors.New(strings.TrimPrefix(string(bytes.TrimSpace(out)), "fatal: "))
		}
		return "", err
	}
	return string(bytes.TrimSpace(out)), nil
}

// Range returns a string for specifying a range between two commits.
func Range(start, end string) string {
	return fmt.Sprintf("%s..%s", start, end)
}

// Merges returns a list of all revisions where a merge occurred.
func Merges(revs ...string) ([]string, error) {
	args := []string{"rev-list", "--reverse", "--min-parents=2"}
	if len(revs) > 0 {
		args = append(args, revs...)
	} else {
		args = append(args, "HEAD")
	}
	cmd := exec.Command("git", args...)
	r, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("could not pipe output: %s", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	var revisions []string
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		rev := strings.TrimSpace(scanner.Text())
		revisions = append(revisions, rev)
	}

	if err := cmd.Wait(); err != nil {
		return nil, err
	}
	return revisions, nil
}

// LastTag finds the last tag if it exists. If no tag can be found, then
// this returns a blank string.
func LastTag(revs ...string) (string, error) {
	args := []string{"describe", "--abbrev=0", "--tags", "--first-parent"}
	if len(revs) > 0 {
		args = append(args, revs...)
	}
	cmd := exec.Command("git", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		if _, ok := err.(*exec.ExitError); ok {
			errOut := string(bytes.TrimSpace(out))
			if errOut == "fatal: No names found, cannot describe anything." {
				return "", nil
			} else if strings.HasPrefix(errOut, "fatal: No tags can describe") {
				return "", nil
			}
		}
		return "", err
	}
	return string(bytes.TrimSpace(out)), nil
}

type Revision struct {
	id            string
	subject, body string
}

func (r *Revision) ID() string {
	return r.id
}

func (r *Revision) Subject() string {
	return r.subject
}

func (r *Revision) Message() string {
	return r.body
}

func Show(rev string) (*Revision, error) {
	cmd := exec.Command("git", "show", "-q", "--format=format:%s", rev)
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	subject := string(bytes.TrimSpace(out))

	cmd = exec.Command("git", "show", "-q", "--format=format:%b", rev)
	out, err = cmd.Output()
	if err != nil {
		return nil, err
	}
	body := string(bytes.TrimSpace(out))
	return &Revision{
		id:      rev,
		subject: subject,
		body:    body,
	}, nil
}
