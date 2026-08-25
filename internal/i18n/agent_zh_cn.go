package i18n

var agentZH = map[string]string{
	"cmd.agent.short":               "Agent 源端代理管理",
	"cmd.agent.long":                "读取源端主机 Agent 安装命令。",
	"cmd.agent.examples":            "./hyperbdrctl agent install\n./hyperbdrctl --output json agent install",
	"cmd.agent.notes":               "- install 命令上的未识别参数会透传为查询参数。",
	"cmd.agent.usage_line":          "hyperbdrctl agent <命令> [参数]",
	"cmd.agent.usage_notes":         "管理 Agent 模式下的源端主机准备流程。\n\n输出源端代理安装说明：\n  hyperbdrctl agent install",
	"cmd.agent.install.short":       "输出 Agent 源端代理安装说明",
	"cmd.agent.install.long":        "读取 Agent 源端代理安装元数据，并输出精简后的 Linux 与 Windows 安装命令。",
	"cmd.agent.install.examples":    "./hyperbdrctl agent install\n./hyperbdrctl --output json agent install",
	"cmd.agent.install.notes":       "- 未识别参数会透传为查询参数。\n- 默认输出会按 Linux 和 Windows 卡片展示支持系统、建议和精简安装命令。\n- 若需要完整的各系统原始元数据，请使用 JSON 输出。",
	"cmd.agent.install.usage_line":  "hyperbdrctl agent install [参数]",
	"cmd.agent.install.usage_notes": "输出 Agent 模式下源端代理安装所需的操作说明。\n\n查看默认安装说明：\n  hyperbdrctl agent install\n\n默认输出按 Linux 和 Windows 分段展示安装命令。\n\n安装完成后，可返回 CLI 查看源端主机：\n  hyperbdrctl host list\n\n如需脚本化比对完整原始元数据，可附加：\n  --output json",
}
