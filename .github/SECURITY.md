# Security Policy

## Reporting a Vulnerability

Please reach us at security@blinkhealth.com if you believe you've found a security issue in `go-config-yourself`. We will assess the risk and make a fix available before we add a bug report to the Github repository.

Thank you for improving the security of `go-config-yourself`. We appreciate your efforts and responsible disclosure; we're committed to make every effort in acknowledging your contributions.

## Disclosure Policy

When the security team receives a security bug report, they will assign it to a primary handler. This person will coordinate the fix and release process, involving the following steps:

1. Confirm the problem and determine the affected versions
2. Audit code to find any potential similar problems
3. Prepare fixes for all releases still under maintenance
4. Release a GitHub security advisory

## Comments on this Policy
If you have suggestions on how this process could be improved please submit a pull request.

## Source Code Audit

`go-config-yourself` was audited in November 2019 by a third party to identify security issues, including:

- Reviewing of source code using static and dynamic analysis tools
- Tests for memory management issues, command injection, and data injection/validation
- Reviewing of source code for logging issues and race conditions

The resulting recommendations have been addressed before the initial stable release. The resulting binaries and source code have not been audited since. Therefore, any current or future binaries or source code available on this repository are provided as-is, without any kind of warranty as mentioned in this project's [LICENSE](https://github.com/blinkhealth/go-config-yourself/LICENSE).
