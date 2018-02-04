package changelog

import (
	"errors"
	"net/url"
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
