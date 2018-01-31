package changelog

import (
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strconv"

	"github.com/octokit/go-octokit/octokit"
)

type EntryType int

const (
	UnknownEntryType EntryType = iota
	FeatureRequest
	Bugfix
)

var ErrNoEntry = errors.New("no entry processed from revision")

type Revision interface {
	Subject() string
	Message() string
}

type Entry struct {
	Number  int
	Type    EntryType
	URL     *url.URL
	Message string
}

// Updater determines how the changelog will be updated based on the new commits.
type Updater interface {
	// NewEntry creates a new entry from the revision. If there was a problem processing the revision
	// then an error is returned. If no entry could be processed from the revision, but there was no
	// error, this method should return ErrNoEntry.
	NewEntry(rev Revision) (*Entry, error)
}

type GitHubUpdater struct {
	client *octokit.Client
}

func NewGitHubUpdater(authMethod octokit.AuthMethod) *GitHubUpdater {
	return &GitHubUpdater{
		client: octokit.NewClient(authMethod),
	}
}

var reSubjectLine = regexp.MustCompile(`^Merge pull request #(\d+) from .*$`)

func (u *GitHubUpdater) NewEntry(rev Revision) (*Entry, error) {
	// Determine from the subject line if this is a merge commit from a pull request.
	m := reSubjectLine.FindStringSubmatch(rev.Subject())
	if m == nil {
		return nil, ErrNoEntry
	}

	// Parse the pull request number from the subject line.
	number, err := strconv.Atoi(m[1])
	if err != nil {
		return nil, fmt.Errorf("could not parse pull request number: %s", err)
	}

	// Determine the entry type using the pull request number.
	typ, err := u.findIssueType(number)
	if err != nil {
		return nil, fmt.Errorf("could not identify issue type: %s", err)
	}

	return &Entry{
		Number: number,
		Type:   typ,
		URL: &url.URL{
			Scheme: "https",
			Host:   "github.com",
			Path:   fmt.Sprintf("/influxdata/influxdb/pull/%d", number),
		},
		Message: rev.Message(),
	}, nil
}

func (u *GitHubUpdater) findIssueType(n int) (EntryType, error) {
	labels, result := u.client.IssueLabels().All(nil, octokit.M{
		"owner":  "influxdata",
		"repo":   "influxdb",
		"number": n,
	})
	if result.Err != nil {
		return UnknownEntryType, result.Err
	}

	for _, l := range labels {
		switch l.Name {
		case "kind/feature request":
			return FeatureRequest, nil
		case "kind/bugfix":
			return Bugfix, nil
		}
	}
	return UnknownEntryType, nil
}
