package commands

import (
	"strings"

	"github.com/spf13/cobra"
)

func newRootCommand(ctx *context) *cobra.Command {
	root := &cobra.Command{
		Use:     "hyperbdrctl",
		Short:   ctx.loc.T("cmd.root.short"),
		Long:    ctx.loc.T("cmd.root.long"),
		Example: strings.TrimSpace(ctx.loc.T("cmd.root.examples")),
		RunE: func(cmd *cobra.Command, args []string) error {
			return renderHelp(cmd, ctx)
		},
	}
	root.SetHelpFunc(func(cmd *cobra.Command, _ []string) {
		_ = renderHelp(cmd, ctx)
	})
	root.SetHelpCommand(newHelpCommand(ctx))
	addGlobalHelpFlags(root, ctx)
	addHelpLayout(root, helpLayoutRoot)
	addUsageLine(root, ctx, "cmd.root.usage_line")
	addUsageNotes(root, ctx, "cmd.root.usage_notes")

	root.AddCommand(
		newAPICommand(ctx),
		newConfigCommand(ctx),
		newHostCommand(ctx),
		newTopLevelBootConfigCommand(ctx),
		newBootConfigCLIAliasCommand(ctx),
		newBatchBootConfigCommand(ctx),
		newTasksCommand(ctx),
		newSourcesCommand(ctx),
		newLicensesCommand(ctx),
		newTargetCommand(ctx),
		newUpgradeCommand(ctx),
	)
	configureBuiltinHelpArtifacts(root, ctx)
	return root
}

func newHelpCommand(ctx *context) *cobra.Command {
	return &cobra.Command{
		Use:   "help [command]",
		Short: ctx.loc.T("cmd.help.short"),
		Long:  ctx.loc.T("cmd.help.long"),
		Args:  cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			target := cmd.Root()
			if len(args) > 0 {
				found, _, err := cmd.Root().Find(args)
				if err != nil {
					return err
				}
				target = found
			}
			return renderHelp(target, ctx)
		},
	}
}

func addGlobalHelpFlags(cmd *cobra.Command, ctx *context) {
	flags := cmd.PersistentFlags()
	flags.String("lang", "", ctx.loc.T("flag.lang"))
	flags.StringP("output", "o", "", ctx.loc.T("flag.output"))
	flags.BoolP("vertical", "G", false, ctx.loc.T("flag.vertical"))
	flags.Bool("debug", false, ctx.loc.T("flag.debug"))
}

func newGroupCommand(ctx *context, use, shortKey, longKey, exampleKey, notesKey, groupName string) *cobra.Command {
	cmd := &cobra.Command{
		Use:     use,
		Short:   ctx.loc.T(shortKey),
		Long:    ctx.loc.T(longKey),
		Example: strings.TrimSpace(ctx.loc.T(exampleKey)),
		FParseErrWhitelist: cobra.FParseErrWhitelist{
			UnknownFlags: true,
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				return errUnknown(groupName, args[0])
			}
			return errUnknown(groupName, "")
		},
	}
	addNotes(cmd, ctx, notesKey)
	return cmd
}

func newRawLeafCommand(ctx *context, use, shortKey, longKey, exampleKey, notesKey string, addFlags func(*cobra.Command), run func([]string) error) *cobra.Command {
	cmd := &cobra.Command{
		Use:                use,
		Short:              ctx.loc.T(shortKey),
		Long:               ctx.loc.T(longKey),
		Example:            strings.TrimSpace(ctx.loc.T(exampleKey)),
		DisableFlagParsing: true,
		Args:               cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if rawArgsHelp(cmd, args) {
				return renderHelp(cmd, ctx)
			}
			return run(args)
		},
	}
	addNotes(cmd, ctx, notesKey)
	if addFlags != nil {
		addFlags(cmd)
	}
	return cmd
}

