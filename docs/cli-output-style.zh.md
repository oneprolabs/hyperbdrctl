# CLI 输出样式规范

[English](cli-output-style.en.md) | 中文

本文档用于定义 `hyperbdrctl` 非 help 输出的终端呈现样式和文案口径。

## 适用范围

以下输出默认受这份规范约束：

- 默认表格输出
- JSON 输出
- 纵向输出
- key-value 输出
- 多段输出标题
- 资源候选输出
- 请求预览输出
- 错误输出
- Debug 日志输出

## 基本原则

- 默认输出面向操作员阅读，优先清晰、稳定、可复制。
- JSON 输出面向脚本和排障，保持结构化字段名，不做本地化。
- 表格只展示关键字段，不把原始响应字段全部塞进默认输出。
- 同类输出保持相同标题、表头、列顺序和空值表现。
- 中文文案保持正式、直接、操作员导向。
- 英文文案保持简洁、直接，避免长句和解释性背景。

## 输出模式

| 模式 | 触发方式 | 样式定位 |
| --- | --- | --- |
| 表格 | 默认，或 `--output table` | 面向人工阅读，展示关键字段。 |
| JSON | `--output json` | 面向脚本、自动化和排障，保留结构化字段。 |
| 纵向 | `--vertical` 或 `-G` | 面向宽表列表结果，逐条展开字段。 |

规则：

- 默认输出为表格。
- JSON 输出不包含表格表头、分段标题、说明文字或其它装饰内容。
- 纵向输出只改变表格结果的展示方式，不改变字段含义。

## JSON 输出

JSON 输出使用缩进格式：

```json
{
  "id": "host-1",
  "name": "app-01",
  "status": "running"
}
```

规则：

- 使用两个空格缩进。
- 输出必须是合法 JSON。
- 字段名保持结构化字段名，不做本地化。
- 不混入日志、标题、表头或说明文字。
- 不使用中文标点或本地化字段名替换 JSON key。

## 表格输出

表格输出由表头和数据行组成：

```text
ID      Name    Status
host-1  app-01  running
host-2  db-01   stopped
```

规则：

- 第一行为表头。
- 后续每一行为数据。
- 列之间使用两个空格分隔。
- 表头使用当前语言本地化。
- 字段按显示宽度对齐。
- 中文、日文、韩文和全角字符按宽字符处理。
- 空值显示为空字符串。
- 布尔值显示为 `true` 或 `false`。
- 数字按普通数字显示。
- 复杂值只在确实需要时显示为紧凑 JSON 字符串。

## 表格列

规则：

- 列只展示当前命令最重要、最常用于继续操作的字段。
- ID、名称、状态、类型、时间等关键字段优先靠前。
- 状态字段优先展示面向用户的显示值。
- 默认表格不展示大量原始字段。
- 全空列可以隐藏。
- 空列表保留表头，不额外输出说明句。
- 同一类资源在不同命令中保持表头命名一致。

常见列顺序：

```text
ID  Name  Type  Status  Created At
```

## Key-Value 输出

少量标量字段可以使用 key-value 输出：

```text
host      https://example:10443
scene     migration
insecure  false
```

规则：

- key 和 value 使用稳定间距分隔。
- key 按稳定顺序输出。
- 只用于少量标量字段。
- 字段较多或存在嵌套结构时，使用 JSON 输出。
- key 不做句子化描述，保持短字段名风格。

## 纵向输出

纵向输出按记录逐条展开：

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

规则：

- 每条记录以 `*************************** N. row ***************************` 开头。
- `N` 从 1 开始。
- 字段顺序与表格列顺序一致。
- 字段名使用当前语言本地化。
- 多条记录之间保留一个空行。
- 空列表输出为空。

## 多段输出标题

多个并列分段使用标题和分隔线：

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

规则：

- 标题格式固定为 `== 标题 ==`。
- 标题下一行输出分隔线。
- 分隔线使用 `-`。
- 不同分段之间保留一个空行。
- 分段标题使用当前语言本地化。
- JSON 输出不包含分段标题。
- 同类输出不要混用纯文本标题、冒号标题和 `==` 标题。

## 资源候选输出

资源候选结果按资源类型展示：

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

规则：

- 多资源结果按资源类型分段。
- 单资源结果仍保留资源标题。
- 资源分段标题使用当前语言本地化。
- 表格列只展示选择资源所需字段。
- 字段命名优先使用用户能在后续命令中识别和传入的名称。

常见资源列：

| 资源 | 常见列 |
| --- | --- |
| regions | Region ID、Region Name、Local Name |
| zones | Zone ID、Zone Name |
| flavors | Flavor ID、Flavor Name、vCPUs、RAM、Zone ID |
| images | Image ID、Image Name、OS、OS Version、Boot Mode |
| projects | Project ID、Project Name、Domain ID |
| networks | ID、Name、CIDR |
| subnets | ID、Name、Network ID、Zone ID、CIDR |

## 请求预览输出

请求预览输出最终请求体，使用纯 JSON：

```json
{
  "name": "demo",
  "cloud_type": "aliyun",
  "storage_type": "block"
}
```

规则：

- 使用 JSON 样式。
- 只展示请求体本身。
- 不输出标题、表头、说明句或额外提示。
- 字段名保持结构化字段名。

## 错误输出

结构化远端错误使用四行格式：

```text
错误: 认证失败
详情: 用户名或密码不正确
错误码: 00008002
追踪ID: 4ca752f0875f40b18b3c0a887dfbe508
```

英文格式：

```text
Error: Authentication failed
Details: username or password is incorrect
Code: 00008002
Trace ID: 4ca752f0875f40b18b3c0a887dfbe508
```

规则：

- 错误标题使用当前语言本地化。
- 结构化远端错误展示摘要、详情、错误码和追踪 ID。
- 缺失字段使用 `-` 占位。
- 本地参数校验错误保持短句，例如 `id is required`。
- 错误输出不包含 token、密码或密钥。

## Debug 日志

Debug 日志使用单行 key-value 样式：

```text
DEBUG method=GET url=https://example/api/v2/getHosts status=200 duration=120ms trace_id=trace-1
```

规则：

- 每条 Debug 日志占一行。
- 固定以 `DEBUG` 开头。
- 后续字段使用 `key=value` 形式。
- 字段之间使用一个空格分隔。
- 请求体存在时使用 `body=<json>`。
- 敏感字段显示为 `***`。
- token、密码和密钥不得出现在 Debug 日志中。

## 本地化

规则：

- 表格表头使用当前语言本地化。
- 分段标题使用当前语言本地化。
- 错误标题使用当前语言本地化。
- JSON 字段名不做本地化。
- 同一字段在不同输出中使用同一套表头文案。
