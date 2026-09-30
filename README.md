<div align="center">

# hyperbdrctl

**An agent-ready CLI for operating HyperBDR and HyperMotion.**

Give an AI agent a predictable command surface for disaster recovery and migration workflows — with
structured output, explicit flags, safe defaults, and no web-console automation required.

[English](README.md) | [中文](README.zh-CN.md)

[![Go](https://img.shields.io/badge/go-1.18%2B-00ADD8?logo=go)](go.mod)
[![License](https://img.shields.io/badge/license-Apache%202.0-blue.svg)](LICENSE)
[![Status](https://img.shields.io/badge/status-active%20development-orange.svg)](#project-status)

[Repository](https://github.com/oneprolabs/hyperbdrctl) · [Issues](https://github.com/oneprolabs/hyperbdrctl/issues) · [Releases](https://github.com/oneprolabs/hyperbdrctl/releases)

</div>

`hyperbdrctl` is a Go command-line client for HyperBDR / HyperMotion. It is primarily built for AI agents
and automation systems that inspect state, execute operations, wait for asynchronous tasks, and consume
results without browser UI scraping. It supports both `dr` and `migration` scenes.

## Usage

### Configure the environment

Before using the CLI, prepare a reachable HyperBDR / HyperMotion endpoint, valid platform credentials, and a released binary.

Configure the endpoint and credentials:

```sh
hyperbdrctl config set \
  --host https://<host>:10443 \
  --username <username> \
  --password <password> \
  --scene dr \
  --lang zh_cn \
  --timezone Asia/Shanghai

hyperbdrctl config get
```

`config set` validates the credentials before saving the configuration. Enable this only when a test environment uses an untrusted certificate:

```sh
hyperbdrctl config set --insecure
```

Configuration can also be provided through environment variables. Values are resolved in this order:

```text
command-line flags > environment variables > local config file > defaults
```

Common environment variables:

```text
HYPERBDR_HOST
HYPERBDR_USERNAME
HYPERBDR_PASSWORD
HYPERBDR_SCENE
HYPERBDR_LANG
HYPERBDR_OUTPUT
HYPERBDR_TIMEZONE
HYPERBDR_INSECURE
HYPERBDR_DEBUG
```

`--lang` supports `en`, `zh_cn`, and `ja`. `--timezone` accepts `Local` or an IANA time zone such as `Asia/Shanghai`. `Local` resolves to the operating system's concrete time zone identifier, saved explicitly in the config file. The setting affects table and vertical output; JSON output keeps API timestamps unchanged. TLS verification is enabled by default.

### Workflow and command help

The basic HyperBDR workflow is:

```text
Source preparation → Target preparation → DR configuration → Sync and boot → Resource cleanup
```

Use the following commands to view detailed flags and examples for each stage:

```sh
# Source preparation: production sites, sync proxies, and VM registration
hyperbdrctl production-site --help
hyperbdrctl sync-proxy --help
hyperbdrctl host register --help

# Target preparation: cloud accounts, cloud resources, object storage, and sync gateways
hyperbdrctl cloud-account --help
hyperbdrctl cloud-resource --help
hyperbdrctl oss --help
hyperbdrctl cloud-sync-gateway --help

# DR configuration: hosts, snapshots, and boot configuration
hyperbdrctl host --help
hyperbdrctl host snapshots --help
hyperbdrctl boot-config --help

# Sync and boot
hyperbdrctl host sync --help
hyperbdrctl host boot --help
hyperbdrctl host wait --help

# Resource cleanup
hyperbdrctl host clean --help
hyperbdrctl host deregister --help
hyperbdrctl cloud-sync-gateway delete --help
hyperbdrctl oss delete --help
```

For any command, use `--help` to see its available subcommands, flags, and examples.

`sync`, `boot`, and `clean` are asynchronous operations. After starting one, use `host wait` to confirm the result:

```sh
hyperbdrctl host sync --id <host_id>
hyperbdrctl host wait --id <host_id> --operation sync
```

For automation, use JSON output. For troubleshooting, use `--vertical` and `--debug`:

```sh
hyperbdrctl --output json host list
```

## Documentation

See the [CLI style documentation](docs/README.md) for multilingual conventions covering help
pages and terminal output:

- [CLI Help Style](docs/cli-help-style.en.md) · [中文](docs/cli-help-style.zh.md)
- [CLI Output Style](docs/cli-output-style.en.md) · [中文](docs/cli-output-style.zh.md)

## Features

- **Built for AI agents** — predictable subcommands, explicit flags, machine-readable JSON, and asynchronous `wait` operations make commands easy to plan, execute, and verify.
- **Automation-first configuration** — flags, environment variables, or a local config file; no browser session or interactive console is required.
- **One API surface for DR and migration** — switch `--scene` without changing your automation model.
- **End-to-end operational coverage** — hosts, snapshots, boot configuration, source preparation, cloud discovery, gateways, object storage, and licenses.
- **Safe by default** — TLS verification is on, secrets are hidden by `config get`, and configuration is saved only after credentials are validated.
- **Operator-friendly** — table/vertical output, localized messages (`en` / `zh_cn` / `ja`), and command-local help.
- **Portable** — one Go binary with no runtime dependency beyond network access to the platform endpoint.

## Build and development

### Build from source

Building from source requires Go 1.18 or later.

```sh
go build -o hyperbdrctl ./cmd/hyperbdrctl
go build -trimpath -ldflags="-s -w" -o hyperbdrctl ./cmd/hyperbdrctl
```

On Windows, use `hyperbdrctl.exe` when needed.

### Release build and version metadata

```sh
VERSION=v1.2.3
go build -trimpath -buildvcs=true -ldflags "-s -w -X hyperbdr-client/internal/version.Version=${VERSION}" -o hyperbdrctl ./cmd/hyperbdrctl
./hyperbdrctl --version
go version -m ./hyperbdrctl
```

### Development checks

```sh
gofmt -w cmd internal
go test ./...
go build -o hyperbdrctl ./cmd/hyperbdrctl
```

When changing a command, update its localized help text and tests. Keep agent-facing JSON stable and document breaking changes.

## Project status

`hyperbdrctl` is in active development. The command surface and provider-specific request fields may evolve
with HyperBDR / HyperMotion APIs. Pin a known binary version in production automation and validate upgrades first.

## Contributing

Issues and pull requests are welcome. Include tests for behavior changes, run the development checks, and describe the API or automation impact.

## License

This project is licensed under the [Apache License 2.0](LICENSE).
