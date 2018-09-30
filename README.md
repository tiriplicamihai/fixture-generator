# fixture-generator
fixture-generator provides an easy way to generate fixtures for Go structs. It aims to reduce the time one spends writing tests.
It works using a "best effort" approach - if we can't generate data for a field then ignore it and return a partally complete fixture.

### Installation

`go get github.com/tiriplicamihai/fixture-generator/.../`, then `$GOPATH/bin/fixturegen`

### Examples

Consider we have the next struct:

```go
package test

type Person struct {
        Name             string
        Age              int
        Friends          []string
}
```
Run: `fixturegen -struct Person` and the following will be output to stdout:

```go
Person{
	Name: "QuZNvyPcKeEp",
	Age: 651906,
	Friends: []string{"GMV", "dKFjsoGexcbs", "cC"},
}
```

__Tip__: You can use `fixturegen -struct Person | gofmt` to get a formatted fixture.

### Command Line Arguments

- `struct` - required argument, expects the struct name for which the fixture will be generated

Fixtures can be generated even if field types are from other packages. It performs a recursive discovery for the types it finds in a package.

### Known Limitations

* Fixtures can not be generated for fields declared as interfaces because we can't know the actual type
* Fixtures can not be generated for a struct that has a self referencing field (we would go in an infinite loop)
* Types from vendored external packages are not supported for know

### Credits

- https://github.com/vektra/mockery was an insipration for this tool
