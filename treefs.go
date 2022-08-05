// Package treefs provides functionality to print a simple graph of an fs.FS
// using the template of the `tree` command.
//
// The version of `tree` whose graph is mimicked is tree v2.0.2 (c) 1996 - 2022
// by Steve Baker, Thomas Moore, Francesc Rocher, Florian Sesser, Kyosuke
// Tokoro.
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

// Tree returns the graph, and metadata, of the fs.FS fsys with name name.
func Tree(fsys fs.FS, name string, opts ...Opt) (string, error) {
	tfs, err := New(fsys, name, opts...)
	if err != nil {
		return "", err
	}
	return tfs.String(), nil
}

// Graph returns only the graph of the fs.FS fsys with name name.
func Graph(fsys fs.FS, name string, opts ...Opt) (string, error) {
	tfs, err := New(fsys, name, opts...)
	if err != nil {
		return "", err
	}
	return tfs.Graph(), nil
}

// Meta returns only the stringified metadata of the fs.FS fsys with name name.
func Meta(fsys fs.FS, name string, opts ...Opt) (string, error) {
	tfs, err := New(fsys, name, opts...)
	if err != nil {
		return "", err
	}
	return tfs.Meta(), nil
}

// New returns a TreeFS whose stringer interface implementation returns the
// graph for the fs.FS fsys and name name, similar to the `tree` command.
//
// It makes use of fs.ReadDir to walk fsys.
func New(fsys fs.FS, name string, opts ...Opt) (tfs TreeFS, err error) {
	tfs = TreeFS{
		fsys: fsys,
		tree: []string{name},
	}
	for _, opt := range opts {
		opt(&tfs)
	}

	// Since the filesystem fsys does not contain any file within it by the
	// name "../*", we substitute name for "." if a directory from any level
	// above CWD is provided.
	// Also, if name is "." (whether provided or due to the fact that it
	// contains a "../") pathPrefix is set to name (before the overwrite) for
	// use in case the FullPathPrefix Opt was applied to tfs.
	if strings.Contains(name, "../") || name == "." {
		tfs.pathPrefix = name
		name = "."
	}

	err = treeFSWithPrefix(&tfs, name, "", 0)
	return
}

// Arg represents argument pairs for aggregate TreeFS constructs using
// NewMulti.
type Arg struct {
	Fsys fs.FS
	Name string
	Opts []Opt
}

// NewMulti returns an aggregate TreeFS.
//
// The graph of each fs.FS, name pair are separated by newlines and the
// metadata is aggregated.
//
// It makes use of fs.ReadDir to walk fsys.
func NewMulti(args ...Arg) (tfs TreeFS, err error) {
	for _, arg := range args {
		var tfs2 TreeFS
		if tfs2, err = New(arg.Fsys, arg.Name, arg.Opts...); err != nil {
			return
		}

		tfs.tree = append(tfs.tree, tfs2.tree...)
		tfs.NDirs += tfs2.NDirs
		tfs.NFiles += tfs2.NFiles
	}

	return
}

// TreeFS contains the required information to construct a graph for an fs.FS.
type TreeFS struct {
	fsys fs.FS
	tree []string
	// The path prefix for cases where the fs.FS has a name that contains "."
	// or "../".
	//
	// It should only be a non-zero valued string when name is one of the
	// aforementioned cases and is only relevant when fullPathPrefix is true.
	pathPrefix string

	NDirs  int // the number of directories that exist within an fs.FS
	NFiles int // the number of files that exist within an fs.Fs

	// Opts ...
	hidden         bool // allow hidden directories and entries
	dirOnly        bool // list directories only
	fullPathPrefix bool // includes the full path prefix for each file
	level          int  // max display depth of the directory tree
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

	if t.dirOnly {
		return fmt.Sprintf("%d %s", t.NDirs, dirs)
	}

	files := "files"
	if t.NFiles == 1 {
		files = "file"
	}

	return fmt.Sprintf("%d %s, %d %s", t.NDirs, dirs, t.NFiles, files)
}

