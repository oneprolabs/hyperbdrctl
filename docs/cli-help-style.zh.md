# CLI Help 样式规范

[English](cli-help-style.en.md) | 中文

本文档用于约束 `hyperbdrctl` 的 CLI help 页面样式。

## 命令示例

### 创建资源类命令示例
```text
创建阿里云块存储账号

用法: hyperbdrctl cloud-account create --cloud-type aliyun --storage-type block [参数]

参数:
      --access-key-id string       云平台 Access Key ID（必须）
      --access-key-secret string   云平台 Access Key Secret（必须）
      --auth-region-id string      认证地域 ID（必须）
      --set stringArray            metadata 路径覆盖，格式 path=value，可重复
      --set-json stringArray       metadata 路径覆盖，格式 path=<json>，可重复
      --preview-request            输出请求体，但不发送请求
      --debug                      输出请求调试日志
      --lang string                显示语言，可选值 en / zh_cn / ja，默认值 en
  -o, --output string              输出格式，可选值 table / json，默认值 table
  -h, --help                       显示帮助信息

使用说明:
  创建阿里云块存储云账号。

  参数来源：
    --access-key-id string
      使用阿里云账号的 Access Key ID。

    --access-key-secret string
      使用阿里云账号的 Access Key Secret。

    --auth-region-id string
      使用阿里云认证地域 ID。

  资源获取：
    先查询可用认证地域：
      hyperbdrctl cloud-resource fetch --cloud-type aliyun --storage-type block \
        --access-key-id <ak> \
        --access-key-secret <sk> \
        --fetch-res regions \
        --output json

  最小创建命令如下：
    hyperbdrctl cloud-account create --cloud-type aliyun --storage-type block \
      --access-key-id <ak> \
      --access-key-secret <sk> \
      --auth-region-id <auth_region_id>

  创建成功后，等待任务完成：
    hyperbdrctl cloud-account wait --id <account_id>

  任务结束后，查看云账号详情：
    hyperbdrctl cloud-account detail --id <account_id>
    hyperbdrctl cloud-account list --storage-type block

  如需更多字段，可重复附加：
    --set path=value
    --set-json path=<json>

  参数覆盖顺序固定为：
    --set-json < --set < 显式创建参数

  可按需补充以下动态参数：
    --account-name <name>
      云账号名称

  如需先检查最终请求体，可附加下面的参数：
    --preview-request

```

## 适用范围

以下场景默认受这份规范约束：

- 新增 CLI help 页面
- 修改已有 CLI help 页面
- 调整 help 生成逻辑、区块顺序或可见性策略
- 为 help 页面补测试、补示例或补操作引导

## 基本设计原则

### 页面结构

叶子命令使用四段式：

1. 描述
2. 用法
3. 参数
4. 使用说明

命令组入口在需要展示稳定子命令时使用五段式：

1. 描述
2. 用法
3. 参数
4. 命令
5. 使用说明

规则：

- 描述必须先说明命令用途。
- `用法` 和 `参数` 必须紧跟描述。
- 叶子命令默认不使用 `命令` 区块。
- 只有存在稳定子命令集合，且 `命令` 区块对用户有直接导航价值时，才使用五段式。
- 命令组入口不要写成单功能命令页。
- 同一命令的 help 结构应长期稳定。

### 参数区规则

- 只保留一个 `参数` 区块。
- 参数顺序固定为：必填参数、当前命令参数、全局参数。
- 必须参数在说明后追加 `（必须）`，不要单独拆分必填参数区块。
- 参数短描述只说明“这是什么”。
- 参数来源、查询步骤、接口来源和自动补全行为写入 `使用说明`。
- 固定候选值写作 `可选值 a / b`。
- 固定回退值写作 `默认值 x`。
- `可选值` 和 `默认值` 后面不加冒号。
- 多个可选值统一使用 `/` 分隔。
- 没有最终回退值的参数不写默认值。
- 默认值依赖查询结果、接口返回或自动补全逻辑时，放在 `参数来源` 或 `使用说明` 中解释。
- 全局参数固定放在参数区末尾。

示例：

```text
      --name string     资源名称（必须）
      --type string     资源类型，可选值 a / b，默认值 a
      --debug           输出请求调试日志
      --lang string     显示语言，可选值 en / zh_cn / ja，默认值 en
  -o, --output string   输出格式，可选值 table / json，默认值 table
  -h, --help            显示帮助信息
```

### 使用说明规则

`使用说明` 应面向操作员，按真实执行顺序组织：

1. 准备什么
2. 先查什么
3. 最小命令
4. 可选扩展
5. 预览或验证

规则：

- 使用完整 `hyperbdrctl ...` 命令，不使用命令片段，但允许单个参数说明。
- 参数来源、查询步骤和自动补全行为写在 `使用说明`，不要写进参数短描述。
- 避免解释内部实现、历史兼容和接口字段来源，除非这些信息直接影响操作。
- 关键业务参数存在固定来源或需要先查询时，可在 `使用说明` 中增加 `参数来源` 子段。
- `参数来源` 只描述当前命令的关键业务参数，不重复解释 `--debug`、`--lang`、`--output`、`--help` 等全局参数。
- 多个参数使用一个描述时，参数要换行，每个参数占一行。