func addNotes(cmd *cobra.Command, ctx *context, key string) {
	if key == "" {
		return
	}
	notes := strings.TrimSpace(ctx.loc.T(key))
	if notes == "" || notes == key {
		return
	}
	if cmd.Annotations == nil {
		cmd.Annotations = map[string]string{}
	}
	cmd.Annotations[notesAnnotation] = notes
}

func newAPICommand(ctx *context) *cobra.Command {
	cmd := newGroupCommand(ctx, "api", "cmd.api.short", "cmd.api.long", "cmd.api.examples", "cmd.api.notes", "api")
	cmd.Hidden = true
	cmd.AddCommand(newRawLeafCommand(ctx, "request", "cmd.api.request.short", "cmd.api.request.long", "cmd.api.request.examples", "cmd.api.request.notes", func(cmd *cobra.Command) {
		addFlagString(cmd, ctx, "method")
		addFlagString(cmd, ctx, "path")
		addFlagString(cmd, ctx, "file")
		addFlagString(cmd, ctx, "body")
		addFlagString(cmd, ctx, "query")
		addFlagString(cmd, ctx, "header")
	}, func(args []string) error {
		return runAPI(ctx, append([]string{"request"}, args...))
	}))
	return cmd
}

func newConfigCommand(ctx *context) *cobra.Command {
	cmd := newGroupCommand(ctx, "config", "cmd.config.short", "cmd.config.long", "cmd.config.examples", "cmd.config.notes", "config")
	addHelpLayout(cmd, helpLayoutFourSection)
	addUsageLine(cmd, ctx, "cmd.config.usage_line")
	addUsageNotes(cmd, ctx, "cmd.config.usage_notes")
	addWorkflow(cmd, ctx, "cmd.config.workflow")
	addRelatedCommands(cmd, ctx, "cmd.config.related")
	addNextSteps(cmd, ctx, "cmd.config.next_steps")

	getCmd := newRawLeafCommand(ctx, "get", "cmd.config.get.short", "cmd.config.get.long", "cmd.config.get.examples", "cmd.config.get.notes", func(cmd *cobra.Command) {
		addFlagBool(cmd, ctx, "show-secret")
	}, func(args []string) error {
		return runConfig(ctx, append([]string{"get"}, args...))
	})
	addHelpLayout(getCmd, helpLayoutFourSection)
	addUsageLine(getCmd, ctx, "cmd.config.get.usage_line")
	addUsageNotes(getCmd, ctx, "cmd.config.get.usage_notes")
	addRelatedCommands(getCmd, ctx, "cmd.config.get.related")
	addNextSteps(getCmd, ctx, "cmd.config.get.next_steps")
	cmd.AddCommand(getCmd)

	setCmd := newRawLeafCommand(ctx, "set", "cmd.config.set.short", "cmd.config.set.long", "cmd.config.set.examples", "cmd.config.set.notes", func(cmd *cobra.Command) {
		addFlagString(cmd, ctx, "host")
		addFlagString(cmd, ctx, "username")
		addFlagString(cmd, ctx, "password")
		addFlagString(cmd, ctx, "scene")
		addFlagBool(cmd, ctx, "insecure")
	}, func(args []string) error {
		return runConfig(ctx, append([]string{"set"}, args...))
	})
	addHelpLayout(setCmd, helpLayoutFourSection)
	addUsageLine(setCmd, ctx, "cmd.config.set.usage_line")
	addUsageNotes(setCmd, ctx, "cmd.config.set.usage_notes")
	addMinimumFlags(setCmd, ctx, "cmd.config.set.minimum_flags")
	addAutomaticBehavior(setCmd, ctx, "cmd.config.set.automatic_behavior")
	addWorkflow(setCmd, ctx, "cmd.config.set.workflow")
	addCommonOptionalFlags(setCmd, ctx, "cmd.config.set.common_optional_flags")
	addRelatedCommands(setCmd, ctx, "cmd.config.set.related")
	addNextSteps(setCmd, ctx, "cmd.config.set.next_steps")
	cmd.AddCommand(setCmd)
	return cmd
}

