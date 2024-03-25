# Schulze voting method on Ethereum Attestation Service - CLI

[![Go](https://github.com/janos/schulzeoneas/workflows/Go/badge.svg)](https://github.com/janos/schulzeoneas/actions)
[![NewReleases](https://newreleases.io/badge.svg)](https://newreleases.io/github/janos/schulzeoneas)

A command line client for managing votings using the Schulze voting method on Ethereum Attestation Service.

# Installation

SchulzeOnEAS binaries have no external dependencies and can be just copied and executed locally.

Binary downloads can be found on the [Releases page](https://github.com/janos/schulzeoneas/releases/latest).

To install on macOS:

```sh
wget https://github.com/janos/schulzeoneas/releases/latest/download/schulzeoneas-darwin-amd64 -O /usr/local/bin/schulzeoneas
chmod +x /usr/local/bin/schulzeoneas
```

You may need additional privileges to write to `/usr/local/bin`, but the file can be saved at any location that you want.

Supported operating systems and architectures:

- macOS ARM 64bit `darwin-arm64`
- macOS 64bit `darwin-amd64`
- Linux 64bit `linux-amd64`
- Linux 32bit `linux-386`
- Linux ARM 64bit `linux-arm64`
- Linux ARM 32bit `linux-armv6`
- Windows 64bit `windows-amd64`
- Windows ARM 64bit `windows-arm64`
- Windows 32bit `windows-386`

Deb and RPM packages are also built.

This tool is implemented using the Go programming language and can be also installed by issuing a `go get` command:

```sh
go get -u resenje.org/schulzeoneas/cmd/schulzeoneas
```

# Versioning

Each version is tagged and the version is updated accordingly in `version.go` file.

# Contributing

We love pull requests! Please see the [contribution guidelines](CONTRIBUTING.md).

# License

This application is distributed under the BSD-style license found in the [LICENSE](LICENSE) file.
