package i18n

var agentEN = map[string]string{
	"cmd.agent.short":               "Agent source proxy management",
	"cmd.agent.long":                "Read Agent installation commands for source hosts.",
	"cmd.agent.examples":            "./hyperbdrctl agent install\n./hyperbdrctl --output json agent install",
	"cmd.agent.notes":               "- Unknown flags are passed through as query parameters on install.",
	"cmd.agent.usage_line":          "hyperbdrctl agent <command> [flags]",
	"cmd.agent.usage_notes":         "Manage the source host preparation workflow for Agent mode.\n\nOutput source proxy installation guidance:\n  hyperbdrctl agent install",
	"cmd.agent.install.short":       "Output Agent source proxy installation guidance",
	"cmd.agent.install.long":        "Fetch Agent source proxy installation metadata and render the simplified Linux and Windows install commands.",
	"cmd.agent.install.examples":    "./hyperbdrctl agent install\n./hyperbdrctl --output json agent install",
	"cmd.agent.install.notes":       "- Unknown flags are passed through as query parameters.\n- Default output renders Linux and Windows install cards with supported systems, recommendations, and simplified install commands.\n- Use JSON output when you need the full raw per-OS metadata.",
	"cmd.agent.install.usage_line":  "hyperbdrctl agent install [flags]",
	"cmd.agent.install.usage_notes": "Output the operating instructions required to install the source proxy in Agent mode.\n\nView the default installation guidance:\n  hyperbdrctl agent install\n\nDefault output shows the installation commands in separate Linux and Windows sections.\n\nAfter installation, return to the CLI to view source hosts:\n  hyperbdrctl host list\n\nTo compare the complete raw metadata in scripts, add:\n  --output json",
}
