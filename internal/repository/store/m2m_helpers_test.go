package store

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

func sortStrs(s []string) []string { sort.Strings(s); return s }

func TestDiffJunctionSets(t *testing.T) {
	cases := []struct {
		name             string
		current, target  []string
		wantAdd, wantRem []string
	}{
		{
			name: "both empty",
		},
		{
			name:    "add all",
			target:  []string{"a", "b", "c"},
			wantAdd: []string{"a", "b", "c"},
		},
		{
			name:    "remove all",
			current: []string{"a", "b"},
			wantRem: []string{"a", "b"},
		},
		{
			name:    "no diff",
			current: []string{"a", "b"},
			target:  []string{"a", "b"},
		},
		{
			name:    "partial overlap",
			current: []string{"a", "b", "c"},
			target:  []string{"b", "c", "d"},
			wantAdd: []string{"d"},
			wantRem: []string{"a"},
		},
		{
			name:    "fully disjoint",
			current: []string{"a", "b"},
			target:  []string{"x", "y"},
			wantAdd: []string{"x", "y"},
			wantRem: []string{"a", "b"},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			add, rem := diffJunctionSets(tc.current, tc.target)
			assert.Equal(t, sortStrs(tc.wantAdd), sortStrs(add))
			assert.Equal(t, sortStrs(tc.wantRem), sortStrs(rem))
		})
	}
}
