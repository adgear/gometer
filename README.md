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

# Examples #

Usage examples are available in this [**test suite**](meter/example_test.go).

# Why not gometrics? #

gometer was designed to do away with some of the fundamental performance issues
of gometrics while also improving on the overall architecture. Unfortunately,
performance comes at the cost of being a little more intrusive the code base
where it's used.

Main improvements are:
- Lower overhead (CPU and allocations)
- Modular architecture for both the metrics and the handlers.
- Global poller removes the arbitrary function call boundaries.
- Metric output is not coupled to the go struct representation.

The downsides are:
- User is responsible for storing and registring the meters.

## License ##

The source code is available under the Apache License. See the LICENSE file for
more details.
