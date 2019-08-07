# Developing `go-config-yourself`

Run `make os-dependencies` to install project dependencies, you'll need [`brew`](https://brew.sh/) installed


```sh
make setup-dev
```

You may also install the latest snapshot, generated after every merge to the mainline branch. Make sure you uninstall any current versions:

```sh
brew update && brew uninstall --ignore-dependencies go-config-yourself
brew install --HEAD blinkhealth/opensource-formulas/go-config-yourself
```

# Testing

See [test/README.md](test/README.md).

# Releasing

See [RELEASING.md](RELEASING.md).
