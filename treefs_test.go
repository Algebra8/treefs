package treefs

import (
	"bytes"
	"embed"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"
	"testing/fstest"
)

var diffFlag = flag.Bool("diff", false, `
include a diff between "got" and "expected when any of the tests fail (uses 
unix diff utility)`[1:])

//go:embed testdata/*
var testFS embed.FS

func TestTreeFSWithEmbedFS(t *testing.T) {
	tests := []struct {
		tcname   string // test case's name
		name     string
		opts     []Opt
		expected string
	}{
		{
			tcname: "testdata/a",
			name:   "testdata/a",
			// Expected output from calling `tree testdata/a` from directory
			// containing testdata/.
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
			tcname: "testdata/a/b",
			name:   "testdata/a/b",
			// Expected output from calling `tree testdata/a/b` from directory
			// containing testdata/.
			expected: `
testdata/a/b
├── b1.test
├── b2.test
├── b3.test
└── d
    └── d1.test

1 directory, 4 files`[1:],
		},
		{
			tcname: "level=2",
			name:   "testdata",
			opts: []Opt{
				Level(2),
			},
			// Expected output from calling `tree -L 2 testdata` from directory
			// containing testdata/.
			expected: `
testdata
└── a
    ├── a1.test
    ├── a2.test
    ├── a3.test
    ├── b
    └── c

3 directories, 3 files`[1:],
		},
		{
			tcname: "full path prefix",
			name:   "testdata",
			opts: []Opt{
				FullPathPrefix,
			},
			// Expected output from calling `tree -f testdata` from directory
			// containing testdata/.
			expected: `
testdata
└── testdata/a
    ├── testdata/a/a1.test
    ├── testdata/a/a2.test
    ├── testdata/a/a3.test
    ├── testdata/a/b
    │   ├── testdata/a/b/b1.test
    │   ├── testdata/a/b/b2.test
    │   ├── testdata/a/b/b3.test
    │   └── testdata/a/b/d
    │       └── testdata/a/b/d/d1.test
    └── testdata/a/c
        ├── testdata/a/c/c1.test
        └── testdata/a/c/c2.test

4 directories, 9 files`[1:],
		},
		{
			tcname: "dir only",
			name:   "testdata",
			opts: []Opt{
				DirOnly,
			},
			// Expected output from calling `tree -d testdata` from directory
			// containing testdata/.
			expected: `
testdata
└── a
    ├── b
    │   └── d
    └── c

4 directories`[1:],
		},
	}

	for _, tc := range tests {
		t.Run(fmt.Sprintf("testing %s", tc.tcname), func(t *testing.T) {
			tfs, err := New(testFS, tc.name, tc.opts...)
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
		tcname   string // test case's name
		name     string
		mapfs    fstest.MapFS
		opts     []Opt
		expected string
	}{
		{
			tcname: ".",
			name:   ".",
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
		{
			tcname: "hidden",
			name:   ".",
			mapfs: fstest.MapFS{
				".hidden1": {},
				".hidden2": {},
				".hidden3": {},

				"a1.test": {},
				"a2.test": {},
				"a3.test": {},

				"b/b1.test":  {},
				"b/b2.test":  {},
				"b/b3.test":  {},
				"b/.hidden1": {},

				"b/d/d1.test":  {},
				"b/d/.hidden1": {},

				"c/c1.test":  {},
				"c/c2.test":  {},
				"c/.hidden1": {},
			},
			opts: []Opt{
				Hidden,
			},
			expected: `
.
├── .hidden1
├── .hidden2
├── .hidden3
├── a1.test
├── a2.test
├── a3.test
├── b
│   ├── .hidden1
│   ├── b1.test
│   ├── b2.test
│   ├── b3.test
│   └── d
│       ├── .hidden1
│       └── d1.test
└── c
    ├── .hidden1
    ├── c1.test
    └── c2.test

3 directories, 15 files`[1:],
		},
		{
			tcname: "level=2",
			name:   ".",
			mapfs: fstest.MapFS{
				"a/a1.test": {},
				"a/a2.test": {},
				"a/a3.test": {},

				"a/b/b1.test": {},
				"a/b/b2.test": {},

				"a/b/d/d1.test": {},
				"a/b/d/d2.test": {},
			},
			opts: []Opt{
				Level(2),
			},
			expected: `
.
└── a
    ├── a1.test
    ├── a2.test
    ├── a3.test
    └── b

2 directories, 3 files`[1:],
		},
		{
			tcname: "full path prefix",
			name:   ".",
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
			opts: []Opt{
				FullPathPrefix,
			},
			expected: `
.
├── ./a1.test
├── ./a2.test
├── ./a3.test
├── ./b
│   ├── ./b/b1.test
│   ├── ./b/b2.test
│   ├── ./b/b3.test
│   └── ./b/d
│       └── ./b/d/d1.test
└── ./c
    ├── ./c/c1.test
    └── ./c/c2.test

3 directories, 9 files`[1:],
		},
		{
			tcname: "dir only",
			name:   ".",
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
			opts: []Opt{
				DirOnly,
			},
			expected: `
.
├── b
│   └── d
└── c

3 directories`[1:],
		},
	}

	for _, tc := range tests {
		t.Run(fmt.Sprintf("testing %s", tc.tcname), func(t *testing.T) {
			tfs, err := New(tc.mapfs, tc.name, tc.opts...)
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
		dif := ""
		if *diffFlag {
			dif = fmt.Sprintf("---\ndiff:\n%s\n", diff(t, got, expected))
		}
		t.Fatalf("mismatch!\nexpected:\n%s\ngot:\n%s\n%s",
			expected, got, dif)
	}
}

func diff(t *testing.T, a, b string) string {
	fa, err := os.Create("diff_a.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(fa.Name())
	if _, err = fa.WriteString(a); err != nil {
		t.Fatal(err)
	}

	fb, err := os.Create("diff_b.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(fb.Name())
	if _, err = fb.WriteString(b); err != nil {
		t.Fatal(err)
	}

	cmd := exec.Command("diff", fa.Name(), fb.Name())
	var outb bytes.Buffer
	cmd.Stdout = &outb
	if err = cmd.Run(); err != nil {
		if exiterr, ok := err.(*exec.ExitError); ok {
			// diff command returns an exit code of 1 when there is a diff,
			// which is what we expect in this case.
			if exiterr.ExitCode() != 1 {
				t.Fatal(err)
			}
		}
	}

	return outb.String()
}