// Filter the displaying of entries based on t's internal state.
func (t TreeFS) allow(entry fs.DirEntry) bool {
	// Disallow hidden entries if t.hidden is false.
	name := entry.Name()
	isHidden := strings.HasPrefix(name, ".") && name != "." && name != "..."
	if isHidden && !t.hidden {
		return false
	}

	// Skip if t.DirOnly and entry is not a directory.
	if t.dirOnly && !entry.IsDir() {
		return false
	}

	return true
}

// Append the prefix, connector, name combo to the tree t.
func (t *TreeFS) append(prefix, connector, dirPath, name string) {
	if !t.fullPathPrefix {
		t.tree = append(t.tree, fmt.Sprintf("%s%s %s", prefix, connector, name))
		return
	}

	if t.pathPrefix != "" {
		t.tree = append(t.tree, fmt.Sprintf("%s%s %s/%s", prefix, connector, t.pathPrefix, path.Join(dirPath, name)))
		return
	}

	t.tree = append(t.tree, fmt.Sprintf("%s%s %s", prefix, connector, path.Join(dirPath, name)))
}

// Recursively generate the tree of the TreeFS treefs.
//
// XXX(algebra8):
//	This implementation for recursively creating a filesystem tree is inspired
//	by the Python tutorial "Build a Python Directory Tree Generator for the
//	Command Line" at realpython.com
//	(https://realpython.com/directory-tree-generator-python/).
//
//	Credits to the author, Leodanis Pozo Ramos.
func treeFSWithPrefix(tfs *TreeFS, name, prefix string, lvl int) (err error) {
	// Return if max level has been set and reached.
	if tfs.level > 0 && lvl == tfs.level {
		return
	}

	var entries []fs.DirEntry
	if entries, err = fs.ReadDir(tfs.fsys, name); err != nil {
		return
	}
	numEntries := len(entries)

	for i, entry := range entries {
		if !tfs.allow(entry) {
			continue
		}

		connector := teeConnector
		if i == numEntries-1 {
			connector = elbowConnector
		}

		if entry.IsDir() {
			tfs.NDirs++
			// XXX(algebra8):
			// 	One benefit to using addDir as a separate function is the
			// 	handling of prefix state.
			// 	The outer prefix won't be affected by the state change in
			// 	addDir so recursion handles any necessary prefix trimming.
			if err = addDir(tfs, addDirArgs{
				path:      name,
				name:      entry.Name(),
				idx:       i,
				numFiles:  numEntries,
				prefix:    prefix,
				connector: connector,
				lvl:       lvl,
			}); err != nil {
				return
			}
			continue
		}

		tfs.NFiles++
		tfs.append(prefix, connector, name, entry.Name())
	}

	return
}

// Container for addDir args.
type addDirArgs struct {
	path, name         string
	idx, numFiles, lvl int
	prefix, connector  string
}

func addDir(tfs *TreeFS, args addDirArgs) error {
	tfs.append(args.prefix, args.connector, args.path, args.name)

	if args.idx != args.numFiles-1 {
		args.prefix += pipePrefix
	} else {
		args.prefix += spacePrefix
	}

	return treeFSWithPrefix(tfs, path.Join(args.path, args.name), args.prefix, args.lvl+1)
}

// Opt defines an optional argument for generating an fs.FS's tree.
type Opt func(*TreeFS)

// Hidden allows hidden directories and entries.
func Hidden(t *TreeFS) {
	t.hidden = true
}

// DirOnly includes only directories.
func DirOnly(t *TreeFS) {
	t.dirOnly = true
}

// FullPathPrefix includes the full path prefix for each file.
func FullPathPrefix(t *TreeFS) {
	t.fullPathPrefix = true
}

// Level sets the max display depth of the directory tree.
func Level(lvl int) Opt {
	return func(tfs *TreeFS) {
		// Ignore if lvl <= 0.
		if lvl <= 0 {
			return
		}
		tfs.level = lvl
	}
}
