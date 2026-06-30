# hyperbdrctl

`hyperbdrctl` is the command-line client for HyperBDR / HyperMotion. It is written in Go and designed for daily operations, batch execution, and script integration.

The CLI talks to the platform through a unified HTTP API and supports both `dr` and `migration` scenes.

## Core Capabilities

- Manage local connection settings, language, and default output mode
- Query hosts, snapshots, tasks, licenses, and upgrade information
- Run host lifecycle operations such as register, sync, boot, deregister, and wait
- Manage host boot configurations and the top-level `boot-config apply` flow
- Manage source-side preparation, including Agent install metadata, Agentless sync proxies, and production sites
- Manage target-side cloud accounts, cloud resource discovery, cloud sync gateways, and object storage
- Use JSON output for troubleshooting and additional automation

## Typical Use Cases

- Integrate HyperBDR / HyperMotion operations into shell scripts, CI/CD jobs, or batch tasks
- Consume raw JSON output from custom automation systems
- Manage multiple environments and switch between `dr` and `migration`
- Complete common operational tasks without using the web UI

## Requirements

- Go 1.18 or later
- A reachable HyperBDR / HyperMotion endpoint, for example `https://<host>:10443`
- Valid platform credentials

## Build

Run in the current module directory:

```sh
go build -o hyperbdrctl ./cmd/hyperbdrctl
```

To generate a smaller binary for distribution:

```sh
go build -trimpath -ldflags="-s -w" -o hyperbdrctl ./cmd/hyperbdrctl
```

On Windows, output `hyperbdrctl.exe` if needed.

## Quick Start

### 1. Save Local Default Configuration

```sh
hyperbdrctl config set \
  --host https://example:10443 \
  --username admin \
  --password <password> \
  --scene migration \
  --lang zh_cn
```

If you are in a test environment and must skip TLS verification, add:

```sh
--insecure
```

`config set` validates the final merged credentials before saving. If authentication fails, nothing is written to the local config.

### 2. Run Basic Queries

```sh
hyperbdrctl host list --page 1 --page-size 10
hyperbdrctl cloud-account list
hyperbdrctl oss list
```

### 3. Use JSON Output in Scripts

```sh
hyperbdrctl host list --page 1 --page-size 10 --output json
```

The default output mode is table. When `--output json` is used, the CLI returns raw API field names for direct script consumption.

## Configuration Sources

The CLI supports these configuration sources:

1. Command-line flags
2. Environment variables
3. Local config file
4. Defaults

The fixed priority order is:

```text
command-line flags > environment variables > config file > defaults
```

Common environment variables:

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

Notes:

- `--lang` controls both CLI text and the HTTP `X-LANG` header
- Supported languages are `en` and `zh_cn`
- The default output mode is `table`
- `config get` hides saved passwords by default
- TLS verification is enabled by default, and `--insecure` should only be used explicitly in test environments

## Common Workflows

### Hosts and Boot Configuration

```sh
hyperbdrctl host list
hyperbdrctl host detail --id <host_id>
hyperbdrctl boot-config get --id <host_id>
hyperbdrctl boot-config apply --id <host_id> --file ./boot-config.json
hyperbdrctl host wait --id <host_id>
```

Notes:

- Stable single-host boot configuration management uses `boot-config get` and `boot-config apply`

### Source Preparation

Inspect source-side install metadata:

```sh
hyperbdrctl agent install
hyperbdrctl sync-proxy install
hyperbdrctl sync-proxy list
```

Create an Agentless production site:

```sh
hyperbdrctl production-site create \
  --type vmware \
  --synch-node-id <node_id> \
  --auth-url https://vcenter.example:443 \
  --auth-key <username> \
  --auth-cert <password>
```

After creation, validate the binding result:

```sh
hyperbdrctl production-site list --type vmware --binding-status binding
```

### Target Cloud Accounts

```sh
hyperbdrctl cloud-account list
hyperbdrctl cloud-account create --help
hyperbdrctl cloud-resource fetch --help
```

`cloud-account create --help` adapts its guidance from the selected `--cloud-type` and `--storage-type`. Use `cloud-resource fetch --help` to resolve cloud-side resource IDs before creating accounts, gateways, or boot configurations.

### Cloud Sync Gateways

```sh
hyperbdrctl cloud-sync-gateway list
hyperbdrctl cloud-resource fetch --cloud-account-id <account_id> --output json
hyperbdrctl cloud-sync-gateway create --help
hyperbdrctl cloud-sync-gateway wait --id <storage_id>
```

Recommended sequence:

1. Confirm or create the target cloud account
2. Query the required creation resources
3. Run the provider-specific gateway creation command
4. Use `wait` to confirm the final state

### Object Storage

```sh
hyperbdrctl oss list
hyperbdrctl oss detail --id <storage_id>
hyperbdrctl oss catalog
hyperbdrctl oss buckets --help
hyperbdrctl oss create --help
```

### License Management

```sh
hyperbdrctl license list
hyperbdrctl license reg-code
hyperbdrctl license activate --kkty <reg_code> --ddty <activation_code>
```

## Command Overview

```text
hyperbdrctl
|- boot-config
|  `- get / apply
|- completion
|- config
|  `- get / set
|- host
|  `- list / detail / snapshots / register / sync / boot / clean / deregister / wait
|- license
|  `- list / reg-code / activate
|- production-site
|  `- list / detail / create / delete / vm-list
|- agent
|  `- install
|- sync-proxy
|  `- install / list / delete
|- cloud-account
|  `- list / detail / create / wait / delete
|- cloud-resource
|  `- fetch
|- cloud-sync-gateway
|  `- list / detail / create / wait / delete
`- oss
   `- list / detail / catalog / buckets / create / wait / delete
```

For full flags and examples for a specific command, run:

```sh
hyperbdrctl --help
hyperbdrctl config set --help
hyperbdrctl host --help
hyperbdrctl production-site create --help
hyperbdrctl cloud-account create --help
hyperbdrctl cloud-sync-gateway create --help
```

## Release and Integration Notes

- Prefer `--output json` for automation scenarios
- Set `--scene` explicitly when switching across environments
- Do not enable `--insecure` by default in production
- To minimize manual input, prefer environment variables or a pre-populated config file
- Add `--debug` when request-level troubleshooting is needed

## Development Verification

When code changes are made, run the standard verification commands in this module directory:

```sh
gofmt -w cmd internal
go test ./...
go build -o hyperbdrctl ./cmd/hyperbdrctl
```