func newTasksCommand(ctx *context) *cobra.Command {
	cmd := newGroupCommand(ctx, "tasks", "cmd.tasks.short", "cmd.tasks.long", "cmd.tasks.examples", "cmd.tasks.notes", "tasks")
	cmd.Hidden = true
	cmd.AddCommand(
		newRawLeafCommand(ctx, "list", "cmd.tasks.list.short", "cmd.tasks.list.long", "cmd.tasks.list.examples", "cmd.tasks.list.notes", func(cmd *cobra.Command) {
			addFlagString(cmd, ctx, "source-id")
			addFlagInt(cmd, ctx, "page")
			addFlagInt(cmd, ctx, "page-size")
		}, func(args []string) error {
			return runTasks(ctx, append([]string{"list"}, args...))
		}),
		newRawLeafCommand(ctx, "steps", "cmd.tasks.steps.short", "cmd.tasks.steps.long", "cmd.tasks.steps.examples", "cmd.tasks.steps.notes", func(cmd *cobra.Command) {
			addFlagString(cmd, ctx, "task-id")
		}, func(args []string) error {
			return runTasks(ctx, append([]string{"steps"}, args...))
		}),
	)
	return cmd
}

func newSourcesCommand(ctx *context) *cobra.Command {
	cmd := newGroupCommand(ctx, "source", "cmd.sources.short", "cmd.sources.long", "cmd.sources.examples", "cmd.sources.notes", "source")
	addHelpLayout(cmd, helpLayoutGroup)
	addUsageLine(cmd, ctx, "cmd.sources.usage_line")
	addUsageNotes(cmd, ctx, "cmd.sources.usage_notes")
	cmd.AddCommand(
		newRawLeafCommand(ctx, "list", "cmd.sources.list.short", "cmd.sources.list.long", "cmd.sources.list.examples", "cmd.sources.list.notes", func(cmd *cobra.Command) {
			addFlagString(cmd, ctx, "type")
			addFlagString(cmd, ctx, "kw")
			addFlagString(cmd, ctx, "binding-status")
			addFlagInt(cmd, ctx, "page")
			addFlagInt(cmd, ctx, "page-size")
		}, func(args []string) error {
			return runSources(ctx, append([]string{"list"}, args...))
		}),
		newRawLeafCommand(ctx, "detail", "cmd.sources.detail.short", "cmd.sources.detail.long", "cmd.sources.detail.examples", "cmd.sources.detail.notes", func(cmd *cobra.Command) {
			addFlagString(cmd, ctx, "id")
			addFlagString(cmd, ctx, "type")
			addFlagString(cmd, ctx, "binding-status")
			addFlagInt(cmd, ctx, "page")
			addFlagInt(cmd, ctx, "page-size")
		}, func(args []string) error {
			return runSources(ctx, append([]string{"detail"}, args...))
		}),
		newRawLeafCommand(ctx, "vms", "cmd.sources.vms.short", "cmd.sources.vms.long", "cmd.sources.vms.examples", "cmd.sources.vms.notes", func(cmd *cobra.Command) {
			addFlagString(cmd, ctx, "connection-type")
			addFlagString(cmd, ctx, "connection-uuid")
			addFlagString(cmd, ctx, "registered")
			addFlagString(cmd, ctx, "kw")
			addFlagInt(cmd, ctx, "page")
			addFlagInt(cmd, ctx, "page-size")
		}, func(args []string) error {
			return runSources(ctx, append([]string{"vms"}, args...))
		}),
		newRawLeafCommand(ctx, "agent-install", "cmd.sources.agent_install.short", "cmd.sources.agent_install.long", "cmd.sources.agent_install.examples", "cmd.sources.agent_install.notes", nil, func(args []string) error {
			return runSources(ctx, append([]string{"agent-install"}, args...))
		}),
		newRawLeafCommand(ctx, "agentless-install", "cmd.sources.agentless_install.short", "cmd.sources.agentless_install.long", "cmd.sources.agentless_install.examples", "cmd.sources.agentless_install.notes", nil, func(args []string) error {
			return runSources(ctx, append([]string{"agentless-install"}, args...))
		}),
		newRawLeafCommand(ctx, "sync-nodes", "cmd.sources.sync_nodes.short", "cmd.sources.sync_nodes.long", "cmd.sources.sync_nodes.examples", "cmd.sources.sync_nodes.notes", func(cmd *cobra.Command) {
			addFlagString(cmd, ctx, "type")
			addFlagString(cmd, ctx, "status")
		}, func(args []string) error {
			return runSources(ctx, append([]string{"sync-nodes"}, args...))
		}),
	)

	createCmd := newRawLeafCommand(ctx, "create", "cmd.sources.create.short", "cmd.sources.create.long", "cmd.sources.create.examples", "cmd.sources.create.notes", func(cmd *cobra.Command) {
		for _, name := range []string{"type", "synch-node-id", "synch-node-ids", "auth-url", "auth-key", "auth-cert", "region-id"} {
			addFlagString(cmd, ctx, name)
		}
		addFlagBool(cmd, ctx, "preview-request")
	}, func(args []string) error {
		return runSources(ctx, append([]string{"create"}, args...))
	})
	cmd.AddCommand(createCmd)
	for _, child := range cmd.Commands() {
		addHelpLayout(child, helpLayoutFourSection)
		addUsageLine(child, ctx, "cmd.sources."+strings.ReplaceAll(child.Name(), "-", "_")+".usage_line")
		addUsageNotes(child, ctx, "cmd.sources."+strings.ReplaceAll(child.Name(), "-", "_")+".usage_notes")
		switch child.Name() {
		case "create", "sync-nodes", "vms":
			addHelpDescription(child, ctx, "cmd.sources."+strings.ReplaceAll(child.Name(), "-", "_")+".help_title")
		}
	}
	return cmd
}

