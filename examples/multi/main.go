/*
MIT License

Copyright (c) 2022-present Milad Michael Nasrollahi

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/

package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/Algebra8/treefs"
)

var (
	hidden        bool
	dirOnly       bool
	fullFilePath  bool
	maxDepthLevel int
)

func init() {
	flag.BoolVar(&hidden, "a", false, `
Include directory entries whose names begin with a dot ('.') except for . and 
...`[1:])
	flag.BoolVar(&dirOnly, "d", false, "List directoris only")
	flag.BoolVar(&fullFilePath, "f", false, "Prints the full path prefix for each file")
	flag.IntVar(&maxDepthLevel, "L", -1, "Max display depth of the directory tree")
}

func main() {
	flag.Parse()

	args := flag.Args()
	if len(args) < 1 {
		fmt.Fprintf(os.Stderr, "%s [-adfL] [directory ...]\n", args[0])
		os.Exit(1)
	}

	var opts []treefs.Opt
	if hidden {
		// Allow hidden directories and entries to be shown.
		opts = append(opts, treefs.Hidden)
	}
	if dirOnly {
		opts = append(opts, treefs.DirOnly)
	}
	if fullFilePath {
		opts = append(opts, treefs.FullPathPrefix)
	}
	// Level is idempotent if maxDepthLevel is less than zero (default).
	opts = append(opts, treefs.Level(maxDepthLevel))

	var tfsArgs []treefs.Arg
	for _, dir := range args {
		tfsArgs = append(tfsArgs, treefs.Arg{
			Fsys: os.DirFS(dir),
			Name: dir,
			Opts: opts,
		})
	}

	tfs, err := treefs.NewMulti(tfsArgs...)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	fmt.Println(tfs)
}
