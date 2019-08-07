# `gpg` provider

The GPG provider use an existing GPG agent to encrypt and decrypt a AES256 data key. Each individual secret is then encrypted using that key.

`go-config-yourself` interacts directly with the GPG agent using the [gpgme library](https://gnupg.org/software/gpgme/index.html), and thus, requires a GPG agent present and available.

## Example

```yaml
crypto:
  provider: gpg
  recipients:
    - test-software@blinkhealth.com
  key: |
    -----BEGIN PGP MESSAGE-----

    hQEMA1XMevbqrQGfAQf+N40/YoVX9zTtlXGw2GDZmq4rbpv8DTDR0xCO1RUwyA23
    Y9+1t9N4wLel9FMaFx3oCNvFkWYcjKCvTmZrZ3jn1WEbuGUqnvlCWN....UdL41=
    -----END PGP MESSAGE-----
zero:
  ciphertext: 0x6IenU6EEErLVMC65tFWnPItFnEI8E/6i3GiOc=
  encrypted: true
  hash: 6b86b273ff34fce19d6b804eff5a3f5747ada4eaa22f1d49c01e52ddb7875b4b
```

# Environment variables

When using the `gpg` provider, these environment variables can affect the way `go-config-yourself` interacts with the GPG agent.

- `GNUPGHOME`: This path will be used as the [GPG homedir](https://www.gnupg.org/gph/en/manual/r1616.html). Usually, a folder with at least `pubring.gpg`, `secring.gpg` and `trustdb.gpg`.
