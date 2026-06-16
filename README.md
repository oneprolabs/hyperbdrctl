# hyperbdrctl

`hyperbdrctl` is the command-line client for HyperBDR / HyperMotion. It is written in Go and designed for daily operations, batch execution, and script integration.

The CLI talks to the platform through a unified HTTP API and supports both `dr` and `migration` scenes.

## Core Capabilities

- Manage local connection settings, language, and default output mode
- Query hosts, snapshots, tasks, licenses, and upgrade information
- Run host lifecycle operations such as register, sync, boot, deregister, and wait
- Manage host boot configurations and the top-level `boot-config apply` flow
- Manage source connections, including Agent / Agentless install metadata and Agentless source creation
- Manage target-side cloud accounts, cloud sync gateways, and object storage
- Use `api request` to call authenticated APIs directly for troubleshooting and additional automation

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
hyperbdrctl target account list
hyperbdrctl tasks list
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
hyperbdrctl host boot-config get --id <host_id>
hyperbdrctl boot-config apply --id <host_id> --file ./boot-config.json
hyperbdrctl host wait --id <host_id>
```

Notes:

- Stable host-level boot configuration management uses `host boot-config`
- The independent single-host override apply flow uses top-level `boot-config apply`

### Source Connections

Inspect source-side install metadata:

```sh
hyperbdrctl source agent-install
hyperbdrctl source agentless-install
hyperbdrctl source sync-nodes
```

Create an Agentless source connection:

```sh
hyperbdrctl source create \
  --type vmware \
  --synch-node-id <node_id> \
  --auth-url https://vcenter.example:443 \
  --auth-key <username> \
  --auth-cert <password>
```

After creation, validate the binding result:

```sh
hyperbdrctl source list --type vmware --binding-status binding
```

### Target Cloud Accounts

```sh
hyperbdrctl target supports
hyperbdrctl target account list
hyperbdrctl target account fetch-block-resources --help
hyperbdrctl target account create-block aliyun --help
hyperbdrctl target account create-oss openstack --help
```

Write operations are exposed through provider-specific subcommands. The supported providers should be confirmed with `target supports` and the corresponding `create-* --help` output at runtime.

### Cloud Sync Gateways

```sh
hyperbdrctl target cloud-sync-gateway list
hyperbdrctl target cloud-sync-gateway resources --cloud-account-id <account_id> --output json
hyperbdrctl target cloud-sync-gateway create aliyun --help
hyperbdrctl target cloud-sync-gateway wait --id <storage_id>
```

Recommended sequence:

1. Confirm or create the target cloud account
2. Query the required creation resources
3. Run the provider-specific gateway creation command
4. Use `wait` to confirm the final state

### Object Storage

```sh
hyperbdrctl target oss list
hyperbdrctl target oss detail --id <storage_id>
hyperbdrctl target oss buckets --help
hyperbdrctl target oss create --help
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
|- api
|- batch-boot-config
|- boot-config
|  `- apply
|- boot-config-wizard
|- completion
|- config
|- host
|  |- list / detail / snapshots / register / sync / boot / cleanup-validation-host / deregister / wait
|  `- boot-config
|- license
|- source
|  |- list / detail / vms / agent-install / agentless-install / sync-nodes / create
|- target
|  |- supports
|  |- account
|  |- cloud-sync-gateway
|  `- oss
|- tasks
`- upgrade
```

For full flags and examples for a specific command, run:

```sh
hyperbdrctl --help
hyperbdrctl config set --help
hyperbdrctl host --help
hyperbdrctl source create --help
hyperbdrctl target account create-block --help
hyperbdrctl target cloud-sync-gateway create --help
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
go build -o ../hyperbdrctl ./cmd/hyperbdrctl
```
