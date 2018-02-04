package changelog_test

import (
	"io/ioutil"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/influxdata/changelog"
)

func TestChangelog_New(t *testing.T) {
	c := changelog.New()
	c.AddEntry(&changelog.Entry{
		Number: 1,
		Type:   changelog.FeatureRequest,
		URL: &url.URL{
			Scheme: "https",
			Host:   "github.com",
			Path:   "/influxdata/changelog/pull/1",
		},
		Message: "Initial commit",
		Version: changelog.MustVersion("1.2.7"),
	})
	expect(t, c, `v1.2.7 [unreleased]
-------------------

### Features

-	[#1](https://github.com/influxdata/changelog/pull/1): Initial commit.
`)
}

func TestChangelog_NewRelease(t *testing.T) {
	c := changelog.Parse([]byte(`## v1.3.0 [2018-02-03]

### Features

-	[#1](https://github.com/influxdata/changelog/pull/1): Initial commit.
`))
	c.AddEntry(&changelog.Entry{
		Number: 2,
		Type:   changelog.FeatureRequest,
		URL: &url.URL{
			Scheme: "https",
			Host:   "github.com",
			Path:   "/influxdata/changelog/pull/2",
		},
		Message: "A new feature",
		Version: changelog.MustVersion("1.4.0"),
	})
	expect(t, c, `v1.4.0 [unreleased]
-------------------

### Features

-	[#2](https://github.com/influxdata/changelog/pull/2): A new feature.

v1.3.0 [2018-02-03]
-------------------

### Features

-	[#1](https://github.com/influxdata/changelog/pull/1): Initial commit.
`)
}

func TestChangelog_NewPatchRelease(t *testing.T) {
	c := changelog.Parse([]byte(`## v1.4.0 [unreleased]

### Features

-	[#2](https://github.com/influxdata/changelog/pull/2): A new feature.

### Bugfixes

-	[#3](https://github.com/influxdata/changelog/pull/3): An embarrassing bug.

## v1.3.0 [2018-02-03]

### Features

-	[#1](https://github.com/influxdata/changelog/pull/1): Initial commit.
`))
	c.AddEntry(&changelog.Entry{
		Number: 4,
		Type:   changelog.Bugfix,
		URL: &url.URL{
			Scheme: "https",
			Host:   "github.com",
			Path:   "/influxdata/changelog/pull/4",
		},
		Message: "An embarrassing bug",
		Version: changelog.MustVersion("1.3.1"),
	})
	expect(t, c, `v1.4.0 [unreleased]
-------------------

### Features

-	[#2](https://github.com/influxdata/changelog/pull/2): A new feature.

### Bugfixes

-	[#3](https://github.com/influxdata/changelog/pull/3): An embarrassing bug.

v1.3.1 [unreleased]
-------------------

### Bugfixes

-	[#4](https://github.com/influxdata/changelog/pull/4): An embarrassing bug.

v1.3.0 [2018-02-03]
-------------------

### Features

-	[#1](https://github.com/influxdata/changelog/pull/1): Initial commit.
`)
}

func TestChangelog_AppendRelease(t *testing.T) {
	c := changelog.Parse([]byte(`## v1.4.0 [unreleased]

### Features

-	[#2](https://github.com/influxdata/changelog/pull/2): A new feature.
`))
	c.AddEntry(&changelog.Entry{
		Number: 1,
		Type:   changelog.FeatureRequest,
		URL: &url.URL{
			Scheme: "https",
			Host:   "github.com",
			Path:   "/influxdata/changelog/pull/1",
		},
		Message: "Initial commit",
		Version: changelog.MustVersion("1.3.0"),
	})
	expect(t, c, `v1.4.0 [unreleased]
-------------------

### Features

-	[#2](https://github.com/influxdata/changelog/pull/2): A new feature.

v1.3.0 [unreleased]
-------------------

### Features

-	[#1](https://github.com/influxdata/changelog/pull/1): Initial commit.
`)
}

