# GoconveyX

[![Build Status](https://travis-ci.org/ng-vu/goconveyx.svg?branch=master)](https://travis-ci.org/ng-vu/goconveyx)
[![Coverage Status](https://coveralls.io/repos/github/ng-vu/goconveyx/badge.svg?branch=master)](https://coveralls.io/github/ng-vu/goconveyx?branch=master)

GoconveyX extends [goconvey](https://github.com/smartystreets/goconvey) by
providing a few more functions:

- ShouldDeepEqual
- ShouldResembleSlice
- ShouldResembleByKey

# Documentation

[![GoDoc](https://img.shields.io/badge/godoc-reference-blue.svg)](http://godoc.org/github.com/ng-vu/goconveyx)

### ShouldDeepEqual

`ShouldDeepEqual` is the same as
[ShouldResemble](https://godoc.org/github.com/smartystreets/assertions#ShouldResemble)
with better error message.

### ShouldResembleSlice

`ShouldResembleSlice` does deep equal comparison on two slices without ordering.

```go
ShouldResembleSlice([]int{1, 2, 3}, []int{1, 2, 3})    // true
ShouldResembleSlice([]int{1, 2, 3}, []int{3, 1, 2})    // true
ShouldResembleSlice([]int{1, 2, 3}, []int{1, 2, 3, 1}) // false
```

### ShouldResembleByKey

`ShouldResembleByKey` does deep equal comparison on two slices sorted by given
key. It works on slices with map and struct as element. It's useful when you
want to compare rows retrieved from database.

```go
type M map[string]interface{}

ShouldResembleByKey("id")(
    []M{{"id": 1, "v": 10}, {"id": 2, "v": 20}},
    []M{{"id": 2, "v": 20}, {"id": 1, "v": 10}},
) // true

ShouldResembleByKey("id")(
    []M{{"id": 1, "v": 10}, {"id": 2, "v": 20}},
    []M{{"id": 2, "v": 10}, {"id": 1, "v": 20}},
) // false
```

# License

- [MIT License](https://opensource.org/licenses/mit-license.php)
