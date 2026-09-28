# CLI Output Style Guide

[English](cli-output-style.en.md) | [中文](cli-output-style.zh.md)

This guide defines terminal presentation and wording for non-help `hyperbdrctl` output.

## Scope

It covers default tables, JSON, vertical output, key-value output, section headings,
resource candidates, request previews, errors, and debug logs.

## Principles

- Optimize default output for clear, stable, copyable operator workflows.
- Keep JSON structured and machine-readable; do not localize field names.
- Show key fields in tables instead of dumping the raw response.
- Keep headings, columns, ordering, and empty-value behavior consistent for like resources.
- Use formal, direct Chinese wording and concise, direct English wording.

## Output modes

| Mode | Trigger | Purpose |
| --- | --- | --- |
| Table | Default or `--output table` | Human-readable key fields |
| JSON | `--output json` | Scripts, automation, and troubleshooting |
| Vertical | `--vertical` or `-G` | Expand wide list records one field per line |

JSON must contain no table headers, section headings, explanatory text, or other decoration.
Vertical output changes presentation only; it does not change field meaning.

## JSON

Use two-space indentation and emit valid JSON only:

```json
{
  "id": "host-1",
  "name": "app-01",
  "status": "running"
}
```

Field names remain structured API names. Do not replace JSON keys with localized names or
Chinese punctuation.

## Tables

```text
ID      Name    Status
host-1  app-01  running
host-2  db-01   stopped
```

The first line is the header; subsequent lines are data. Separate columns with two spaces,
align by display width, treat CJK characters as wide characters, show empty values as empty
strings, booleans as `true` or `false`, and compact complex values as JSON only when needed.

Show only fields important for the current command and likely to be used next. Put ID, name,
status, type, and time fields near the front. Hide an all-empty column. Keep the header when
the result is empty, and use consistent localized names for the same resource type.

Common ordering:

```text
ID  Name  Type  Status  Created At
```

## Key-value output

Use key-value output for a small number of scalar fields:

```text
host      https://example:10443
scene     migration
insecure  false
```

Keep stable spacing and key order. Use JSON for many fields or nested structures.

## Vertical output

```text
*************************** 1. row ***************************
ID      account-1
Name    demo
Status  Available

*************************** 2. row ***************************
ID      account-2
Name    prod
Status  Creating
```

Start each record with the fixed row marker, number from 1, preserve table column order,
localize field names according to the current language, and leave one blank line between rows.
An empty list produces no rows.

## Multiple sections

Use a localized `== Title ==` heading and a dash separator for each peer section:

```text
== Block Storage Clouds ==
-------------------------
Provider  Name
aliyun    Alibaba Cloud

== Object Storage Clouds ==
-------------------------
Provider  Name
aliyun    Alibaba Cloud
```

Keep one blank line between sections. JSON never contains these headings.

## Resource candidates

Group candidates by resource type, retain a heading even for one resource type, localize the
heading, and show only fields needed to select a value for a later command.

```text
== Regions ==
-------------------------
Region ID    Region Name
cn-beijing   Beijing

== Zones ==
-------------------------
Zone ID       Zone Name
cn-beijing-a  Beijing A
```

Common columns include `Region ID`, `Region Name`, `Zone ID`, `Flavor ID`, `Flavor Name`,
`vCPUs`, `RAM`, `Image ID`, `Image Name`, `OS`, `Project ID`, `Project Name`, `Network ID`,
and `CIDR`.

## Request preview

Print the final request body as plain JSON only:

```json
{
  "name": "demo",
  "cloud_type": "aliyun",
  "storage_type": "block"
}
```

Do not add a title, header, explanation, or hint. Keep structured field names unchanged.

## Errors

Localized structured remote errors use four lines:

```text
Error: Authentication failed
Details: username or password is incorrect
Code: 00008002
Trace ID: 4ca752f0875f40b18b3c0a887dfbe508
```

Localize the labels, show summary, details, code, and trace ID, and use `-` for missing
fields. Local validation errors remain short, such as `id is required`. Never print tokens,
passwords, or keys.

## Debug logs

Use one key-value line per event:

```text
DEBUG method=GET url=https://example/api/v2/getHosts status=200 duration=120ms trace_id=trace-1
```

Start with `DEBUG`, separate fields with one space, use `key=value`, and represent a request
body as `body=<json>`. Mask sensitive values as `***`; tokens, passwords, and keys must never
appear in debug logs.

## Localization

Localize table headers, section headings, and error labels according to the active language.
Do not localize JSON field names. Use one stable label for a field in every output mode.
