# `password` provider

The Password provider encrypts values with a key derived from a user-defined password using [golang.org/x/crypto/scrypt](https://godoc.org/golang.org/x/crypto/scrypt). The values are encrypted using AES in GCM mode.


## A note on password security

Passwords are checked with [muesli/crunchy](github.com/muesli/crunchy) unless `--allow-insecure-passwords` is provided, which disables all checks. Regardless, you should **ensure a password with appropriate complexity is selected for your use case!**

## Example

```yaml
# CONFIG_PASSWORD=password
crypto:
  provider: password
  # This is a random key encrypted with the provided password
  key: azfUzNRdpdbYHb3AlML2asSo/gpDF5I4I7graqxvvD1VxXLsOitnrlVgLrRXk1YWX6sqFtNfnE7V0l9wMCmoYAV60qMO7IxQkjmAY3ObZa8RC5cW6P5M1b5UJjA=
zero:
  # These are encrypted with the random key described above
  ciphertext: i9gzOO+rpVk0XvZAbeDnMPdBsCA0oHbQ28oevBylmMdwFPCeR1qIPnnPIdx5rcfPfFhZHcMQeyFi5Q==
  encrypted: true
  hash: 6ac095169b05c043f89c9957f097f405d683bc15531824dca61bce544682d3b2
```

## Environment Variables

For all operations, you may set the `CONFIG_PASSWORD` environment variable and this provider will use that instead of prompting the user for the file's password. This is obviously **very insecure**, since the password will be available in your shell history!

```sh
# a space prefixes this `export` to prevent the shell from storing
# this password in its history file
export CONFIG_PASSWORD="a very long password that's hard to guess!"
go-config-yourself init --provider password file.yml
# INFO[0000] Creating config at file.yml
go-config-yourself set file.yml secret <<<"some secret"
go-config-yourself get file.yml secret
# some secret
```
