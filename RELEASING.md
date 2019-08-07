# Releasing

## Stable builds

Releasing stable builds takes a few short steps:

1. Create a Github release with a changelog.

Then, automation will:

1. A CircleCI job will be triggered and built artifacts deployed to Github.
2. A pull request in the [Homebrew formulae repo](https://github.com/blinkhealth/opensource-formulas/pulls) will be opened for this version.
3. A debian package will be published to the [stable ppa](https://launchpad.net/~blinkhealth/+archive/ubuntu/stable).
4. [Documentation](https://blinkhealth.github.com/go-config-yourself) will be published.

## Unstable builds

Every merge to the mainline branch triggers the creation of a pre-release in Github.

1. A CircleCI job will be triggered and built artifacts deployed to Github.
2. A debian package will be published to the [unstable ppa](https://launchpad.net/~blinkhealth/+archive/ubuntu/unstable)