### 先检查已有代码逻辑，再设计样式改造方式

更新命令行样式时，应先检查已有 help 渲染逻辑、命令构造方式和现有测试约束。

规则：

- 优先复用已有的 help 渲染入口、命令构造函数和文案挂载方式。
- 优先通过共用逻辑调整样式，不要为单个命令无理由新增一套独立实现。
- 优先最小改动完成样式更新，避免把样式调整扩散成无关重构。
- 如果多个命令存在相同的样式问题，应优先抽取共用修正，而不是逐个命令重复修改。
- 更新样式方案时，应同时说明哪些逻辑已经可以复用，哪些地方必须单独处理。

## 文案与排版规则

### 标题风格

- 区块标题命名应稳定，不随命令随意变体。
- 同一含义只保留一套标题，例如不要并存多个近义标题来表达相同区块。

### 多语言术语

CLI help 必须同时支持 `zh_cn`、`en` 和 `ja`。新增或调整 help 文案时，应使用下列稳定术语：

| 含义 | zh_cn | en | ja |
| --- | --- | --- | --- |
| 用法区块 | `用法` | `Usage` | `使い方` |
| 参数区块 | `参数` | `Flags` | `フラグ` |
| 命令区块 | `命令` | `Commands` | `コマンド` |
| 使用说明区块 | `使用说明` | `Usage Notes` | `使用上の注意` |
| 参数来源子段 | `参数来源` | `Parameter Sources` | `パラメータの出所` |
| 必填标记 | `（必须）` | `(required)` | `(必須)` |
| 可选值 | `可选值 a / b` | `allowed values a / b` | `使用可能な値 a / b` |
| 默认值 | `默认值 x` | `default x` | `デフォルト x` |
| 显示帮助信息 | `显示帮助信息` | `Show help information` | `ヘルプ情報を表示` |

规则：

- 同一语言内不要混用标题近义词，例如不要同时使用 `参数`、`选项`、`Flags` 表达同一区块。
- 中文 help 使用全角中文标点；英文 help 使用英文标点。
- 可选值在中英文中都使用 `/` 分隔，分隔符两侧保留一个空格。
- 所有语言的区块顺序、参数顺序和示例结构必须保持一致。

### 终端多段输出标题

规则：

- 当非 help 命令的默认终端输出需要展示多个并列分段时，分段标题统一使用 `== 标题 ==`。
- 标题下一行统一补分隔线，保持与现有多段表格输出一致。
- 同类输出不要混用纯文本标题、冒号标题和 `==` 标题。

示例：

```text
== 块存储支持列表 ==
-------------------------
Provider       Name
aliyun         Alibaba Cloud(Recommended, SDK v2.0)

== 对象存储支持列表 ==
-------------------------
Provider       Name
aliyun         Alibaba Cloud(Recommended, SDK v2.0)
```

### 缩进

- 区块正文默认保持稳定缩进。
- bullet 或编号换行后，续行必须与上文保持同一层级的视觉对齐。

### 长命令换行

规则：

- 只要命令在普通终端宽度下明显难读，就应手工换行。
- shell 命令续行统一使用结尾 `\`。
- 续行参数相对命令主体增加两个空格缩进。
- 不要完全依赖终端软换行。

### 文案规则

- 保持正式、直接、操作员导向。
- 短描述一句话说明用途。
- 长描述只补充命令边界、自动行为和适用场景。
- 避免 `这是...`、`该命令是...`、`这里可以...` 等口语化判断句。
- 避免把实现细节、历史兼容和接口字段解释写成主要说明。
- 不把大段背景说明塞进短描述。

## 命令分类

按当前公开命令分类。代表命令必须来自实际 `hyperbdrctl --help` 输出；已隐藏、已移除或兼容入口不要写入规范示例。

| 类型 | 代表命令 | 描述重点 | 使用说明重点 |
| --- | --- | --- | --- |
| 变更操作类 | `cloud-account create`、`cloud-sync-gateway create`、`oss create`、`boot-config apply`、`host sync`、`host boot`、`license activate` | 创建、应用、激活、删除或发起什么操作 | 前置准备、最小命令、按需扩展、预览、验证 |
| 查询列表类 | `host list`、`cloud-account list`、`production-site list`、`sync-proxy list`、`license list`、`oss list` | 查询什么列表 | 最小查询、筛选方式、结果如何作为后续输入 |
| 查询详情类 | `host detail`、`cloud-account detail`、`production-site detail`、`cloud-sync-gateway detail`、`oss detail`、`boot-config get` | 查看哪个对象或配置详情 | ID 来源、最小查询、查询后如何继续操作 |
| 说明输出类 | `agent install`、`sync-proxy install`、`license reg-code`、`completion` | 输出什么安装说明、注册码或辅助信息 | 输出内容如何使用、执行后下一步 |
| 引导助手类 | `cloud-resource fetch`、`oss catalog`、`oss buckets`、`production-site vm-list`、`host snapshots` | 帮助收集什么后续输入 | 先查什么、再查什么、结果用于哪个命令 |
