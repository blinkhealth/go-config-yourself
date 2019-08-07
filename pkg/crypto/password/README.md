# `password` provider

The Password provider encrypts values with a key derived from a user-defined password using [golang.org/x/crypto/scrypt](https://godoc.org/golang.org/x/crypto/scrypt). The values are encrypted using AES in GCM mode.

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

# Environment Variables

For all operations, you may set the `CONFIG_PASSWORD` environment variable and this provider will use that instead of prompting the user for the file's password.

```sh
export CONFIG_PASSWORD="password"
go-config-yourself init --provider password file.yml
# INFO[0000] Creating config at file.yml
go-config-yourself set file.yml secret <<<"some secret"
go-config-yourself get file.yml secret
# some secret
```
