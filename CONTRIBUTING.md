# Developing `go-config-yourself`

This project uses [Go Modules](https://github.com/golang/go/wiki/Modules), `go` version 1.11 or greater is required. This project's folder layout is inspired by the [golang-standards/project-layout](https://github.com/golang-standards/project-layout) repository.

## MacOS

Developing on MacOS requires [homebrew](https://brew.sh/) to be installed on the system. Setting up your system for it is as easy as:

```sh
make setup-dev
```

## Other OS

On other systems, you should install the following components:

- [bats](https://github.com/bats-core/bats-core)
- [gpg](https://gnupg.org/)
- [gpgme](https://gnupg.org/software/gpgme/index.html)

The following packages are required for cross-compiling compressed binaries:

- [upx](https://upx.github.io/)
- [docker](https://www.docker.com/)

# Testing

See [test/README.md](test/README.md).

# Releasing

See [RELEASING.md](RELEASING.md).
