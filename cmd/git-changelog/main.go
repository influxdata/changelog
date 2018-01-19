package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"bufio"

	"bytes"

	"regexp"

	"strconv"

	"github.com/influxdata/changelog/git"
)

func Fatalf(s string, v ...interface{}) {
	fmt.Printf(fmt.Sprintf("fatal: %s\n", s), v...)
	os.Exit(1)
}

//var reSubjectLine = regexp.MustCompile(`^Merge pull request #(\d+) from (\w+\/)?(\w+-)?((\d+)-)?.*\/(feature|bugfix)$`)
var reSubjectLine = regexp.MustCompile(`^Merge pull request #(\d+) from (\w+\/)?(\w+-)?((\d+)-)?.*$`)

func main() {
	// Change to the root of the git repository.
	dir, err := git.Root()
	if err != nil {
		Fatalf("%s", err)
	}
	os.Chdir(dir)

	// Find the last modified commit of the changelog.
	rev, err := git.LastModified("CHANGELOG.md")
	if err != nil {
		Fatalf("Could not find revision history for CHANGELOG.md: %s", err)
	}

	var revRange string
	if rev != "" {
		revRange = fmt.Sprintf("%s..HEAD", rev)
	} else {
		revRange = "HEAD"
	}

	// List the commits we need to read in reverse order. That means we will read
	// the older commits before the more recent ones.
	cmd := exec.Command("git", "rev-list", "--reverse", revRange)
	r, err := cmd.StdoutPipe()
	if err != nil {
		Fatalf("Could not pipe output: %s", err)
	}

	if err := cmd.Start(); err != nil {
		Fatalf("Could not list revisions: %s", err)
	}

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		rev := strings.TrimSpace(scanner.Text())
		if err := func() error {
			cmd := exec.Command("git", "show", "-q", "--format=format:%s", rev)
			out, err := cmd.Output()
			if err != nil {
				return err
			}

			subject := string(bytes.TrimSpace(out))
			m := reSubjectLine.FindStringSubmatch(subject)
			if m == nil {
				return nil
			}

			pr, err := strconv.Atoi(m[1])
			if err != nil {
				return err
			}

			var issue int
			if s := m[5]; s != "" {
				if i, err := strconv.Atoi(s); err != nil {
					return err
				} else {
					issue = i
				}
			}
			//typ := m[6]

			cmd = exec.Command("git", "show", "-q", "--format=format:%b", rev)
			title, err := cmd.Output()
			if err != nil {
				return err
			}

			if issue != 0 {
				fmt.Printf("- [#%d](https://github.com/influxdata/influxdb/issues/%d): %s\n", issue, issue, string(bytes.TrimSpace(title)))
			} else {
				fmt.Printf("- [#%d](https://github.com/influxdata/influxdb/pull/%d): %s\n", pr, pr, string(bytes.TrimSpace(title)))
			}
			return nil
		}(); err != nil {
			fmt.Println("Error:", err)
		}
	}

	if err := cmd.Wait(); err != nil {
		os.Exit(1)
	}
}
