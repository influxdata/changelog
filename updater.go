package changelog

import (
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strconv"

	"github.com/influxdata/changelog/git"
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
	ID() string
	Subject() string
	Message() string
}

type Entry struct {
	// Number is the entry number. This usually corresponds to the pull request number.
	Number int

	// Type is the entry type. It determines which section of the changelog the entry should be placed in.
	Type EntryType

	// URL is the url of the entry.
	URL *url.URL

	// Message is the human-readable label for the entry.
	Message string

	// Version is the version this entry should be added to.
	Version *Version
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

var (
	reSubjectLine = regexp.MustCompile(`^Merge pull request #(\d+) from .*$`)
	reVersion     = regexp.MustCompile(`^v?(\d+(\.\d+)*)[-.]?(rc\d+)?$`)
)

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

	// Retrieve the last tag for this PR so we can select which version this belongs in.
	tag, err := git.LastTag(rev.ID())
	if err != nil {
		return nil, err
	}

	m = reVersion.FindStringSubmatch(tag)
	if m == nil {
		return nil, errors.New("could not find version information")
	}

	ver, err := NewVersion(m[1])
	if err != nil {
		return nil, err
	}

	// If rc is present, then we have a release candidate and should not increment the version number.
	// If we do not, then increment the version.
	if m[3] == "" {
		segments := ver.Segments()
		if len(segments) > 1 {
			segments[1]++
		}
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
		Version: ver,
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
