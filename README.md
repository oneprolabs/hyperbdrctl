# hyperbdrctl

`hyperbdrctl` 是 HyperBDR / HyperMotion 平台的命令行客户端，使用 Go 开发，面向日常运维、批量执行和脚本集成场景。

CLI 通过统一的 HTTP API 与平台交互，支持 `dr` 和 `migration` 两种场景。

## 核心能力

- 管理本地连接配置、语言和默认输出格式
- 查询主机、快照、任务、许可证和升级信息
- 执行主机注册、同步、引导、注销和等待类操作
- 管理主机启动配置与顶层 `boot-config apply` 流程
- 管理源连接，包括 Agent / Agentless 安装信息和 Agentless 源连接创建
- 管理目标侧云账号、云同步网关和对象存储
- 通过 `api request` 直接访问已认证 API，便于排障和补充自动化

## 适用场景

- 需要将 HyperBDR / HyperMotion 操作纳入 Shell、CI/CD 或批处理脚本
- 需要通过 JSON 输出对接自定义自动化系统
- 需要统一管理多环境配置，并在 `dr` / `migration` 场景间切换
- 需要在不登录 Web 页面时完成常见运维动作

## 环境要求

- Go 1.18 或更高版本
- 可访问的 HyperBDR / HyperMotion 平台地址，例如 `https://<host>:10443`
- 有效的平台用户名和密码

## 构建

在项目根目录执行：

```sh
go build -o hyperbdrctl ./cmd/hyperbdrctl
```

如需生成更适合分发的精简二进制，可执行：

```sh
go build -trimpath -ldflags="-s -w" -o hyperbdrctl ./cmd/hyperbdrctl
```

Windows 下可根据需要输出为 `hyperbdrctl.exe`。

## 快速开始

### 1. 保存本地默认配置

```sh
hyperbdrctl config set \
  --host https://example:10443 \
  --username admin \
  --password <password> \
  --scene migration \
  --lang zh_cn
```

如果是测试环境，且确实需要跳过证书校验，可显式增加：

```sh
--insecure
```

`config set` 会先执行登录校验；如果认证失败，不会写入本地配置。

### 2. 执行基础查询

```sh
hyperbdrctl host list --page 1 --page-size 10
hyperbdrctl target account list
hyperbdrctl tasks list
```

### 3. 在脚本中使用 JSON 输出

```sh
hyperbdrctl host list --page 1 --page-size 10 --output json
```

默认输出为表格；当使用 `--output json` 时，CLI 输出原始 API 字段名，便于脚本直接消费。

## 配置方式

CLI 支持以下配置来源：

1. 命令行参数
2. 环境变量
3. 本地配置文件
4. 默认值

固定优先级为：

```text
命令行参数 > 环境变量 > 配置文件 > 默认值
```

常用环境变量如下：

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

说明：

- `--lang` 同时控制 CLI 文案和 HTTP 请求头 `X-LANG`
- 支持语言值：`en`、`zh_cn`
- 默认输出格式：`table`
- `config get` 默认隐藏密码
- TLS 默认校验证书，仅在测试环境下建议显式使用 `--insecure`

## 典型工作流

### 主机与启动配置

```sh
hyperbdrctl host list
hyperbdrctl host detail --id <host_id>
hyperbdrctl host boot-config get --id <host_id>
hyperbdrctl boot-config apply --id <host_id> --file ./boot-config.json
hyperbdrctl host wait --id <host_id>
```

说明：

- 稳定的主机级启动配置管理使用 `host boot-config`
- 独立的顶层单主机覆盖式 apply 流程使用 `boot-config apply`

### 源连接

查看源端安装信息：

```sh
hyperbdrctl source agent-install
hyperbdrctl source agentless-install
hyperbdrctl source sync-nodes
```

创建 Agentless 源连接：

```sh
hyperbdrctl source create \
  --type vmware \
  --synch-node-id <node_id> \
  --auth-url https://vcenter.example:443 \
  --auth-key <username> \
  --auth-cert <password>
```

创建完成后，可执行：

```sh
hyperbdrctl source list --type vmware --binding-status binding
```

### 目标云账号

```sh
hyperbdrctl target supports
hyperbdrctl target account list
hyperbdrctl target account fetch-block-resources --help
hyperbdrctl target account create-block aliyun --help
hyperbdrctl target account create-oss openstack --help
```

CLI 通过按云厂商拆分的子命令承载写入类流程。当前支持的 provider 列表以运行时 `target supports` 和对应 `create-* --help` 输出为准。

### 云同步网关

```sh
hyperbdrctl target cloud-sync-gateway list
hyperbdrctl target cloud-sync-gateway resources --cloud-account-id <account_id> --output json
hyperbdrctl target cloud-sync-gateway create aliyun --help
hyperbdrctl target cloud-sync-gateway wait --id <storage_id>
```

建议流程：

1. 先确认或创建目标云账号
2. 查询创建所需资源
3. 按云厂商执行网关创建
4. 通过 `wait` 等待最终状态

### 对象存储

```sh
hyperbdrctl target oss list
hyperbdrctl target oss detail --id <storage_id>
hyperbdrctl target oss buckets --help
hyperbdrctl target oss create --help
```

## 命令总览

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
|- licenses
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

如需查看某个命令的完整参数和示例，请执行：

```sh
hyperbdrctl --help
hyperbdrctl config set --help
hyperbdrctl host --help
hyperbdrctl source create --help
hyperbdrctl target account create-block --help
hyperbdrctl target cloud-sync-gateway create --help
```

## 发布与集成建议

- 优先为自动化场景固定 `--output json`
- 在多环境切换时，建议显式设置 `--scene`
- 生产环境不要默认开启 `--insecure`
- 如果需要最小化人工输入，优先使用环境变量或预置配置文件
- 如需排障，可附加 `--debug` 输出请求调试日志

## 开发验证

在仓库规范下，涉及代码修改时应在项目根目录执行：

```sh
gofmt -w cmd internal
go test ./...
go build -o ../hyperbdrctl ./cmd/hyperbdrctl
```
