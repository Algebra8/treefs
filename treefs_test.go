package treefs

import (
	"embed"
	"fmt"
	"strings"
	"testing"
	"testing/fstest"
)

//go:embed testdata/*
var testFS embed.FS

func TestTreeFSWithEmbedFS(t *testing.T) {
	tests := []struct {
		name     string
		expected string
	}{
		{
			name: "testdata/a",
			// Expected output from calling `tree testdata/a` from treefs' root
			// directory.
			expected: `
testdata/a
├── a1.test
├── a2.test
├── a3.test
├── b
│   ├── b1.test
│   ├── b2.test
│   ├── b3.test
│   └── d
│       └── d1.test
└── c
    ├── c1.test
    └── c2.test

3 directories, 9 files`[1:],
		},
		{
			name: "testdata/a/b",
			expected: `
testdata/a/b
├── b1.test
├── b2.test
├── b3.test
└── d
    └── d1.test

1 directory, 4 files`[1:],
		},
	}

	for _, tc := range tests {
		t.Run(fmt.Sprintf("testing %s", tc.name), func(t *testing.T) {
			tfs, err := New(testFS, tc.name)
			if err != nil {
				t.Fatal(err)
			}
			got := tfs.String()

			compare(t, got, tc.expected)
		})
	}
}

func TestTreeFSWithMapFS(t *testing.T) {
	tests := []struct {
		name     string
		mapfs    fstest.MapFS
		expected string
	}{
		{
			name: ".",
			mapfs: fstest.MapFS{
				"a1.test": {},
				"a2.test": {},
				"a3.test": {},

				"b/b1.test": {},
				"b/b2.test": {},
				"b/b3.test": {},

				"b/d/d1.test": {},

				"c/c1.test": {},
				"c/c2.test": {},
			},
			expected: `
.
├── a1.test
├── a2.test
├── a3.test
├── b
│   ├── b1.test
│   ├── b2.test
│   ├── b3.test
│   └── d
│       └── d1.test
└── c
    ├── c1.test
    └── c2.test

3 directories, 9 files`[1:],
		},
	}

	for _, tc := range tests {
		t.Run(fmt.Sprintf("testing %s", tc.name), func(t *testing.T) {
			tfs, err := New(tc.mapfs, tc.name)
			if err != nil {
				t.Fatal(err)
			}
			got := tfs.String()

			compare(t, got, tc.expected)
		})
	}
}

func compare(t *testing.T, got, expected string) {
	if strings.Compare(got, expected) != 0 {
		t.Fatalf("mismatch!\nexpected:\n%s\ngot:\n%s\n",
			expected, got)
	}
}