// A duplicate entry being added to the same section should result in the entry
// being discarded, even if the entry is in the wrong category.
func TestChangelog_DuplicateEntry(t *testing.T) {
	c := changelog.Parse([]byte(`## v1.3.0 [unreleased]

### Features

-	[#1](https://github.com/influxdata/changelog/pull/1): Initial commit.
`))
	c.AddEntry(&changelog.Entry{
		Number: 1,
		Type:   changelog.Bugfix,
		URL: &url.URL{
			Scheme: "https",
			Host:   "github.com",
			Path:   "/influxdata/changelog/pull/1",
		},
		Message: "Initial commit",
		Version: changelog.MustVersion("1.3.0"),
	})
	expect(t, c, `v1.3.0 [unreleased]
-------------------

### Features

-	[#1](https://github.com/influxdata/changelog/pull/1): Initial commit.
`)
}

func TestChanglog_AppendEntry(t *testing.T) {
	c := changelog.Parse([]byte(`## v1.3.0 [unreleased]

### Features

-	[#1](https://github.com/influxdata/changelog/pull/1): Initial commit.
`))
	c.AddEntry(&changelog.Entry{
		Number: 2,
		Type:   changelog.FeatureRequest,
		URL: &url.URL{
			Scheme: "https",
			Host:   "github.com",
			Path:   "/influxdata/changelog/pull/2",
		},
		Message: "An additional feature",
		Version: changelog.MustVersion("1.3.0"),
	})
	expect(t, c, `v1.3.0 [unreleased]
-------------------

### Features

-	[#1](https://github.com/influxdata/changelog/pull/1): Initial commit.
-	[#2](https://github.com/influxdata/changelog/pull/2): An additional feature.
`)
}

func TestChanglog_UnknownEntry(t *testing.T) {
	c := changelog.Parse([]byte(`## v1.3.0 [unreleased]

### Features

-	[#1](https://github.com/influxdata/changelog/pull/1): Initial commit.
`))
	c.AddEntry(&changelog.Entry{
		Number: 2,
		Type:   changelog.UnknownEntryType,
		URL: &url.URL{
			Scheme: "https",
			Host:   "github.com",
			Path:   "/influxdata/changelog/pull/2",
		},
		Message: "An additional feature",
		Version: changelog.MustVersion("1.3.0"),
	})
	expect(t, c, `v1.3.0 [unreleased]
-------------------

### Features

-	[#1](https://github.com/influxdata/changelog/pull/1): Initial commit.
`)
}

func TestChangelog_InsertSection(t *testing.T) {
	c := changelog.Parse([]byte(`## v1.3.0 [unreleased]

### Bugfixes

-	[#2](https://github.com/influxdata/changelog/pull/2): Fixing a bug.
`))
	c.AddEntry(&changelog.Entry{
		Number: 1,
		Type:   changelog.FeatureRequest,
		URL: &url.URL{
			Scheme: "https",
			Host:   "github.com",
			Path:   "/influxdata/changelog/pull/1",
		},
		Message: "Initial commit",
		Version: changelog.MustVersion("1.3.0"),
	})
	expect(t, c, `v1.3.0 [unreleased]
-------------------

### Features

-	[#1](https://github.com/influxdata/changelog/pull/1): Initial commit.

### Bugfixes

-	[#2](https://github.com/influxdata/changelog/pull/2): Fixing a bug.
`)
}

func expect(t *testing.T, c *changelog.Changelog, exp string) {
	tmpdir, err := ioutil.TempDir("", "changelog")
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	defer os.RemoveAll(tmpdir)

	os.Mkdir(filepath.Join(tmpdir, "a"), 0777)
	ioutil.WriteFile(
		filepath.Join(tmpdir, "a", "CHANGELOG.md"),
		[]byte(exp),
		0666,
	)
	os.Mkdir(filepath.Join(tmpdir, "b"), 0777)
	c.WriteFile(filepath.Join(tmpdir, "b", "CHANGELOG.md"))

	if out, _ := diff(
		filepath.Join(tmpdir, "a", "CHANGELOG.md"),
		filepath.Join(tmpdir, "b", "CHANGELOG.md"),
	); len(out) != 0 {
		t.Fatal(string(out))
	}
}

func diff(f1, f2 string) ([]byte, error) {
	cmd := exec.Command("diff", "-u", f1, f2)
	cmd.Stderr = os.Stderr
	return cmd.Output()
}
