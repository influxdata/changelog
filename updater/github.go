package updater

import (
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strconv"

	"github.com/influxdata/changelog"
	"github.com/influxdata/changelog/git"
	"github.com/octokit/go-octokit/octokit"
)

type githubUpdater struct {
	owner, name string
	client      *octokit.Client

	// Implementation methods that can be mocked in unit tests.
	pullRequest func(u *githubUpdater, n int) (*octokit.PullRequest, error)
	labels      func(u *githubUpdater, n int) ([]octokit.Label, error)
	lastTag     func(rev string) (string, error)
}

func GitHub(owner, name string, authMethod octokit.AuthMethod) changelog.Updater {
	return &githubUpdater{
		owner:  owner,
		name:   name,
		client: octokit.NewClient(authMethod),

		pullRequest: pullRequest,
		labels:      labels,
		lastTag:     lastTag,
	}
}

var (
	reSubjectLine = regexp.MustCompile(`^Merge pull request #(\d+) from .*$`)
	reVersion     = regexp.MustCompile(`^v?(\d+(\.\d+)*)[-.]?(rc\d+)?$`)
)

func (u *githubUpdater) NewEntry(rev changelog.Revision) (*changelog.Entry, error) {
	// Determine from the subject line if this is a merge commit from a pull request.
	m := reSubjectLine.FindStringSubmatch(rev.Subject())
	if m == nil {
		return nil, changelog.ErrNoEntry
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

	// Get the target version for this revision.
	ver, err := u.getTargetVersion(number, rev)
	if err != nil {
		return nil, err
	}

	return &changelog.Entry{
		Number: number,
		Type:   typ,
		URL: &url.URL{
			Scheme: "https",
			Host:   "github.com",
			Path:   fmt.Sprintf("/%s/%s/pull/%d", u.owner, u.name, number),
		},
		Message: rev.Message(),
		Version: ver,
	}, nil
}

func (u *githubUpdater) findIssueType(n int) (changelog.EntryType, error) {
	labels, err := u.labels(u, n)
	if err != nil {
		return changelog.UnknownEntryType, err
	}

	for _, l := range labels {
		switch l.Name {
		case "kind/feature request":
			return changelog.FeatureRequest, nil
		case "kind/bug":
			return changelog.Bugfix, nil
		}
	}
	return changelog.UnknownEntryType, nil
}

func (u *githubUpdater) getTargetVersion(n int, rev changelog.Revision) (*changelog.Version, error) {
	// Retrieve the last tag for this PR so we can select which version this belongs in.
	tag, err := u.lastTag(rev.ID())
	if err != nil {
		return nil, err
	} else if tag == "" {
		return changelog.MustVersion("1.0.0"), nil
	}

	m := reVersion.FindStringSubmatch(tag)
	if m == nil {
		return nil, errors.New("could not find version information")
	}

	ver, err := changelog.NewVersion(m[1])
	if err != nil {
		return nil, err
	}

	// If rc is present, then we have a release candidate and should not increment the version number.
	// If we do not, then increment the version.
	if m[3] == "" {
		// This is not a release candidate. We need to determine if this is a pull request merged into
		// master or if this is merged into a release branch. If it is merged into a release branch,
		// it is a patch change.
		branch, err := u.findTargetBranch(n)
		if err != nil {
			return nil, err
		}

		if branch != "master" {
			v, err := changelog.NewVersion(branch)
			if err != nil {
				return nil, err
			}

			// There is a version. Ensure that the version we are merging into
			// has fewer segments than the last tag version AND that the prefixes
			// match.
			if len(v.Segments()) >= len(ver.Segments()) {
				return nil, errors.New("release branch has equal to or more segments than the version tag")
			} else if !ver.HasPrefix(v) {
				return nil, fmt.Errorf("release branch and tag prefixes do not match: %s != %s", v, ver.Slice(len(v.Segments())))
			}

			// Increment the segment directly after the one for the release branch.
			segments := ver.Segments()
			segments[len(v.Segments())]++
		} else {
			// Increment the minor patch version.
			segments := ver.Segments()
			if len(segments) > 1 {
				segments[1]++
			}
		}
	}
	return ver, nil
}

func (u *githubUpdater) findTargetBranch(n int) (string, error) {
	pr, err := u.pullRequest(u, n)
	if err != nil {
		return "", err
	}
	return pr.Base.Ref, nil
}

func pullRequest(u *githubUpdater, n int) (*octokit.PullRequest, error) {
	reqURL, err := octokit.PullRequestsURL.Expand(octokit.M{
		"owner":  u.owner,
		"repo":   u.name,
		"number": n,
	})
	if err != nil {
		return nil, err
	}
	pr, result := u.client.PullRequests(reqURL).One()
	return pr, result.Err
}

func labels(u *githubUpdater, n int) ([]octokit.Label, error) {
	labels, result := u.client.IssueLabels().All(nil, octokit.M{
		"owner":  u.owner,
		"repo":   u.name,
		"number": n,
	})
	return labels, result.Err
}

func lastTag(rev string) (string, error) {
	return git.LastTag(rev)
}
