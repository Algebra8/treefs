/*
Package treefs provides functionality to print a simple graph of an fs.FS using
the template of the `tree` command.

The version of `tree` whose graph is mimicked is tree v2.0.2 (c) 1996 - 2022 by
Steve Baker, Thomas Moore, Francesc Rocher, Florian Sesser, Kyosuke Tokoro.

To get the graph representation and metadata of an fs.FS, construct a TreeFS
and use its String method.

Consider the following directory (whose graph was generated by tree v2.0.2)

    testdata
    └── a
        ├── a1.test
        ├── a2.test
        ├── a3.test
        ├── b
        │   ├── b1.test
        │   ├── b2.test
        │   ├── b3.test
        │   └── d
        │       └── d1.test
        └── c
            ├── c1.test
            └── c2.test

    4 directories, 9 files

Its graph could be constructed using treefs in the following way

	var testdataFS fs.FS // assume this is embeded or read with os.DirFS
	tfs, err := New(testdataFS, "testdata")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(tfs)

which would generate the following output

    testdata
    └── a
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

    4 directories, 9 files

Aggregated trees are allowed as well:

	var args []Arg // see internal/examples/multi
	multitfs, err := NewMutli(args...)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(multitfs)

could result in the following aggregate

    .
    └── main.go
    ../../../../treefs
    ├── LICENSE
    ├── go.mod
    ├── internal
    │   └── examples
    │       ├── multi
    │       │   └── main.go
    │       └── single
    │           └── main.go
    ├── testdata
    │   └── a
    │       ├── a1.test
    │       ├── a2.test
    │       ├── a3.test
    │       ├── b
    │       │   ├── b1.test
    │       │   ├── b2.test
    │       │   ├── b3.test
    │       │   └── d
    │       │       └── d1.test
    │       └── c
    │           ├── c1.test
    │           └── c2.test
    ├── treefs.go
    └── treefs_test.go

    9 directories, 16 files

*/
package treefs

import (
	"fmt"
	"io/fs"
	"path"
	"strings"
)

const (
	elbowConnector = "└──"
	teeConnector   = "├──"

	pipePrefix  = "│   "
	spacePrefix = "    "
)

// New returns a TreeFS whose stringer interface implementation returns the
// graph for the fs.FS fsys and name name, similar to the `tree` command.
//
// It makes use of fs.ReadDir to walk fsys.
func New(fsys fs.FS, name string) (treeFS TreeFS, err error) {
	treeFS = TreeFS{
		fsys: fsys,
		tree: []string{name},
	}
	// Since the filesystem fsys does not contain any file within it by the
	// name "../*", we substitute name for "." if a directory from any level
	// above CWD is provided.
	if strings.Contains(name, "../") {
		name = "."
	}

	err = treeFSWithPrefix(&treeFS, name, "")
	return
}

// Arg represents argument pairs for aggregate TreeFS constructs using
// NewMulti.
type Arg struct {
	Fsys fs.FS
	Name string
}

// NewMulti returns an aggregate TreeFS.
//
// The graph of each fs.FS, name pair are separated by newlines and the
// metadata is aggregated.
//
// It makes use of fs.ReadDir to walk fsys.
func NewMulti(args ...Arg) (treeFS TreeFS, err error) {
	for _, arg := range args {
		var treeFS2 TreeFS
		if treeFS2, err = New(arg.Fsys, arg.Name); err != nil {
			return
		}

		treeFS.tree = append(treeFS.tree, treeFS2.tree...)
		treeFS.NDirs += treeFS2.NDirs
		treeFS.NFiles += treeFS2.NFiles
	}

	return
}

// TreeFS contains the required information to construct a graph for an fs.FS.
type TreeFS struct {
	fsys fs.FS
	tree []string

	NDirs  int // the number of directories that exist within an fs.FS
	NFiles int // the number of files that exist within an fs.Fs
}

// String implements the stringer interface for TreeFS.
//
// It returns the stringified graph of the TreeFS t with metadata at the
// bottom, similar to the `tree` command.
func (t TreeFS) String() string {
	return t.Graph() + "\n\n" + t.Meta()
}

// Graph returns the stringified graph of the TreeFS t without any metadata.
func (t TreeFS) Graph() string {
	return strings.Join(t.tree, "\n")
}

// Meta returns the stringified metadata for the TreeFS t.
func (t TreeFS) Meta() string {
	dirs := "directories"
	if t.NDirs == 1 {
		dirs = "directory"
	}

	files := "files"
	if t.NFiles == 1 {
		files = "file"
	}

	return fmt.Sprintf("%d %s, %d %s", t.NDirs, dirs, t.NFiles, files)
}

func treeFSWithPrefix(treeFS *TreeFS, name, prefix string) (err error) {
	var entries []fs.DirEntry
	if entries, err = fs.ReadDir(treeFS.fsys, name); err != nil {
		return
	}
	numEntries := len(entries)

	for i, entry := range entries {
		connector := teeConnector
		if i == numEntries-1 {
			connector = elbowConnector
		}

		if entry.IsDir() {
			treeFS.NDirs++

			name1 := path.Join(name, entry.Name())
			// XXX:
			// 	One benefit to using addDir as a separate function is the
			// 	handling of prefix state.
			// 	The outer prefix won't be affected by the state change in
			// 	addDir so recursion handles any necessary prefix trimming.
			if err = addDir(treeFS, name1, i, numEntries, prefix, connector); err != nil {
				return
			}
			continue
		}

		treeFS.NFiles++

		treeFS.tree = append(treeFS.tree, fmt.Sprintf("%s%s %s", prefix, connector, entry.Name()))
	}

	return
}

func addDir(treeFS *TreeFS, path string, idx, numFiles int, prefix, connector string) error {
	sep := strings.Split(path, "/")
	name := sep[len(sep)-1]
	treeFS.tree = append(treeFS.tree, fmt.Sprintf("%s%s %s", prefix, connector, name))

	if idx != numFiles-1 {
		prefix += pipePrefix
	} else {
		prefix += spacePrefix
	}

	return treeFSWithPrefix(treeFS, path, prefix)
}
