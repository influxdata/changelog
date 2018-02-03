package main

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/github/hub/github"
	"github.com/influxdata/changelog"
	"github.com/influxdata/changelog/git"
	"github.com/octokit/go-octokit/octokit"
	flag "github.com/spf13/pflag"
)

const ChangelogPath = "CHANGELOG.md"

func Fatalf(s string, v ...interface{}) {
	fmt.Printf(fmt.Sprintf("fatal: %s\n", s), v...)
	os.Exit(1)
}

func detectAuthMethod() (octokit.AuthMethod, error) {
	localRepo, err := github.LocalRepo()
	if err != nil {
		return nil, err
	}

	project, err := localRepo.MainProject()
	if err != nil {
		return nil, err
	}

	// Check to see if we just already have an access token.
	// If we do, we don't need to fuss with whether or not we want to ask for one.
	host := github.CurrentConfig().Find(project.Host)
	if host == nil {
		// Attempt to access the repository without an access token.
		// Do this using octokit so we avoid the internals of hub which requires
		// the access token to always exist.
		if ok := func() bool {
			client := octokit.NewClient(nil)
			_, result := client.Repositories().One(nil, octokit.M{
				"owner": project.Owner,
				"repo":  project.Name,
			})
			return result.Err == nil
		}(); ok {
			return nil, nil
		}
		host = &github.Host{Host: project.Host}
	}

	// Attempt to use hub to check if the repository exists. This will automatically
	// ask for a username/password for hub if one does not exist.
	client := github.NewClientWithHost(host)
	if !client.IsRepositoryExist(project) {
		return nil, errors.New("repository does not exist")
	}
	return octokit.TokenAuth{AccessToken: client.Host.AccessToken}, nil
}

func main() {
	all := flag.BoolP("all", "a", false, "Select from the first commit to the selected commit (HEAD as the default)")
	flag.Parse()

	// Change to the root of the git repository.
	dir, err := git.Root()
	if err != nil {
		Fatalf("%s", err)
	}
	os.Chdir(dir)

	// Parse the changelog if it exists or generate an empty one.
	var c *changelog.Changelog
	if _, err := os.Stat(ChangelogPath); err == nil {
		c, err = changelog.ParseFile(ChangelogPath)
		if err != nil {
			Fatalf("Could not read %s: %s", ChangelogPath, err)
		}
	} else if !os.IsNotExist(err) {
		Fatalf("Could not read %s: %s", ChangelogPath, err)
	} else {
		c = changelog.New()
	}

	// Process the arguments and convert them to ranges that go to head if they are
	// not already a range.
	args := flag.Args()
	if *all {
		if len(args) == 0 {
			args = append(args, "HEAD")
		}
	} else {
		if len(args) == 0 {
			// If there are no arguments, use the last tag as the range.
			tag, err := git.LastTag()
			if err != nil {
				Fatalf("Could not find last tag: %s", err)
			}

			if tag != "" {
				args = append(args, git.Range(tag, "HEAD"))
			} else {
				args = append(args, "HEAD")
			}
		} else {
			for i, arg := range args {
				if !strings.Contains(arg, "..") {
					args[i] = git.Range(arg, "HEAD")
				}
			}
		}
	}

	revisions, err := git.Merges(args...)
	if err != nil {
		Fatalf("Could not list revisions: %s", err)
	}

	authMethod, err := detectAuthMethod()
	if err != nil {
		Fatalf("Could not access repository: %s", err)
	}

	updater := changelog.NewGitHubUpdater(authMethod)
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

	if err := c.WriteFile(ChangelogPath); err != nil {
		Fatalf("Could not update %s: %s", ChangelogPath, err)
	}
}