func newTargetOSSCommand(ctx *context) *cobra.Command {
	cmd := newGroupCommand(ctx, "oss", "cmd.target.oss.short", "cmd.target.oss.long", "cmd.target.oss.examples", "cmd.target.oss.notes", "target oss")
	addHelpLayout(cmd, helpLayoutGroup)
	addUsageLine(cmd, ctx, "cmd.target.oss.usage_line")
	addUsageNotes(cmd, ctx, "cmd.target.oss.usage_notes")
	cmd.AddCommand(
		newRawLeafCommand(ctx, "list", "cmd.target.oss.list.short", "cmd.target.oss.list.long", "cmd.target.oss.list.examples", "cmd.target.oss.list.notes", func(cmd *cobra.Command) {
			addFlagInt(cmd, ctx, "page")
			addFlagInt(cmd, ctx, "page-size")
			addFlagString(cmd, ctx, "type")
		}, func(args []string) error {
			return runTargetOSS(ctx, append([]string{"list"}, args...))
		}),
		newRawLeafCommand(ctx, "detail", "cmd.target.oss.detail.short", "cmd.target.oss.detail.long", "cmd.target.oss.detail.examples", "cmd.target.oss.detail.notes", func(cmd *cobra.Command) {
			addFlagString(cmd, ctx, "id")
		}, func(args []string) error {
			return runTargetOSS(ctx, append([]string{"detail"}, args...))
		}),
		newRawLeafCommand(ctx, "wait", "cmd.target.oss.wait.short", "cmd.target.oss.wait.long", "cmd.target.oss.wait.examples", "cmd.target.oss.wait.notes", func(cmd *cobra.Command) {
			addFlagString(cmd, ctx, "id")
			addFlagInt(cmd, ctx, "interval-seconds")
			addFlagInt(cmd, ctx, "timeout-seconds")
		}, func(args []string) error {
			return runTargetOSSWait(ctx, args)
		}),
		newRawLeafCommand(ctx, "buckets", "cmd.target.oss.buckets.short", "cmd.target.oss.buckets.long", "cmd.target.oss.buckets.examples", "cmd.target.oss.buckets.notes", func(cmd *cobra.Command) {
			for _, name := range []string{"auth-url", "region-id", "access-key-id", "access-key-secret", "protocol", "bucket-lookup"} {
				addFlagString(cmd, ctx, name)
			}
			addFlagBool(cmd, ctx, "use-tls")
		}, func(args []string) error {
			return runTargetOSS(ctx, append([]string{"buckets"}, args...))
		}),
		newRawLeafCommand(ctx, "create", "cmd.target.oss.create.short", "cmd.target.oss.create.long", "cmd.target.oss.create.examples", "cmd.target.oss.create.notes", func(cmd *cobra.Command) {
			for _, name := range []string{"file", "display-name", "auth-url", "region-id", "access-key-id", "access-key-secret", "protocol", "bucket-lookup", "bucket-mode", "bucket-name", "public-endpoint", "internal-endpoint", "cloud-type-select", "app-id"} {
				addFlagString(cmd, ctx, name)
			}
			addFlagBool(cmd, ctx, "use-tls")
			addFlagBool(cmd, ctx, "preview-request")
		}, func(args []string) error {
			return runTargetOSS(ctx, append([]string{"create"}, args...))
		}),
		newRawLeafCommand(ctx, "delete", "cmd.target.oss.delete.short", "cmd.target.oss.delete.long", "cmd.target.oss.delete.examples", "cmd.target.oss.delete.notes", func(cmd *cobra.Command) {
			addFlagString(cmd, ctx, "id")
			addFlagBool(cmd, ctx, "force")
		}, func(args []string) error {
			return runTargetOSS(ctx, append([]string{"delete"}, args...))
		}),
	)
	for _, child := range cmd.Commands() {
		addHelpLayout(child, helpLayoutFourSection)
		addUsageLine(child, ctx, "cmd.target.oss."+child.Name()+".usage_line")
		addUsageNotes(child, ctx, "cmd.target.oss."+child.Name()+".usage_notes")
	}
	return cmd
}

