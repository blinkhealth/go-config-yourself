# `kms` provider

The KMS provider uses the [AWS KMS](https://aws.amazon.com/kms/) service to encrypt every secret value.

## Example

```yaml
crypto:
  provider: kms
  key: arn:aws:kms:us-east-1:000000000000:key/00000000-0000-0000-0000-000000000000
zero:
  ciphertext: AKMSTHiNG0123456789AAAA...0XyZ
  encrypted: true
  hash: 6b86b273ff34fce19d6b804eff5a3f5747ada4eaa22f1d49c01e52ddb7875b4b
```

## Environment variables

`go-config-yourself` strives to behave like [any other AWS SDK client](https://docs.aws.amazon.com/sdk-for-go/v1/developer-guide/configuring-sdk.html#specifying-credentials).

The following `AWS_*` environment variables are implemented:

- `AWS_PROFILE`: if set, go-config-yourself will use the profile configuration from `~/.aws/config` to prompt for MFA tokens when required or assuming a role.

## Known Issues

- Since not all regions are [enabled by default](https://docs.aws.amazon.com/general/latest/gr/rande.html), and there is currently no API endpoint to list the available regions for the user's AWS account, `gcy (init|rekey)` will only warn on `UnrecognizedClientException` and continue querying for keys in all regions.
