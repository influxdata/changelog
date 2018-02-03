package git_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"reflect"
	"testing"

	"github.com/influxdata/changelog/git"
)

var RepoDir string

func TestRoot(t *testing.T) {
	dir, err := git.Root()
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	} else if got, want := dir, RepoDir; got != want {
		t.Fatalf("unexpected directory: got=%s want=%s", got, want)
	}
}

func TestMerges(t *testing.T) {
	got, err := git.Merges()
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	want := []string{"fbb1c17877bdc849d774d5da3c9d621ee0104e90"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected revisions: got=%v want=%v", got, want)
	}
}

func TestLastTag(t *testing.T) {
	tag, err := git.LastTag()
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	} else if got, want := tag, "v1.0.0"; got != want {
		t.Fatalf("unexpected tag: got=%v want=%v", got, want)
	}
}

func TestLastTag_NotExist(t *testing.T) {
	tag, err := git.LastTag("08bef2922de13ae30779128e6d3c218a26042749")
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	} else if got, want := tag, ""; got != want {
		t.Fatalf("unexpected tag: got=%v want=%v", got, want)
	}
}

func TestShow(t *testing.T) {
	rev, err := git.Show("fbb1c17877bdc849d774d5da3c9d621ee0104e90")
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	if got, want := rev.ID(), "fbb1c17877bdc849d774d5da3c9d621ee0104e90"; got != want {
		t.Errorf("unexpected id: got=%v want=%v", got, want)
	}
	if got, want := rev.Subject(), "Merge pull request #1 from influxdata/sample-repo"; got != want {
		t.Errorf("unexpected subject: got=%v want=%v", got, want)
	}
	if got, want := rev.Message(), "Modify the README"; got != want {
		t.Errorf("unexpected message: got=%v want=%v", got, want)
	}
}

func TestMain(m *testing.M) {
	tmpdir, err := ioutil.TempDir("", "git-test")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Could not create temporary directory: %s\n", err)
		os.Exit(1)
	}

	// Extract the git archive into this directory.
	cmd := exec.Command("git", "clone", "repo.bundle", tmpdir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: Could not clone sample repository into temporary directory: %s\n", err)
		os.RemoveAll(tmpdir)
		os.Exit(1)
	}

	os.Chdir(tmpdir)
	RepoDir, _ = os.Getwd()
	status := m.Run()
	os.RemoveAll(tmpdir)
	os.Exit(status)
}
