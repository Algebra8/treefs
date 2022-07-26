Package treefs provides functionality to print a simple graph of an fs.FS using
the template of the [`tree` command](https://en.wikipedia.org/wiki/Tree_(command)).

The version of `tree` whose graph is mimicked is tree v2.0.2 (c) 1996 - 2022 by
Steve Baker, Thomas Moore, Francesc Rocher, Florian Sesser, Kyosuke Tokoro.

To get the graph representation and metadata of an `fs.FS`, construct a `TreeFS`
and use its `String` method.

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

```go
var testdataFS fs.FS // assume this is embedded or read with os.DirFS
tfs, err := New(testdataFS, "testdata")
if err != nil {
    log.Fatal(err)
}
fmt.Println(tfs)
```

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

```go
var args []Arg // see internal/examples/multi
multitfs, err := NewMutli(args...)
if err != nil {
    log.Fatal(err)
}
fmt.Println(multitfs)
```

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

Options can be provided with `Opt`s. 

For example, to display hidden directories and files (which are excluded by default), use the `Hidden` option:

```go
tree, err := Tree(fsys, ".", Hidden)
if err != nil {
    log.Fatal(err)
}
fmt.Println(tree)
```
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

    3 directories, 15 files

The `DirOnly` option only displays directories:

```go
tree, err := Tree(fsys, ".", DirOnly)
if err != nil {
    log.Fatal(err)
}
fmt.Println(tree)
```

    testdata
    └── a
        ├── b
        │   └── d
        └── c

    4 directories

`FullPathPrefix` includes the full path prefix for each file:

```go
tree, err := Tree(fsys, ".", FullPathPrefix)
if err != nil {
    log.Fatal(err)
}
fmt.Println(tree)
```

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

    4 directories, 9 files

`Level` sets the max display depth of the directory tree:

```go
tree, err := Tree(fsys, ".", Level(2))
if err != nil {
    log.Fatal(err)
}
fmt.Println(tree)
```

    testdata
    └── a
        ├── a1.test
        ├── a2.test
        ├── a3.test
        ├── b
        └── c

    3 directories, 3 files

See [`examples`](https://github.com/Algebra8/treefs/tree/main/examples) for example usage.
