# Testing go-config-yourself

In a nutshell, run:

```sh
make test
```

Tests run this way will mock out the AWS KMS service.

## Unit tests

Unit tests are written in go for the `github.com/blinkhealth/go-config-yourself` package. Use the `-tags test` flag to mock out KMS.

```sh
# easily
make unit-test
# run all tests
go test ./... -tags test
# run one test
go test ./... -tags test --run TestSetValues/password
```

## Integration tests

CLI tests are written in [Bats](https://github.com/bats-core/bats-core) and these build then run the binary, emulating the perspective of a CLI user.

```sh
make integration-test
```
