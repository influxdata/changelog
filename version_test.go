package changelog_test

import (
	"reflect"
	"testing"

	"github.com/influxdata/changelog"
)

func TestVersion(t *testing.T) {
	for _, tt := range []struct {
		s   string
		exp []int
		err string
	}{
		{s: "1.2.3", exp: []int{1, 2, 3}},
		{s: "0.3.8.2", exp: []int{0, 3, 8, 2}},
		{s: "v1.2.3", err: "strconv.Atoi: parsing \"v1\": invalid syntax"},
	} {
		t.Run(tt.s, func(t *testing.T) {
			v, err := changelog.NewVersion(tt.s)
			if err != nil {
				if tt.err == "" {
					t.Fatalf("unexpected error: %s", err)
				} else if got, want := err.Error(), tt.err; got != want {
					t.Fatalf("unexpected error: got=%v, want=%v", got, want)
				}
			} else {
				if tt.err != "" {
					t.Fatal("expected error")
				} else if !reflect.DeepEqual(v.Segments(), tt.exp) {
					t.Fatalf("unexpected segments: got=%v, want=%v", v.Segments(), tt.exp)
				}
			}
		})
	}
}

func TestVersion_Compare(t *testing.T) {
	for _, tt := range []struct {
		name  string
		s     string
		other string
		value int
	}{
		{
			name:  "Equal",
			s:     "1.2.3",
			other: "1.2.3",
			value: 0,
		},
		{
			name:  "TrailingZeros",
			s:     "1.2.3.0",
			other: "1.2.3",
			value: 1,
		},
		{
			name:  "LessThan",
			s:     "1.3.0",
			other: "2.4.7",
			value: -1,
		},
		{
			name:  "GreaterThan",
			s:     "7.8.2",
			other: "1.2.7",
			value: 1,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			v1, v2 := changelog.MustVersion(tt.s), changelog.MustVersion(tt.other)
			if got, want := v1.Compare(v2), tt.value; got != want {
				t.Fatalf("unexpected value: got=%d want=%d", got, want)
			}
		})
	}
}
