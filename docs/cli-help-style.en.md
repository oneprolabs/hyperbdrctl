# CLI Help Style Guide

[English](cli-help-style.en.md) | [中文](cli-help-style.zh.md)

This guide defines the layout, wording, and examples used by `hyperbdrctl` CLI help pages.

## Example

### Resource creation command

```text
Create an Alibaba Cloud block storage account

Usage: hyperbdrctl cloud-account create --cloud-type aliyun --storage-type block [flags]

Flags:
      --access-key-id string       Cloud provider Access Key ID (required)
      --access-key-secret string   Cloud provider Access Key Secret (required)
      --auth-region-id string      Authentication region ID (required)
      --set stringArray            Override metadata path with path=value; repeatable
      --set-json stringArray       Override metadata path with path=<json>; repeatable
      --preview-request            Print the request body without sending it
      --debug                      Print request debugging logs
      --lang string                Display language, allowed values en / zh_cn / ja, default en
  -o, --output string              Output format, allowed values table / json, default table
  -h, --help                       Show help information

Usage Notes:
  Create an Alibaba Cloud block storage account.

  Parameter Sources:
    --access-key-id string
      Use the Alibaba Cloud Access Key ID.

    --access-key-secret string
      Use the Alibaba Cloud Access Key Secret.

    --auth-region-id string
      Use the Alibaba Cloud authentication region ID.

  Resource discovery:
    First list available authentication regions:
      hyperbdrctl cloud-resource fetch --cloud-type aliyun --storage-type block \
        --access-key-id <ak> \
        --access-key-secret <sk> \
        --fetch-res regions \
        --output json

  Minimal creation command:
    hyperbdrctl cloud-account create --cloud-type aliyun --storage-type block \
      --access-key-id <ak> \
      --access-key-secret <sk> \
      --auth-region-id <auth_region_id>

  After creation, wait for the task to finish:
    hyperbdrctl cloud-account wait --id <account_id>

  Then inspect the account:
    hyperbdrctl cloud-account detail --id <account_id>
    hyperbdrctl cloud-account list --storage-type block

  Add optional fields as needed:
    --set path=value
    --set-json path=<json>

  Override precedence is fixed:
    --set-json < --set < explicit creation flags

  Optional dynamic parameter:
    --account-name <name>
      Account name

  Add the following flag to inspect the final request body first:
    --preview-request
```

## Scope

This guide applies to:

- New CLI help pages
- Changes to existing CLI help pages
- Help rendering logic, section order, and visibility changes
- Help tests, examples, and operational guidance

## Design principles

### Page structure

Leaf commands use four sections:

1. Description
2. Usage
3. Flags
4. Usage Notes

Command-group entry points use five sections when stable subcommands need to be shown:

1. Description
2. Usage
3. Flags
4. Commands
5. Usage Notes

Rules:

- Start the description with the command's purpose.
- Place `Usage` and `Flags` immediately after the description.
- Do not add a `Commands` section to a leaf command.
- Use `Commands` only when it provides direct navigation value for a stable subcommand set.
- Keep the structure of a command's help output stable over time.

### Flags

- Keep exactly one `Flags` section.
- Order flags as required flags, command-specific flags, then global flags.
- Mark required flags with `(required)` in the description.
- Keep short descriptions focused on what the flag is.
- Put parameter sources, lookup steps, and auto-completion behavior in `Usage Notes`.
- Write fixed choices as `allowed values a / b`.
- Write fixed fallbacks as `default x`.
- Do not add a colon after `allowed values` or `default`.
- Keep global flags at the end.

Example:

```text
      --name string     Resource name (required)
      --type string     Resource type, allowed values a / b, default a
      --debug           Print request debugging logs
      --lang string     Display language, allowed values en / zh_cn / ja, default en
  -o, --output string   Output format, allowed values table / json, default table
  -h, --help            Show help information
```

### Usage Notes

Organize notes in the operator's execution order:

1. Preparation
2. Discovery
3. Minimal command
4. Optional extensions
5. Preview or verification

Use complete `hyperbdrctl ...` commands. Explain important parameter sources and lookup
steps here, without exposing implementation details unless they affect operation.

### Reuse existing logic

Before changing help style, inspect the existing renderer, command constructors, and tests.
Prefer the shared help entry points and the smallest common change. If several commands have
the same issue, fix the shared logic instead of adding repeated command-specific code.

## Wording and layout

### Stable terminology

| Meaning | zh_cn | en | ja |
| --- | --- | --- | --- |
| Usage section | `用法` | `Usage` | `使い方` |
| Flags section | `参数` | `Flags` | `フラグ` |
| Commands section | `命令` | `Commands` | `コマンド` |
| Usage notes section | `使用说明` | `Usage Notes` | `使用上の注意` |
| Parameter sources | `参数来源` | `Parameter Sources` | `パラメータの出所` |
| Required marker | `（必须）` | `(required)` | `(必須)` |
| Allowed values | `可选值 a / b` | `allowed values a / b` | `使用可能な値 a / b` |
| Default value | `默认值 x` | `default x` | `デフォルト x` |
| Help description | `显示帮助信息` | `Show help information` | `ヘルプ情報を表示` |

Chinese help uses full-width Chinese punctuation; English help uses English punctuation.
All languages must retain the same section order, flag order, and example structure.

### Multiline output headings

For multiple peer sections in terminal output, use `== Title ==` followed by a separator:

```text
== Block Storage Clouds ==
-------------------------
Provider       Name
aliyun         Alibaba Cloud (Recommended, SDK v2.0)
```

### Indentation and long commands

- Keep indentation stable across sections and continuation lines.
- Break commands that are hard to read at normal terminal widths.
- Use a trailing `\\` for shell continuations.
- Indent continuation flags by two spaces relative to the command.

### Wording

Use formal, direct, operator-oriented language. Keep short descriptions to one sentence.
Avoid conversational openings such as “This command is” and avoid placing background or
implementation history in the short description.

## Command categories

Representative commands must come from the current public `hyperbdrctl --help` output.

| Category | Representative commands | Description focus | Usage-notes focus |
| --- | --- | --- | --- |
| Mutating operations | `cloud-account create`, `oss create`, `boot-config apply`, `host sync`, `host boot`, `license activate` | What is created, applied, activated, deleted, or started | Preparation, minimal command, extensions, preview, verification |
| List queries | `host list`, `cloud-account list`, `production-site list`, `license list` | Which list is queried | Minimal query, filters, and using results as later input |
| Detail queries | `host detail`, `cloud-account detail`, `oss detail`, `boot-config get` | Which object or configuration is inspected | ID source, minimal query, and next operation |
| Informational output | `agent install`, `sync-proxy install`, `license reg-code`, `completion` | Which instructions or auxiliary data are printed | How to use the output and what to do next |
| Discovery helpers | `cloud-resource fetch`, `oss catalog`, `oss buckets`, `host snapshots` | Which follow-up input is collected | What to query first and where the result is used |
