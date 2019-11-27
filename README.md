# go-config-yourself

[![Test status](https://github.com/blinkhealth/go-config-yourself/workflows/Tests/badge.svg)](https://github.com/blinkhealth/go-config-yourself/actions?query=branch%3Amaster+event%3Apush)
[![Coverage Status](https://coveralls.io/repos/github/blinkhealth/go-config-yourself/badge.svg?branch=master)](https://coveralls.io/github/blinkhealth/go-config-yourself?branch=master)

A secrets-management CLI tool and language-specific runtimes to deal with everyday application configuration right from your repository. **go-config-yourself** aims to simplify the management of secrets from your terminal and their use within application code. Configuration files are kept as human-readable as possible, so change management is easily achievable in any version-control system (like git).

`go-config-yourself` comes with the following cryptographic providers that do the work of encrypting and decrypting secrets, along with managing the keys for it:

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
brew tap blinkhealth/opensource-formulas
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
last_release="https://api.github.com/repos/blinkhealth/go-config-yourself/releases/latest"
version=$(curl --silent "$last_release" | awk -F'"' '/tag_name/{print $4}' )
curl -vO https://github.com/blinkhealth/go-config-yourself/releases/download/$version/gcy-linux-amd64.deb
apt install ./gcy-linux-amd64.deb
```

## Other Linux distros:

```sh
last_release="https://api.github.com/repos/blinkhealth/go-config-yourself/releases/latest"
version=$(curl --silent "$last_release" | awk -F'"' '/tag_name/{print $4}' )
curl -vO https://github.com/blinkhealth/go-config-yourself/releases/download/$version/gcy-linux-amd64.tar.gz
tar xfz gcy-linux-amd64.tar.gz
make install
```

---

# Usage

The command line interface is a program named `gcy` with four main commands: [`init`](#init), [`set`](#set), [`get`](#get) and [`rekey`](#rekey). Before diving in, here's how to get some help:

```sh
# See the included reference manual
gcy help
# or help for any command
gcy help init
# also a flag, so you don't even need to hit backspace
gcy init --provider kms --key some-long-arn --help
# see this exact page
man gcy
# or pages about providers
man gcy-password
# Show verbose output
gcy --verbose # ...rest of the command
```

## `init`

```sh
gcy init [options] CONFIG_FILE
```

Creates a YAML config file at `CONFIG_FILE`.

If needed, `gcy init` will query your provider for a list of keys to choose from when using the `aws` or `gpg` providers, and a password will be prompted for when using the `password` provider. `gcy init` will select the `aws` provider by default, and you can override it with the `--provider` flag.

See `gcy help config-file` for more information about `CONFIG_FILE`.

### Options:

- `--provider value`, `-p value`: The provider to encrypt values with (value is one of: [kms](pkg/crypto/kms), [gpg](pkg/crypto/gpg), [password](pkg/crypto/password))
- `--key value`: The AWS KMS key ARN to use.
- `--public-key value`: One gpg public key's identity (fingerprint or email) to use as a recipient to encrypt this file's data key. Pass multiple times for multiple recipients, or omit completely and `gcy` prompts you to select a key available to your gpg agent.
- `--password value`: A password to use for encryption and decryption. To prevent your shell from remembering the password in its history, start your command with a space: `[space]gcy ...`. Can be set via the environment variable: `CONFIG_PASSWORD`.
- `--skip-password-validation`: Skips password validation, potentially making encrypted secrets easier to crack.

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

```sh
gcy set [options] CONFIG_FILE KEYPATH
```

Stores a value at `KEYPATH`, encrypting it by default, and saves it to `CONFIG_FILE`.

`KEYPATH` is a dot-delimited path to values, see `gcy help keypath` for examples.

`gcy set` prompts for input, unless a value is provided via `stdin` or the `--input-file` flag. Values will be interpreted with golang’s default JSON parser before storage, so for example the string `“true”` will be stored as the boolean `true`. Due to existing AWS KMS service limitations, `gcy set` will read up to 4096 bytes before exiting with an error and closing its input.

A properly configured `crypto` property must exist `CONFIG_FILE` for encryption to succeed, `gcy set` will exit with a non-zero status code otherwise. See `gcy help config-file` for more information about `CONFIG_FILE`.

If a `defaults` or `default` file with the same extension as `CONFIG_FILE` exists in the same directory, `gcy set` will add a nil value for `KEYPATH` in said file.

### Options

- `-p|--plain-text`: Store the value as plain text with no encryption
- `-i|--input-file PATH`: Use the specified file path instead of prompting for input from `stdin`

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

```sh
gcy get CONFIG_FILE KEYPATH
```

Outputs the plain-text value for `KEYPATH` in `CONFIG_FILE`.

`KEYPATH` refers to a dot-delimited path to values, see `gcy help keypath` for examples.

If the value at `KEYPATH` is a dictionary or a list, it will be encoded as JSON, with all of the encrypted values within decrypted. If no value `KEYPATH` exists, `gcy get` will fail with exit code 2.

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

```sh
gcy rekey [options] CONFIG_FILE
```

Re-encrypts all the secret values with specified arguments in `CONFIG_FILE`.

By default, it will reuse the same provider for this operation, unless `--provider` is passed. If needed, `gcy rekey` will query your provider for a list of keys to choose from when using the `aws` or `gpg` providers, and a password will be prompted for when using the `password` provider.

### Options:

- `--provider value`, `-p value`: The provider to encrypt values with (value is one of: [kms](pkg/crypto/kms), [gpg](pkg/crypto/gpg), [password](pkg/crypto/password))
- `--key value`: The AWS KMS key ARN to use.
- `--public-key value`: One gpg public key's identity (fingerprint or email) to use as a recipient to encrypt this file's data key. Pass multiple times for multiple recipients, or omit completely and `gcy` prompts you to select a key available to your gpg agent.
- `--password value`: A password to use for encryption and decryption. To prevent your shell from remembering the password in its history, start your command with a space: `[space]gcy ...`. Can be set via the environment variable: `CONFIG_PASSWORD`.
- `--skip-password-validation`: Skips password validation, potentially making encrypted secrets easier to crack.

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

# Rekey between AWS profiles by temporarily rekeying with a password
 export CONFIG_PASSWORD="VERY-INSECURE-TEMPORARY-PASSWORD"
AWS_PROFILE=source gcy rekey --provider password config/file.yml
AWS_PROFILE=destination gcy rekey --provider kms config/file.yml
```

### Shell completion

`gcy` provides shell completion scripts for `bash` and `zsh`, which will be installed automatically by package managers. Shell completion is available for commands, options, `CONFIG_FILE` and `KEYPATH`.

---

## Config files

Config files are [YAML](https://yaml.org/) files with nested objects representing a configuration tree. Storing encrypted values requires the presence of a `crypto` property with configuration for that provider, but the rest is up to you. `gcy` keeps keys ordered alphabetically, doing its best-effort to keep comments in place. Here's a typical example of such a file, using the `kms` provider:

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
