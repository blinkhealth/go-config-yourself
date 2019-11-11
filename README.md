# go-config-yourself

[![CircleCI](https://circleci.com/gh/blinkhealth/go-config-yourself.svg?style=svg)](https://circleci.com/gh/blinkhealth/go-config-yourself)

A secrets-management CLI tool and language-specific runtimes to deal with everyday application configuration right from your repository. **go-config-yourself** aims to simplify the management of secrets from your terminal and its use within application code, keeping configuration files as human-readable as possible, so change management of these files, and their secrets are easily achievable any version-control system, like git.

`go-config-yourself` comes with the following cryptographic providers that do the work of encrypting and decrypting secrets, along managing the keys for it:

- [AWS KMS](pkg/crypto/kms)
- [GPG](pkg/crypto/gpg)
- [Password](pkg/crypto/password) (scrypt)

This repository contains code and documentation for the `gcy` command-line tool. Packaged libraries to read secrets from these files are available for these languages:

- [Golang](pkg/file)
- [Javascript + Typescript](https://github.com/blinkhealth/config-yourself-javascript)
- [Python](https://github.com/blinkhealth/config-yourself-python)

---

# gcy

- [Installation](#installation)
- [Usage](#usage)
- [Config files](#config-files)
- [Developing go-config-yourself](CONTRIBUTING.md)

# Installation

## OSX

```sh
brew tap blinkhealth/opensource-formulas git@github.com:blinkhealth/opensource-formulas.git
brew install blinkhealth/opensource-formulas/go-config-yourself
```

**Install the latest snapshot**, uninstalling any current versions:

```sh
brew update && brew uninstall --ignore-dependencies go-config-yourself
brew install --HEAD blinkhealth/opensource-formulas/go-config-yourself
```

**Update** with:

```sh
brew update && brew upgrade go-config-yourself
```

## Debian/Ubuntu

```sh
add-apt-repository ppa:blinkhealth/stable
apt update
apt-get install go-config-yourself
```

## Other Linux distros:

```sh
latest=$(curl --silent "https://api.github.com/repos/blinkhealth/blink-config/releases/latest" | awk -F'"' '/tag_name/{print $4}' )
curl -vO https://github.com/blinkhealth/blink-config/releases/download/$latest/gcy-linux-amd64.tar.gz
tar xfz gcy-linux-amd64.tar.gz
cp gcy /usr/local/bin
```

---

# Usage

The command line interface is a program named `gcy` with four main commands: [`init`](#init), [`set`](#set), [`get`](#get) and [`rekey`](#rekey)

## `init`

```sh
go-config-yourself init
  [--provider PROVIDER]
  [--key KMS_KEY]
  [--password PLAINTEXT_PASSWORD]
  [--public-key GPG_IDENTITY]...
  CONFIG_FILE
```

Creates a YAML config file at `$(pwd)/${CONFIG_FILE}`. You may omit the key flag to have `gcy` query your provider for a list of them you can choose from. You can specify which encryption provider to use by specifying the `--provider` flag. By default, `gcy` will use the [AWS KMS](https://aws.amazon.com/kms/) service.

### Options:

- `--provider value`, `-p value`: The provider to encrypt values with (value is one of: [kms](pkg/crypto/kms), [gpg](pkg/crypto/gpg), [password](pkg/crypto/password))
- `--key value`: The kms key ARN to use
- `--public-key value`: A gpg public key's identity: a fingerprint like or email to use as a recipient to encrypt this file's data key. This option can be entered multiple times. If no recipients are specified, a list of available keys will be printed for the user to choose from.
- `--password value`: A password to use for encryption and decryption, also read from the environment variable `$CONFIG_PASSWORD`. To prevent your shell from remembering the password, start your command with a space: `[space]gcy ...`

```sh
# For kms
gcy init config/my-first-config.yml
# INFO Creating config at config/my-first-config.yml
# Use the arrow keys to navigate: ↓ ↑ → ←  and / toggles search
# ? Select a key to continue:
#   ▸ arn:aws:kms:us-east-1:an-account:alias/an-alias
#     arn:aws:kms:us-east-1:an-account:alias/another-alias
#     ....
#     arn:aws:kms:us-east-1:an-account:alias/and-so-on
# ↓   arn:aws:kms:us-east-1:an-account:alias/and-so-forth

# or specify the key if you know it
gcy init config/my-first-config.yml --provider kms --key arn:aws:kms:an-aws-region:an-account:alias/an-alias

cat config/my-first-config.yml
```

Outputs:

```yaml
crypto:
  key: arn:aws:kms:an-aws-region:an-account:alias/an-alias
  provider: kms
```

## `set`

`gcy set [-p,--plain-text] [-i,--input-file PATH] CONFIG_FILE KEYPATH`

Sets a value at `KEYPATH`, prompting you for the input or reading from `stdin`. If the `crypto` property does not exist in `CONFIG_FILE` and `-p|--plain-text` is not specified, `gcy` will exit with a non-zero status code. Encrypted values read through `stdin` will be later accessible as their interpreted type by golang’s default JSON parser. This means that the string `“true”` becomes the boolean `true`. If encrypting the contents of a file, you can pass its path to the `-i|--input-file` flag and `gcy` will read from it instead of `stdin`.

`set` will read up to 4096 bytes before exiting with an error, to account for AWS's KMS service limitations in this regard.

If a `defaults` or `default` file with the same extension as `CONFIG_FILE` exists in the same directory, `gcy` will add a nil value for `KEYPATH` in said file.

```sh
gcy set --plain-text config-up-there.yml someInt # user inputs "1"
gcy set --plain-text config-up-there.yml someBool # "true"
gcy set --plain-text config-up-there.yml someList # "[1,2,3]"
gcy set --plain-text config-up-there.yml someList.1 # "3"
gcy set --plain-text config-up-there.yml someList.3 # "7"
gcy set --plain-text config-up-there.yml someString # "hello i am a string"
gcy set --plain-text config-up-there.yml nestedList.0.prop # "true"
gcy set --plain-text config-up-there.yml nestedList.1.prop # "false"
gcy set --plain-text config-up-there.yml some.nested.object # "down here"
gcy set --plain-text --input-file ~/.ssh/id_rsa config-up-there.yml someFile
gcy set config-up-there.yml someSecret

# Please enter the value for "someSecret": **************

cat config-up-there.yml
```

Outputs:

```yaml
nestedList:
  - prop: true
  - prop: false
some:
  nested:
    object: down here
    secret:
      encrypted: true
      ciphertext: "D34DB33fb4d455="
      hash: "ABDCDEF0987654321"
someInt: 1
someBool: true
someFile: |
  -----BEGIN SOME KEY-----
  ...
  -----END SOME KEY-----
someList: [1, 3, 3, 7]
someString: "hello i am a string"
someSecret:
  encrypted: true
  ciphertext: "D34DB33fb4d455="
  hash: "ABDCDEF0987654321"
```

## `get`

`gcy get CONFIG_FILE KEYPATH`

Outputs the value for `KEYPATH` in `CONFIG_FILE`. `KEYPATH` is a dot delimited path to objects. The output value will be encoded as JSON if said value is a dictionary, with all encrypted values within decrypted.

```sh
gcy get config-up-there.yml some.nested.object
# Outputs:
# down here

gcy get config-up-there.yml some
```

Outputs:

```json
{
  "nested": {
    "object": "down here",
    "secret": "plaintext value of some.nested.secret"
  }
}
```

## `rekey`

`gcy rekey [flags] CONFIG_FILE`

Re-encrypts all the secret values with specified arguments. By default, it will use the same provider for this operation, unless `--provider` is passed.

### Options:

- `--provider value`, `-p value`: The provider to encrypt values with (value is one of: [kms](pkg/crypto/kms), [gpg](pkg/crypto/gpg), [password](pkg/crypto/password))
- `--key value`: The AWS KMS key ARN to use. If no key is specified, `gcy` will prompt the user to select it from a list.
- `--public-key value`: A gpg public key's identity: a fingerprint like or email to use as a recipient to encrypt this file's data key. This option can be entered multiple times. If no recipients are specified, a list of available keys will be printed for the user to choose from.
- `--password value`: A password to use for encryption and decryption, also read from the environment variable `$CONFIG_PASSWORD`. To prevent your shell from remembering the password, start your command with a space: `[space]gcy ...`

```sh
gcy rekey config-up-there.yml
# Use the arrow keys to navigate: ↓ ↑ → ←  and / toggles search
# ? Select a key to continue:
#   ▸ arn:aws:kms:us-east-1:an-account:alias/an-alias
#     arn:aws:kms:us-east-1:an-account:alias/another-alias
#     ....
#     arn:aws:kms:us-east-1:an-account:alias/and-so-on
# ↓   arn:aws:kms:us-east-1:an-account:alias/and-so-forth
# ✔ arn:aws:kms:us-east-1:an-account:alias/another-alias
# INFO Re-encryption successful

# or specify the key if you know it
gcy rekey config-up-there.yml arn:aws:kms:an-aws-region:an-account:alias/an-alias
```

---

## Config files

Config files are [YAML](https://yaml.org/) files with nested objects representing a configuration tree. Storing encrypted values requires the presence of a `crypto` property with configuration for that provider, but the rest is up to you. `gcy` keeps keys ordered alphabetically doing a best-effort to keep comments in place. Here's a typical example of such a file, using the `kms` provider:

```yaml
crypto:
  provider: kms
  key: arn:aws:kms:an-aws-region:an-account:alias/an-alias

# and any arbitrary yaml afterwards
# Comments will be preserved by go-config-yourself, and all keys will be ordered
someKey: someValue

# Prefer nested objects over LONG_SCREAM_UNGROUPABLE_KEYNAMES
someObject:
  because: We use the right datatypes and
  wereNotCrazy: true
  verySecret:
    encrypted: true
    ciphertext: "...base64-encoded string"
    hash: "aSHA256hashOfTheSecret"
```

The recommended location for config files for projects is in the `config/` directory of a repository. A common usage pattern is to start with a `config/defaults.yml` file and then add override files for each environment the application will run in, like so:

```
- your-awesome-project/
  | - config/
      | - defaults.yml
      | - staging.yml
      | - production.yml
```

In the above scenario, you may store defaults or placeholders in `defaults.yml` with no encryption, while storing only the necessary secrets to override these placeholders in separate files. `staging.yml` and `production.yml` will only contain overrides to be applied on top of `defaults.yml`. `gcy` automatically adds placeholder values to `defaults.yml` after storing secrets in environment-specific files.

---

# Contributing to `go-config-yourself`

See [CONTRIBUTING.md](CONTRIBUTING.md) for details on how to start developing.
