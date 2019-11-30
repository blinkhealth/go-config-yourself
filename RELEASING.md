# Releasing

## Stable builds

Releasing stable builds takes a few short steps:

1. Create a Github release with a changelog.
2. The [Release action](/.github/workflows/release.yml) job will be triggered and built artifacts deployed to Github.
3. A pull request in the [Homebrew formulae repo](https://github.com/blinkhealth/homebrew-opensource-formulas/pulls) will be opened for this version.
4. [Documentation](https://blinkhealth.github.com/go-config-yourself) will be published.

## Unstable builds

Every merge to the mainline branch triggers the creation of a pre-release in Github.

1. The [Unstable Release action](/.github/workflows/release-unstable.yml) job will be triggered and built artifacts deployed to Github.
