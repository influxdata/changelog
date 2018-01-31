package main

import (
	"fmt"
	"os"

	"github.com/influxdata/changelog"
	"github.com/influxdata/changelog/git"
)

func Fatalf(s string, v ...interface{}) {
	fmt.Printf(fmt.Sprintf("fatal: %s\n", s), v...)
	os.Exit(1)
}

//var reSubjectLine = regexp.MustCompile(`^Merge pull request #(\d+) from (\w+\/)?(\w+-)?((\d+)-)?.*\/(feature|bugfix)$`)

func main() {
	// Change to the root of the git repository.
	dir, err := git.Root()
	if err != nil {
		Fatalf("%s", err)
	}
	os.Chdir(dir)

	// Parse the changelog.
	c, err := changelog.ParseFile("CHANGELOG.md")
	if err != nil {
		Fatalf("Could not load CHANGELOG.md: %s", err)
	}

	// Find the last modified commit of the changelog.
	rev, err := git.LastModified("CHANGELOG.md")
	if err != nil {
		Fatalf("Could not find revision history for CHANGELOG.md: %s", err)
	}

	var revisions []string
	if rev != "" {
		revisions, err = git.RevList(rev, git.HEAD)
	} else {
		revisions, err = git.RevList(git.HEAD)
	}

	if err != nil {
		Fatalf("Could not list revisions: %s", err)
	}

	updater := &changelog.GitHubUpdater{}
	for _, rev := range revisions {
		if err := func() error {
			rev, err := git.Show(rev)
			if err != nil {
				return err
			}

			entry, err := updater.NewEntry(rev)
			if err != nil {
				if err == changelog.ErrNoEntry {
					return nil
				}
				return err
			}
			return c.AddEntry(entry)
		}(); err != nil {
			fmt.Println("Error:", err)
		}
	}

	if err := c.WriteFile("CHANGELOG.md"); err != nil {
		Fatalf("Could not update CHANGELOG.md: %s", err)
	}
}
