# gometer #

Metrics gathering library.

Advantages over gometrics:
- Lower overhead (see the various benchmarks)
- Even more modular and flexible architecture.
- Is global so isn't hindered by function calls.
- Doesn't couple the output with the internal go representation of metrics.
- Supports all the existing gometrics features and a few more.

Disadvantages over gometrics:
- Is a bit more intrusive in the existing object classes.
- Makes it harder to document metrics (can be worked around).
- Doesn't impose naming convention on metrics.

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


## License ##

The source code is available under the Apache License. See the LICENSE file for
more details.
