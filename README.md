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

### Requirements

- A reachable HyperBDR / HyperMotion endpoint, for example `https://<host>:10443`
- Valid platform credentials
- A released binary

### Configure an endpoint

`config set` validates merged credentials before saving. Failed authentication does not overwrite the existing configuration.

```sh
hyperbdrctl config set --host https://example:10443 --username admin --password <password> --scene migration --lang zh_cn
hyperbdrctl config get
```

For an explicitly trusted test environment with an untrusted certificate:

```sh
hyperbdrctl config set --insecure
```

### Query and operate

```sh
hyperbdrctl host list --page 1 --page-size 10
hyperbdrctl cloud-account list
hyperbdrctl oss list
hyperbdrctl host wait --id <host_id>
```

Use `--help` at any level to discover supported flags and provider-specific examples:

```sh
hyperbdrctl --help
hyperbdrctl host --help
hyperbdrctl cloud-account create --help
```

### Machine-readable output for agents

Use JSON for tool calls, planning, validation, and follow-up actions. API field names are preserved.

```sh
hyperbdrctl --output json host list --page 1 --page-size 10
hyperbdrctl --output json cloud-resource fetch --cloud-account-id <account_id>
```

The default output is a table. `--vertical` helps inspect one record and `--debug` enables request-level troubleshooting.

### Configuration sources

Values are resolved in this order:

```text
command-line flags > environment variables > local config file > defaults
```

Supported environment variables:

```text
HYPERBDR_HOST
HYPERBDR_USERNAME
HYPERBDR_PASSWORD
HYPERBDR_SCENE
HYPERBDR_LANG
HYPERBDR_OUTPUT
HYPERBDR_INSECURE
HYPERBDR_DEBUG
```

`--lang` controls CLI messages and the HTTP `X-LANG` header. Supported values are `en` and `zh_cn`.
TLS verification is enabled by default; use `--insecure` only in a test environment.

### Common workflows

```sh
# Hosts and boot configuration
hyperbdrctl host list
hyperbdrctl host detail --id <host_id>
hyperbdrctl boot-config get --id <host_id>
hyperbdrctl boot-config apply --id <host_id> --file ./boot-config.json
hyperbdrctl host wait --id <host_id>

# Source preparation
hyperbdrctl agent install
hyperbdrctl sync-proxy install
hyperbdrctl sync-proxy list
hyperbdrctl production-site create --type vmware --synch-node-id <node_id> --auth-url https://vcenter.example:443 --auth-key <username> --auth-cert <password>

# Target cloud resources
hyperbdrctl cloud-account list
hyperbdrctl cloud-resource fetch --cloud-account-id <account_id> --output json
hyperbdrctl cloud-sync-gateway create --help
hyperbdrctl cloud-sync-gateway wait --id <storage_id>
hyperbdrctl oss catalog

# Licenses
hyperbdrctl license list
hyperbdrctl license reg-code
hyperbdrctl license activate --kkty <reg_code> --ddty <activation_code>
```

### Command map

```text
hyperbdrctl
├── config                 get / set
├── completion             generate shell completion scripts
├── host                   list / detail / snapshots / register / sync / boot / clean / deregister / wait
├── boot-config            get / apply
├── production-site        list / detail / create / delete / vm-list
├── agent                  install
├── sync-proxy             install / list / delete
├── cloud-account          list / detail / create / wait / delete
├── cloud-resource         fetch
├── cloud-sync-gateway     list / detail / create / wait / delete
├── oss                    list / detail / catalog / buckets / create / wait / delete
└── license                list / reg-code / activate
```

## Features

- **Built for AI agents** — predictable subcommands, explicit flags, machine-readable JSON, and asynchronous `wait` operations make commands easy to plan, execute, and verify.
- **Automation-first configuration** — flags, environment variables, or a local config file; no browser session or interactive console is required.
- **One API surface for DR and migration** — switch `--scene` without changing your automation model.
- **End-to-end operational coverage** — hosts, snapshots, boot configuration, source preparation, cloud discovery, gateways, object storage, and licenses.
- **Safe by default** — TLS verification is on, secrets are hidden by `config get`, and configuration is saved only after credentials are validated.
- **Operator-friendly** — table/vertical output, bilingual messages (`en` / `zh_cn`), and command-local help.
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
