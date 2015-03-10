# gometer #

Metrics gathering library.

## Installation ##

You can download the code via the usual go utilities:

```
go get github.com/datacratic/gometer/meter
```

To build the code and run the test suite along with several static analysis
tools, use the provided Makefile:

```
make test
```

Note that the usual go utilities will work just fine but we require that all
commits pass the full suite of tests and static analysis tools.

# Documentation #

Documentation is available at
[godoc](https://godoc.org/github.com/datacratic/gometer/meter) and usage
examples are available in this [**test suite**](meter/example_test.go).

## License ##

The source code is available under the Apache License. See the LICENSE file for
more details.
