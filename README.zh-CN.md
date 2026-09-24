<div align="center">

# hyperbdrctl

**面向 AI agent 的 HyperBDR / HyperMotion 操作 CLI。**

为 AI agent 和自动化系统提供稳定、可预测的命令接口，用于执行容灾与迁移流程：结构化输出、
显式参数、安全默认值，无需通过浏览器控制台抓取或模拟操作。

[English](README.md) | [中文](README.zh-CN.md)

[![Go](https://img.shields.io/badge/go-1.18%2B-00ADD8?logo=go)](go.mod)
[![License](https://img.shields.io/badge/license-Apache%202.0-blue.svg)](LICENSE)
[![Status](https://img.shields.io/badge/status-active%20development-orange.svg)](#项目状态)

[代码仓库](https://github.com/oneprolabs/hyperbdrctl) · [问题反馈](https://github.com/oneprolabs/hyperbdrctl/issues) · [版本发布](https://github.com/oneprolabs/hyperbdrctl/releases)

</div>

`hyperbdrctl` 是使用 Go 编写的 HyperBDR / HyperMotion 命令行客户端，主要面向 AI agent 和自动化
系统。它可以查询平台状态、执行操作、等待异步任务完成，并返回无需解析网页的机器可读结果。
CLI 通过统一的 HTTP API 与平台通信，同时支持 `dr` 和 `migration` 场景。

## 使用方式

### 前置条件

- 一个可访问的 HyperBDR / HyperMotion 地址，例如 `https://<host>:10443`
- 有效的平台凭据
- 已发布的二进制

### 配置平台地址

`config set` 会先校验合并后的凭据，只有认证成功才会写入文件；认证失败不会覆盖已有配置。

```sh
hyperbdrctl config set --host https://example:10443 --username admin --password <password> --scene migration --lang zh_cn
hyperbdrctl config get
```

如果测试环境使用不受信任的证书，需要显式启用：

```sh
hyperbdrctl config set --insecure
```

### 查询和执行操作

```sh
hyperbdrctl host list --page 1 --page-size 10
hyperbdrctl cloud-account list
hyperbdrctl oss list
hyperbdrctl host wait --id <host_id>
```

在任意层级使用 `--help` 查看可用参数和云厂商专属示例：

```sh
hyperbdrctl --help
hyperbdrctl host --help
hyperbdrctl cloud-account create --help
```

### 为 AI agent 输出机器可读结果

在工具调用、规划、校验和后续动作中使用 JSON。JSON 会保留 API 字段名，适合直接交给下一个自动化步骤处理。

```sh
hyperbdrctl --output json host list --page 1 --page-size 10
hyperbdrctl --output json cloud-resource fetch --cloud-account-id <account_id>
```

默认输出为表格；`--vertical` 适合查看单条记录，`--debug` 可开启请求级故障排查。

### 配置来源

配置按以下优先级解析：

```text
命令行参数 > 环境变量 > 本地配置文件 > 默认值
```

支持的环境变量：

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

`--lang` 同时控制 CLI 文案和 HTTP `X-LANG` 请求头，支持 `en` 和 `zh_cn`。默认启用 TLS 校验；
只有在明确可信的测试环境中才应使用 `--insecure`。

### 常见工作流

```sh
# 主机和启动配置
hyperbdrctl host list
hyperbdrctl host detail --id <host_id>
hyperbdrctl boot-config get --id <host_id>
hyperbdrctl boot-config apply --id <host_id> --file ./boot-config.json
hyperbdrctl host wait --id <host_id>

# 源端准备
hyperbdrctl agent install
hyperbdrctl sync-proxy install
hyperbdrctl sync-proxy list
hyperbdrctl production-site create --type vmware --synch-node-id <node_id> --auth-url https://vcenter.example:443 --auth-key <username> --auth-cert <password>

# 目标云资源
hyperbdrctl cloud-account list
hyperbdrctl cloud-resource fetch --cloud-account-id <account_id> --output json
hyperbdrctl cloud-sync-gateway create --help
hyperbdrctl cloud-sync-gateway wait --id <storage_id>
hyperbdrctl oss catalog

# License 管理
hyperbdrctl license list
hyperbdrctl license reg-code
hyperbdrctl license activate --kkty <reg_code> --ddty <activation_code>
```

### 命令概览

```text
hyperbdrctl
├── config                 get / set
├── completion             生成 Shell 自动补全脚本
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

## 特性

- **为 AI agent 设计** — 子命令清晰、参数显式、支持 JSON 和异步 `wait`，便于 agent 规划、执行和校验每一步操作。
- **自动化优先的配置** — 支持命令行参数、环境变量和本地配置文件，不依赖浏览器会话或交互式控制台。
- **统一支持 DR 和迁移** — 只需切换 `--scene`，无需改变自动化模型。
- **覆盖完整运维流程** — 查询主机和快照、启动配置、源端准备、云资源发现、网关、对象存储和 License。
- **安全默认值** — 默认开启 TLS 校验，`config get` 默认隐藏敏感信息，凭据校验成功后才保存配置。
- **兼顾人工排障** — 提供表格、纵向输出、双语文案（`en` / `zh_cn`）和命令级帮助。
- **易集成、易分发** — 单个 Go 二进制即可运行，除访问平台地址外没有额外运行时依赖。

## 构建和开发

### 从源码构建

从源码构建需要 Go 1.18 或更高版本。

```sh
go build -o hyperbdrctl ./cmd/hyperbdrctl
go build -trimpath -ldflags="-s -w" -o hyperbdrctl ./cmd/hyperbdrctl
```

Windows 环境可按需将输出文件名改为 `hyperbdrctl.exe`。

### 发布构建和版本信息

```sh
VERSION=v1.2.3
go build -trimpath -buildvcs=true -ldflags "-s -w -X hyperbdr-client/internal/version.Version=${VERSION}" -o hyperbdrctl ./cmd/hyperbdrctl
./hyperbdrctl --version
go version -m ./hyperbdrctl
```

### 开发检查

```sh
gofmt -w cmd internal
go test ./...
go build -o hyperbdrctl ./cmd/hyperbdrctl
```

新增或修改命令时，请同步更新双语帮助文案和测试。保持 agent 使用的 JSON 输出稳定，并明确记录破坏性变更。

## 项目状态

`hyperbdrctl` 当前处于积极开发阶段。随着 HyperBDR / HyperMotion API 演进，命令和云厂商专属请求字段
可能继续调整。生产自动化请固定已知版本，并先在测试环境验证升级。

## 参与贡献

欢迎提交 Issue 和 Pull Request。行为变更请补充测试、运行上述开发检查，并说明对 API 或自动化流程的影响。

## License

本项目使用 [Apache License 2.0](LICENSE) 授权。

