# Ubuntu Resolute Raccoon

## Paketo Resolute base images

### What are these base images for?

Ideal for:

- Java apps and .NET Core apps
- Go apps that require some C libraries
- Node.js/Python/Ruby/etc. apps **without** many native extensions

### What's in the build and run images?

The build and the run images are based on Ubuntu Resolute Raccoon.

- To see the **list of all packages installed** in the build or run images for a given release, see the `ubuntu-resolute-{base image type}-{version}-{architecture}-receipt.cyclonedx.json` attached to each [release](https://github.com/paketo-buildpacks/ubuntu-resolute-base-images/releases). For a quick overview of the packages you can expect to find, see the [base images file descriptor](images/resolute-base-images/stack.toml).

## Paketo Resolute Tiny run image

### What is this base image for?

Ideal for:

- most Golang apps
- Java [GraalVM Native Images](https://www.graalvm.org/docs/reference-manual/native-image/)

### What's the tiny run image?

This image is based on Ubuntu Resolute Raccoon. The image does not include a Linux distribution.

- To see the **list of all packages installed** in the run image for a given release, see the `ubuntu-resolute-run-{version}-{architecture}-receipt.cyclonedx.json` attached to each [release](https://github.com/paketo-buildpacks/ubuntu-resolute-base-images/releases). For a quick overview of the packages you can expect to find, see the [tiny images file descriptor](images/resolute-tiny-images/stack.toml).

## Paketo Resolute Static Run Image

Image for statically-linked binaries on Ubuntu 26.04 LTS (Resolute Raccoon).

### What's the static run image?

This image is based on Ubuntu Resolute Raccoon. The image does not include a Linux distribution.

- To see the **list of all packages installed** in the run image for a given release, see the `ubuntu-resolute-run-{version}-{architecture}-receipt.cyclonedx.json` attached to each [release](https://github.com/paketo-buildpacks/ubuntu-resolute-base-images/releases). For a quick overview of the packages you can expect to find, see the [static images file descriptor](images/resolute-static-images/stack.toml).

## What is a base image?

See Cloud Native Buildpacks [base images documentation](https://buildpacks.io/docs/for-platform-operators/concepts/base-images/).

## How can I contribute?

Contribute changes to the base images via a Pull Request. Depending on the proposed changes, you may need to [submit an RFC](https://github.com/paketo-buildpacks/rfcs) first.

## How do I test the base images locally?

Run [`scripts/test.sh`](scripts/test.sh).

## How do I generate package receipts?

To generate a package receipt based on existing `build.oci` and `run.oci` archives, use [`scripts/receipts.sh`](scripts/receipts.sh).