func newLicensesCommand(ctx *context) *cobra.Command {
	cmd := newGroupCommand(ctx, "license", "cmd.licenses.short", "cmd.licenses.long", "cmd.licenses.examples", "cmd.licenses.notes", "license")
	addHelpLayout(cmd, helpLayoutGroup)
	addUsageLine(cmd, ctx, "cmd.licenses.usage_line")
	addUsageNotes(cmd, ctx, "cmd.licenses.usage_notes")

	listCmd := newRawLeafCommand(ctx, "list", "cmd.licenses.list.short", "cmd.licenses.list.long", "cmd.licenses.list.examples", "cmd.licenses.list.notes", func(cmd *cobra.Command) {
		addFlagInt(cmd, ctx, "page")
		addFlagInt(cmd, ctx, "page-size")
	}, func(args []string) error {
		return runLicenses(ctx, append([]string{"list"}, args...))
	})
	addHelpLayout(listCmd, helpLayoutFourSection)
	addUsageLine(listCmd, ctx, "cmd.licenses.list.usage_line")
	addUsageNotes(listCmd, ctx, "cmd.licenses.list.usage_notes")

	regCodeCmd := newRawLeafCommand(ctx, "reg-code", "cmd.licenses.reg_code.short", "cmd.licenses.reg_code.long", "cmd.licenses.reg_code.examples", "cmd.licenses.reg_code.notes", nil, func(args []string) error {
		return runLicenses(ctx, append([]string{"reg-code"}, args...))
	})
	addHelpLayout(regCodeCmd, helpLayoutFourSection)
	addUsageLine(regCodeCmd, ctx, "cmd.licenses.reg_code.usage_line")
	addUsageNotes(regCodeCmd, ctx, "cmd.licenses.reg_code.usage_notes")

	activateCmd := newRawLeafCommand(ctx, "activate", "cmd.licenses.activate.short", "cmd.licenses.activate.long", "cmd.licenses.activate.examples", "cmd.licenses.activate.notes", func(cmd *cobra.Command) {
		addFlagString(cmd, ctx, "ddty")
	}, func(args []string) error {
		return runLicenses(ctx, append([]string{"activate"}, args...))
	})
	addHelpLayout(activateCmd, helpLayoutFourSection)
	addUsageLine(activateCmd, ctx, "cmd.licenses.activate.usage_line")
	addUsageNotes(activateCmd, ctx, "cmd.licenses.activate.usage_notes")

	cmd.AddCommand(listCmd, regCodeCmd, activateCmd)
	return cmd
}

