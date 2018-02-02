package changelog

import (
	"bytes"
	"strconv"
	"strings"
)

type Version struct {
	segments []int
}

func NewVersion(s string) (*Version, error) {
	parts := strings.Split(s, ".")
	segments := make([]int, len(parts))
	for i, part := range parts {
		val, err := strconv.Atoi(part)
		if err != nil {
			return nil, err
		}
		segments[i] = val
	}
	return &Version{segments: segments}, nil
}

func MustVersion(s string) *Version {
	v, err := NewVersion(s)
	if err != nil {
		panic(err)
	}
	return v
}

func (v *Version) Compare(other *Version) int {
	for i, segment := range v.segments {
		if i >= len(other.segments) {
			return 1
		}
		if segment < other.segments[i] {
			return -1
		} else if segment > other.segments[i] {
			return 1
		}
	}
	if len(other.segments) > len(v.segments) {
		return -1
	}
	return 0
}

func (v *Version) Equal(other *Version) bool {
	if len(v.segments) != len(other.segments) {
		return false
	}
	return v.Compare(other) == 0
}

func (v *Version) Segments() []int {
	return v.segments
}

func (v *Version) String() string {
	var buf bytes.Buffer
	for i, segment := range v.segments {
		if i > 0 {
			buf.WriteString(".")
		}
		buf.WriteString(strconv.Itoa(segment))
	}
	return buf.String()
}
