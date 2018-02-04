package updater

import (
	"net/url"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/influxdata/changelog"
	"github.com/octokit/go-octokit/octokit"
)

func TestGitHubUpdater_FeatureRequest(t *testing.T) {
	u := GitHub("influxdata", "changelog", nil).(*githubUpdater)
	u.pullRequest = func(u *githubUpdater, n int) (*octokit.PullRequest, error) {
		if got, want := n, 1; got != want {
			t.Errorf("unexpected pull request number: %d != %d", got, want)
		}
		return &octokit.PullRequest{
			Base: octokit.PullRequestCommit{
				Ref: "master",
			},
		}, nil
	}
	u.labels = func(u *githubUpdater, n int) ([]octokit.Label, error) {
		if got, want := n, 1; got != want {
			t.Errorf("unexpected pull request number: %d != %d", got, want)
		}
		return []octokit.Label{
			{Name: "kind/feature request"},
		}, nil
	}
	u.lastTag = func(rev string) (string, error) {
		if got, want := rev, "abc"; got != want {
			t.Errorf("unexpected revision: %s != %s", got, want)
		}
		return "v1.3.0", nil
	}

	rev := NewRevision("abc",
		"Merge pull request #1 from influxdata/add-new-feature",
		"Add a new feature",
	)

	got, err := u.NewEntry(rev)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	want := &changelog.Entry{
		Number: 1,
		Type:   changelog.FeatureRequest,
		URL: &url.URL{
			Scheme: "https",
			Host:   "github.com",
			Path:   "/influxdata/changelog/pull/1",
		},
		Message: "Add a new feature",
		Version: changelog.MustVersion("1.4.0"),
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Fatalf("unexpected changelog entry:\n%s", diff)
	}
}

func TestGitHubUpdater_Bugfix(t *testing.T) {
	u := GitHub("influxdata", "changelog", nil).(*githubUpdater)
	u.pullRequest = func(u *githubUpdater, n int) (*octokit.PullRequest, error) {
		if got, want := n, 1; got != want {
			t.Errorf("unexpected pull request number: %d != %d", got, want)
		}
		return &octokit.PullRequest{
			Base: octokit.PullRequestCommit{
				Ref: "master",
			},
		}, nil
	}
	u.labels = func(u *githubUpdater, n int) ([]octokit.Label, error) {
		if got, want := n, 1; got != want {
			t.Errorf("unexpected pull request number: %d != %d", got, want)
		}
		return []octokit.Label{
			{Name: "kind/bug"},
		}, nil
	}
	u.lastTag = func(rev string) (string, error) {
		if got, want := rev, "abc"; got != want {
			t.Errorf("unexpected revision: %s != %s", got, want)
		}
		return "v1.3.0", nil
	}

	rev := NewRevision("abc",
		"Merge pull request #1 from influxdata/fix-a-bug",
		"Fix a bug",
	)

	got, err := u.NewEntry(rev)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	want := &changelog.Entry{
		Number: 1,
		Type:   changelog.Bugfix,
		URL: &url.URL{
			Scheme: "https",
			Host:   "github.com",
			Path:   "/influxdata/changelog/pull/1",
		},
		Message: "Fix a bug",
		Version: changelog.MustVersion("1.4.0"),
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Fatalf("unexpected changelog entry:\n%s", diff)
	}
}

func TestGitHubUpdater_ReleaseCandidate(t *testing.T) {
	u := GitHub("influxdata", "changelog", nil).(*githubUpdater)
	u.pullRequest = func(u *githubUpdater, n int) (*octokit.PullRequest, error) {
		if got, want := n, 1; got != want {
			t.Errorf("unexpected pull request number: %d != %d", got, want)
		}
		return &octokit.PullRequest{
			Base: octokit.PullRequestCommit{
				Ref: "master",
			},
		}, nil
	}
	u.labels = func(u *githubUpdater, n int) ([]octokit.Label, error) {
		if got, want := n, 1; got != want {
			t.Errorf("unexpected pull request number: %d != %d", got, want)
		}
		return []octokit.Label{
			{Name: "kind/feature request"},
		}, nil
	}
	u.lastTag = func(rev string) (string, error) {
		if got, want := rev, "abc"; got != want {
			t.Errorf("unexpected revision: %s != %s", got, want)
		}
		return "v1.3.0rc0", nil
	}

	rev := NewRevision("abc",
		"Merge pull request #1 from influxdata/add-new-feature",
		"Add a new feature",
	)

	got, err := u.NewEntry(rev)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	want := &changelog.Entry{
		Number: 1,
		Type:   changelog.FeatureRequest,
		URL: &url.URL{
			Scheme: "https",
			Host:   "github.com",
			Path:   "/influxdata/changelog/pull/1",
		},
		Message: "Add a new feature",
		Version: changelog.MustVersion("1.3.0"),
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Fatalf("unexpected changelog entry:\n%s", diff)
	}
}

