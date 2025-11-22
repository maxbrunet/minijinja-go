# minijinja-go

[![License](https://img.shields.io/github/license/maxbrunet/minijinja-go)](https://github.com/maxbrunet/minijinja-go/blob/main/LICENSE)
[![Go reference](https://pkg.go.dev/badge/github.com/maxbrunet/minijinja-go/v2.svg)](https://pkg.go.dev/github.com/maxbrunet/minijinja-go/v2)

`minijinja-go` is a module that wraps
[MiniJinja](https://github.com/mitsuhiko/minijinja) into a Go library using CGo.
This is experimental.

For an example look into [example_test.go](example_test.go).

## Installation

```shell
go get github.com/maxbrunet/minijinja-go/v2@latest
```

Build `minijinja-cabi` library:

```shell
go get -tool github.com/maxbrunet/minijinja-go/v2/hack/build-minijinja-cabi

MJGO_VERSION="$(go list -m all | awk '/github.com\/maxbrunet\/minijinja-go\/v2/{print $2}')"
MJGO_PATH="$(go env GOMODCACHE)/github.com/maxbrunet/minijinja-go/v2@${MJGO_VERSION}"

# Ensure the module path is writeable
chmod +w "${MJGO_PATH}"

# Build
go tool build-minijinja-cabi --mod-dir="${MJGO_PATH}"

# Remove rewrite permission (optional)
chmod -w "${MJGO_PATH}"

# If using vendored modules, simply run:
# go tool build-minijinja-cabi --mod-dir=vendor/github.com/maxbrunet/minijinja-go/v2
```

The dynamic library must be available during runtime of any dependent program.

## License and Links

- [Issue Tracker](https://github.com/maxbrunet/minijinja-go/issues)
- License: [Apache-2.0](https://github.com/maxbrunet/minijinja-go/blob/main/LICENSE)
