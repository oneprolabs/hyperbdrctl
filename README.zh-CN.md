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

### 配置环境

使用前准备一个可访问的 HyperBDR / HyperMotion 地址、有效的平台凭据和已发布的二进制。

配置平台地址和认证信息：

```sh
hyperbdrctl config set \
  --host https://<host>:10443 \
  --username <username> \
  --password <password> \
  --scene dr \
  --lang zh_cn

hyperbdrctl config get
```

`config set` 会先校验凭据，认证成功后才保存配置。测试环境使用不受信任的证书时，才启用：

```sh
hyperbdrctl config set --insecure
```

配置也可以通过环境变量提供，优先级如下：

```text
命令行参数 > 环境变量 > 本地配置文件 > 默认值
```

常用环境变量：

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

`--lang` 支持 `en`、`zh_cn` 和 `ja`；默认启用 TLS 校验。

### 使用流程与命令帮助

HyperBDR 的基本流程是：

```text
源端准备 → 目标端准备 → 容灾配置 → 同步和启动 → 清理资源
```

按流程查看对应命令的详细参数和示例：

```sh
# 源端准备：生产站点、同步代理、虚拟机注册
hyperbdrctl production-site --help
hyperbdrctl sync-proxy --help
hyperbdrctl host register --help

# 目标端准备：云账号、云资源、对象存储、同步网关
hyperbdrctl cloud-account --help
hyperbdrctl cloud-resource --help
hyperbdrctl oss --help
hyperbdrctl cloud-sync-gateway --help

# 容灾配置：主机、快照、启动配置
hyperbdrctl host --help
hyperbdrctl host snapshots --help
hyperbdrctl boot-config --help

# 同步和启动
hyperbdrctl host sync --help
hyperbdrctl host boot --help
hyperbdrctl host wait --help

# 清理资源
hyperbdrctl host clean --help
hyperbdrctl host deregister --help
hyperbdrctl cloud-sync-gateway delete --help
hyperbdrctl oss delete --help
```

自动化场景建议使用 JSON 输出；人工排障时可使用 `--vertical` 和 `--debug`：

```sh
hyperbdrctl --output json host list
```

## 文档

参见[命令行样式文档](docs/README.md)，其中包含 help 页面和终端输出的多语言规范：

- [CLI Help 样式规范](docs/cli-help-style.zh.md) · [English](docs/cli-help-style.en.md)
- [CLI 输出样式规范](docs/cli-output-style.zh.md) · [English](docs/cli-output-style.en.md)

## 特性

- **为 AI agent 设计** — 子命令清晰、参数显式、支持 JSON 和异步 `wait`，便于 agent 规划、执行和校验每一步操作。
- **自动化优先的配置** — 支持命令行参数、环境变量和本地配置文件，不依赖浏览器会话或交互式控制台。
- **统一支持 DR 和迁移** — 只需切换 `--scene`，无需改变自动化模型。
- **覆盖完整运维流程** — 查询主机和快照、启动配置、源端准备、云资源发现、网关、对象存储和 License。
- **安全默认值** — 默认开启 TLS 校验，`config get` 默认隐藏敏感信息，凭据校验成功后才保存配置。
- **兼顾人工排障** — 提供表格、纵向输出、多语言文案（`en` / `zh_cn` / `ja`）和命令级帮助。
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

新增或修改命令时，请同步更新多语言帮助文案和测试。保持 agent 使用的 JSON 输出稳定，并明确记录破坏性变更。

## 项目状态

`hyperbdrctl` 当前处于积极开发阶段。随着 HyperBDR / HyperMotion API 演进，命令和云厂商专属请求字段
可能继续调整。生产自动化请固定已知版本，并先在测试环境验证升级。

## 参与贡献

欢迎提交 Issue 和 Pull Request。行为变更请补充测试、运行上述开发检查，并说明对 API 或自动化流程的影响。

## License

本项目使用 [Apache License 2.0](LICENSE) 授权。