func newTargetCommand(ctx *context) *cobra.Command {
	cmd := newGroupCommand(ctx, "target", "cmd.target.short", "cmd.target.long", "cmd.target.examples", "cmd.target.notes", "target")
	addHelpLayout(cmd, helpLayoutGroup)
	addUsageLine(cmd, ctx, "cmd.target.usage_line")
	addUsageNotes(cmd, ctx, "cmd.target.usage_notes")

	supportsCmd := newRawLeafCommand(ctx, "supports", "cmd.target.supports.short", "cmd.target.supports.long", "cmd.target.supports.examples", "cmd.target.supports.notes", nil, func(args []string) error {
		return runTargetSupports(ctx, args)
	})
	addHelpLayout(supportsCmd, helpLayoutFourSection)
	addUsageLine(supportsCmd, ctx, "cmd.target.supports.usage_line")
	addUsageNotes(supportsCmd, ctx, "cmd.target.supports.usage_notes")

	cmd.AddCommand(
		supportsCmd,
		newTargetAccountCommand(ctx),
		newTargetCloudSyncGatewayCommand(ctx),
		newTargetResourceCommand(ctx),
		newTargetOSSCommand(ctx),
	)
	return cmd
}

func newUpgradeCommand(ctx *context) *cobra.Command {
	cmd := newGroupCommand(ctx, "upgrade", "cmd.upgrade.short", "cmd.upgrade.long", "cmd.upgrade.examples", "cmd.upgrade.notes", "upgrade")
	cmd.Hidden = true
	cmd.AddCommand(newRawLeafCommand(ctx, "host", "cmd.upgrade.host.short", "cmd.upgrade.host.long", "cmd.upgrade.host.examples", "cmd.upgrade.host.notes", func(cmd *cobra.Command) {
		addFlagInt(cmd, ctx, "page")
		addFlagInt(cmd, ctx, "page-size")
	}, func(args []string) error {
		return runUpgrade(ctx, append([]string{"host"}, args...))
	}))
	return cmd
}

func addFlagString(cmd *cobra.Command, ctx *context, name string) {
	cmd.Flags().String(name, "", ctx.loc.T("flag."+name))
}

func addFlagInt(cmd *cobra.Command, ctx *context, name string) {
	cmd.Flags().Int(name, 0, ctx.loc.T("flag."+name))
}

func addFlagBool(cmd *cobra.Command, ctx *context, name string) {
	cmd.Flags().Bool(name, false, ctx.loc.T("flag."+name))
}

func configureBuiltinHelpArtifacts(root *cobra.Command, ctx *context) {
	root.InitDefaultCompletionCmd()
	configureBuiltinHelpArtifactsRecursive(root, ctx)
}

func configureBuiltinHelpArtifactsRecursive(cmd *cobra.Command, ctx *context) {
	cmd.InitDefaultHelpFlag()
	if flag := cmd.Flags().Lookup("help"); flag != nil {
		flag.Usage = ctx.loc.T("flag.help")
	}
	if cmd.CommandPath() == "hyperbdrctl completion" {
		cmd.Short = ctx.loc.T("cmd.completion.short")
		cmd.Long = ctx.loc.T("cmd.completion.long")
	}
	for _, child := range cmd.Commands() {
		configureBuiltinHelpArtifactsRecursive(child, ctx)
	}
}