func TestGitHubUpdater_DefaultVersion(t *testing.T) {
	u := GitHub("influxdata", "changelog", nil).(*githubUpdater)
	u.pullRequest = func(u *githubUpdater, n int) (*octokit.PullRequest, error) {
		if got, want := n, 1; got != want {
			t.Errorf("unexpected pull request number: %d != %d", got, want)
		}
		return &octokit.PullRequest{
			Base: octokit.PullRequestCommit{
				Ref: "master",
			},
		}, nil
	}
	u.labels = func(u *githubUpdater, n int) ([]octokit.Label, error) {
		if got, want := n, 1; got != want {
			t.Errorf("unexpected pull request number: %d != %d", got, want)
		}
		return []octokit.Label{
			{Name: "kind/feature request"},
		}, nil
	}
	u.lastTag = func(rev string) (string, error) {
		if got, want := rev, "abc"; got != want {
			t.Errorf("unexpected revision: %s != %s", got, want)
		}
		return "", nil
	}

	rev := NewRevision("abc",
		"Merge pull request #1 from influxdata/add-new-feature",
		"Add a new feature",
	)

	got, err := u.NewEntry(rev)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	want := &changelog.Entry{
		Number: 1,
		Type:   changelog.FeatureRequest,
		URL: &url.URL{
			Scheme: "https",
			Host:   "github.com",
			Path:   "/influxdata/changelog/pull/1",
		},
		Message: "Add a new feature",
		Version: changelog.MustVersion("1.0.0"),
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Fatalf("unexpected changelog entry:\n%s", diff)
	}
}

func TestGitHubUpdater_PatchRelease(t *testing.T) {
	u := GitHub("influxdata", "changelog", nil).(*githubUpdater)
	u.pullRequest = func(u *githubUpdater, n int) (*octokit.PullRequest, error) {
		if got, want := n, 1; got != want {
			t.Errorf("unexpected pull request number: %d != %d", got, want)
		}
		return &octokit.PullRequest{
			Base: octokit.PullRequestCommit{
				Ref: "1.3",
			},
		}, nil
	}
	u.labels = func(u *githubUpdater, n int) ([]octokit.Label, error) {
		if got, want := n, 1; got != want {
			t.Errorf("unexpected pull request number: %d != %d", got, want)
		}
		return []octokit.Label{
			{Name: "kind/bug"},
		}, nil
	}
	u.lastTag = func(rev string) (string, error) {
		if got, want := rev, "abc"; got != want {
			t.Errorf("unexpected revision: %s != %s", got, want)
		}
		return "v1.3.0", nil
	}

	rev := NewRevision("abc",
		"Merge pull request #1 from influxdata/fix-a-bug",
		"Fix a bug",
	)

	got, err := u.NewEntry(rev)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	want := &changelog.Entry{
		Number: 1,
		Type:   changelog.Bugfix,
		URL: &url.URL{
			Scheme: "https",
			Host:   "github.com",
			Path:   "/influxdata/changelog/pull/1",
		},
		Message: "Fix a bug",
		Version: changelog.MustVersion("1.3.1"),
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Fatalf("unexpected changelog entry:\n%s", diff)
	}
}

func TestGitHubUpdater_NoEntry(t *testing.T) {
	rev := NewRevision("abc",
		"This is a standard commit",
		"A commit message!",
	)
	u := GitHub("influxdata", "changelog", nil)
	if _, err := u.NewEntry(rev); err != changelog.ErrNoEntry {
		t.Fatalf("unexpected error: got=%v want=%v", err, changelog.ErrNoEntry)
	}
}

type Revision struct {
	id, subject, message string
}

func NewRevision(id, subject, message string) *Revision {
	return &Revision{
		id:      id,
		subject: subject,
		message: message,
	}
}

func (r *Revision) ID() string {
	return r.id
}

func (r *Revision) Subject() string {
	return r.subject
}

func (r *Revision) Message() string {
	return r.message
}
