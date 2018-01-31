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

func RevList(start string, end ...string) ([]string, error) {
	var revRange string
	if len(end) > 0 {
		revRange = fmt.Sprintf("%s..%s", start, end[0])
	} else {
		revRange = start
	}

	cmd := exec.Command("git", "rev-list", "--reverse", revRange)
	r, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("could not pipe output: %s", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	var revs []string
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		rev := strings.TrimSpace(scanner.Text())
		revs = append(revs, rev)
	}

	if err := cmd.Wait(); err != nil {
		return nil, err
	}
	return revs, nil
}

type Revision struct {
	subject, body string
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
		subject: subject,
		body:    body,
	}, nil
}
