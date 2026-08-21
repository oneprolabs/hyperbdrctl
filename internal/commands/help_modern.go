package commands

import (
	"fmt"
	"strings"

	"hyperbdr-client/internal/config"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type flagHelpSpec struct {
	name           string
	required       bool
	requiredOnInit bool
	choices        []string
	defaultValue   string
	noteKey        string
}

func renderRootHelp(cmd *cobra.Command, ctx *context) error {
	renderDescription(ctx, cmd)
	fmt.Fprintf(ctx.out, "\n%s: %s\n", ctx.loc.T("help.section_usage"), helpUseLine(cmd))
	if flags := orderedFlagUsages(ctx, cmd, rootFlagSpecs()); flags != "" {
		fmt.Fprintf(ctx.out, "\n%s:\n%s\n", ctx.loc.T("help.section_flags"), flags)
	}
	if subcommands := visibleSubcommands(cmd); len(subcommands) > 0 {
		fmt.Fprintf(ctx.out, "\n%s:\n", ctx.loc.T("help.section_commands"))
		for _, sub := range subcommands {
			fmt.Fprintf(ctx.out, "  %-24s %s\n", sub.Name(), sub.Short)
		}
	}
	if notes := helpUsageNotes(cmd); notes != "" {
		renderHelpSection(ctx, ctx.loc.T("help.section_usage_notes"), notes)
	}
	fmt.Fprintln(ctx.out)
	return nil
}

func renderGroupHelp(cmd *cobra.Command, ctx *context) error {
	renderDescription(ctx, cmd)
	fmt.Fprintf(ctx.out, "\n%s: %s\n", ctx.loc.T("help.section_usage"), helpUseLine(cmd))
	if flags := orderedFlagUsages(ctx, cmd, flagSpecsForCommand(cmd)); flags != "" {
		fmt.Fprintf(ctx.out, "\n%s:\n%s\n", ctx.loc.T("help.section_flags"), flags)
	}
	if subcommands := visibleSubcommands(cmd); len(subcommands) > 0 {
		fmt.Fprintf(ctx.out, "\n%s:\n", ctx.loc.T("help.section_commands"))
		for _, sub := range subcommands {
			fmt.Fprintf(ctx.out, "  %-24s %s\n", sub.Name(), sub.Short)
		}
	}
	if notes := helpUsageNotes(cmd); notes != "" {
		renderHelpSection(ctx, ctx.loc.T("help.section_usage_notes"), notes)
	}
	fmt.Fprintln(ctx.out)
	return nil
}

func renderFourSectionHelp(cmd *cobra.Command, ctx *context) error {
	renderDescription(ctx, cmd)
	fmt.Fprintf(ctx.out, "\n%s: %s\n", ctx.loc.T("help.section_usage"), helpUseLine(cmd))
	if flags := orderedFlagUsages(ctx, cmd, flagSpecsForCommand(cmd)); flags != "" {
		fmt.Fprintf(ctx.out, "\n%s:\n%s\n", ctx.loc.T("help.section_flags"), flags)
	}
	if notes := helpUsageNotes(cmd); notes != "" {
		renderHelpSection(ctx, ctx.loc.T("help.section_usage_notes"), notes)
	}
	fmt.Fprintln(ctx.out)
	return nil
}

func renderDescription(ctx *context, cmd *cobra.Command) {
	description := strings.TrimSpace(cmd.Annotations[helpDescriptionAnnotation])
	if description == "" {
		description = strings.TrimSpace(cmd.Short)
	}
	if description == "" {
		description = strings.TrimSpace(cmd.Long)
	}
	if description == "" {
		return
	}
	fmt.Fprintln(ctx.out, description)
}

func rootFlagSpecs() []flagHelpSpec {
	return []flagHelpSpec{
		{name: "lang"},
		{name: "output"},
		{name: "debug"},
		{name: "version"},
		{name: "help"},
	}
}

func flagSpecsForCommand(cmd *cobra.Command) []flagHelpSpec {
	path := cmd.CommandPath()
	if path == "hyperbdrctl cloud-account create" {
		return omitDynamicParameterHelpFlags(
			cloudAccountCreateFlagSpecsForProfile(cmd.Annotations[cloudAccountCreateHelpProfileAnnotation]),
			dynamicParameterHelpContextFromCommand(cmd),
		)
	}
	if path == "hyperbdrctl cloud-resource fetch" {
		return cloudResourceFetchFlagSpecsForProfile(cmd.Annotations[cloudResourceFetchHelpProfileAnnotation])
	}
	if path == "hyperbdrctl cloud-sync-gateway create" {
		return omitDynamicParameterHelpFlags(
			cloudSyncGatewayCreateFlagSpecsForProfile(cmd.Annotations[cloudSyncGatewayCreateHelpProfileAnnotation]),
			dynamicParameterHelpContextFromCommand(cmd),
		)
	}
	if path == "hyperbdrctl boot-config apply" {
		return bootConfigApplyFlagSpecsForProfile(cmd.Annotations[bootConfigApplyHelpProfileAnnotation])
	}
	if path == "hyperbdrctl oss buckets" || path == "hyperbdrctl oss create" {
		return objectStorageFlagSpecsForProfile(path, cmd.Annotations[objectStorageHelpProfileAnnotation])
	}
	switch {
	case strings.HasPrefix(path, "hyperbdrctl boot-config fetch-block-resources "):
		return bootConfigFetchProviderFlagSpecs()
	case strings.HasPrefix(path, "hyperbdrctl boot-config fetch-oss-resources "):
		return bootConfigFetchProviderFlagSpecs()
	case strings.HasPrefix(path, "hyperbdrctl target resource block "):
		if path == "hyperbdrctl target resource block openstack" {
			return targetResourceOpenStackFlagSpecs()
		}
		return targetResourceDirectAuthFlagSpecs()
	case strings.HasPrefix(path, "hyperbdrctl target resource oss "):
		if path == "hyperbdrctl target resource oss openstack" {
			return targetResourceOpenStackFlagSpecs()
		}
		return targetResourceDirectAuthFlagSpecs()
	case strings.HasPrefix(path, "hyperbdrctl target account fetch-block-resources "):
		if path == "hyperbdrctl target account fetch-block-resources openstack" {
			return []flagHelpSpec{
				{name: "auth-url", required: true},
				{name: "username", required: true},
				{name: "password", required: true},
				{name: "user-domain-id", required: true},
				{name: "fetch-res"},
				{name: "region-id"},
				{name: "project-id"},
				{name: "project-domain-id"},
				{name: "project-name"},
				{name: "compute-zone-id"},
				{name: "block-store-zone-id"},
				{name: "debug"},
				{name: "lang"},
				{name: "output", choices: []string{"table", "json"}, defaultValue: config.DefaultOutput},
				{name: "help"},
			}
		}
		return []flagHelpSpec{
			{name: "cloud-auth-type", choices: []string{"aksk", "password"}},
			{name: "region-id"},
			{name: "boot-mode"},
			{name: "fetch-res"},
			{name: "debug"},
			{name: "lang"},
			{name: "output", choices: []string{"table", "json"}, defaultValue: config.DefaultOutput},
			{name: "help"},
		}
	case strings.HasPrefix(path, "hyperbdrctl target account fetch-oss-resources "):
		if path == "hyperbdrctl target account fetch-oss-resources openstack" {
			return []flagHelpSpec{
				{name: "auth-url", required: true},
				{name: "username", required: true},
				{name: "password", required: true},
				{name: "user-domain-id", required: true},
				{name: "fetch-res"},
				{name: "region-id"},
				{name: "project-id"},
				{name: "project-domain-id"},
				{name: "project-name"},
				{name: "compute-zone-id"},
				{name: "block-store-zone-id"},
				{name: "debug"},
				{name: "lang"},
				{name: "output", choices: []string{"table", "json"}, defaultValue: config.DefaultOutput},
				{name: "help"},
			}
		}
		return []flagHelpSpec{
			{name: "cloud-auth-type", choices: []string{"aksk", "password"}},
			{name: "region-id"},
			{name: "boot-mode"},
			{name: "fetch-res"},
			{name: "debug"},
			{name: "lang"},
			{name: "output", choices: []string{"table", "json"}, defaultValue: config.DefaultOutput},
			{name: "help"},
		}
	case path == "hyperbdrctl target account create-block aliyun":
		return []flagHelpSpec{
			{name: "access-key-id", required: true},
			{name: "access-key-secret", required: true},
			{name: "region-id", required: true},
			{name: "region-name"},
			{name: "account-name"},
			{name: "auth-region-id"},
			{name: "file"},
			{name: "set"},
			{name: "set-json"},
			{name: "preview-request"},
			{name: "debug"},
			{name: "lang"},
			{name: "output", choices: []string{"table", "json"}, defaultValue: config.DefaultOutput},
			{name: "help"},
		}
	case path == "hyperbdrctl target account create-block openstack":
		return []flagHelpSpec{
			{name: "auth-url", required: true},
			{name: "username", required: true},
			{name: "password", required: true},
			{name: "user-domain-id", required: true},
			{name: "project-domain-id", required: true},
			{name: "project-name", required: true},
			{name: "region-name", required: true},
			{name: "ssh-port", defaultValue: "22"},
			{name: "ssh-pass"},
			{name: "linux-hd-username"},
			{name: "linux-hd-password"},
			{name: "linux-hd-port", defaultValue: "10729"},
			{name: "file"},
			{name: "set"},
			{name: "set-json"},
			{name: "preview-request"},
			{name: "debug"},
			{name: "lang"},
			{name: "output", choices: []string{"table", "json"}, defaultValue: config.DefaultOutput},
			{name: "help"},
		}
	case strings.HasPrefix(path, "hyperbdrctl target account create-block "):
		return []flagHelpSpec{
			{name: "cloud-auth-type", choices: []string{"aksk", "password"}},
			{name: "account-name"},
			{name: "file"},
			{name: "set"},
			{name: "set-json"},
			{name: "preview-request"},
			{name: "debug"},
			{name: "lang"},
			{name: "output", choices: []string{"table", "json"}, defaultValue: config.DefaultOutput},
			{name: "help"},
		}
	case path == "hyperbdrctl target account create-oss aliyun":
		return []flagHelpSpec{
			{name: "access-key-id", required: true},
			{name: "access-key-secret", required: true},
			{name: "region-id", required: true},
			{name: "region-name"},
			{name: "custom-name"},
			{name: "file"},
			{name: "set"},
			{name: "set-json"},
			{name: "use-internal-ip"},
			{name: "boot-loader-image-id"},
			{name: "boot-loader-image-name"},
			{name: "boot-loader-flavor-id"},
			{name: "linux-boot-image-id"},
			{name: "windows-boot-image-id"},
			{name: "linux-uefi-boot-image-id"},
			{name: "windows-uefi-boot-image-id"},
			{name: "preview-request"},
			{name: "debug"},
			{name: "lang"},
			{name: "output", choices: []string{"table", "json"}, defaultValue: config.DefaultOutput},
			{name: "help"},
		}
	case path == "hyperbdrctl target account create-oss openstack":
		return []flagHelpSpec{
			{name: "auth-url", required: true},
			{name: "username", required: true},
			{name: "password", required: true},
			{name: "user-domain-id", required: true},
			{name: "file"},
			{name: "set"},
			{name: "set-json"},
			{name: "project-domain-id"},
			{name: "project-id"},
			{name: "project-name"},
			{name: "region-id"},
			{name: "region-name"},
			{name: "boot-loader-image-id"},
			{name: "boot-loader-image-name"},
			{name: "disk-bus-type-id"},
			{name: "disk-bus-type-name"},
			{name: "custom-name"},
			{name: "use-internal-ip"},
			{name: "linux-boot-image-id"},
			{name: "windows-boot-image-id"},
			{name: "preview-request"},
			{name: "debug"},
			{name: "lang"},
			{name: "output", choices: []string{"table", "json"}, defaultValue: config.DefaultOutput},
			{name: "help"},
		}
	case strings.HasPrefix(path, "hyperbdrctl target account create-oss "):
		return []flagHelpSpec{
			{name: "cloud-auth-type", choices: []string{"aksk", "password"}},
			{name: "custom-name"},
			{name: "file"},
			{name: "set"},
			{name: "set-json"},
			{name: "preview-request"},
			{name: "debug"},
			{name: "lang"},
			{name: "output", choices: []string{"table", "json"}, defaultValue: config.DefaultOutput},
			{name: "help"},
		}
	}

	if strings.HasPrefix(path, "hyperbdrctl target cloud-sync-gateway create ") && path != "hyperbdrctl target cloud-sync-gateway create openstack" {
		return []flagHelpSpec{
			{name: "cloud-account-id", required: true},
			{name: "project-id"},
			{name: "region-id"},
			{name: "zone-id"},
			{name: "compute-zone-id"},
			{name: "image-id"},
			{name: "flavor-id"},
			{name: "network-id"},
			{name: "subnet-id"},
			{name: "fixed-ip"},
			{name: "system-disk-type-id"},
			{name: "volume-type-id"},
			{name: "system-disk-size"},
			{name: "block-store-zone-id"},
			{name: "boot-loader-image-id"},
			{name: "boot-loader-flavor-id"},
			{name: "project-domain-id"},
			{name: "boot-types-id", choices: []string{"boot_from_volume", "boot_from_image"}, defaultValue: "boot_from_volume"},
			{name: "volume-proxy-type", defaultValue: "s3"},
			{name: "hg-control-network", choices: []string{"floating_ip_without_proxy", "fixed_ip_without_proxy", "floating_ip_with_proxy", "fixed_ip_with_proxy"}, defaultValue: "floating_ip_without_proxy"},
			{name: "control-nat-ip"},
			{name: "hg-data-network", choices: []string{"floating_ip_without_proxy", "fixed_ip_without_proxy", "floating_ip_with_proxy", "fixed_ip_with_proxy"}, defaultValue: "floating_ip_without_proxy"},
			{name: "data-nat-ip"},
			{name: "bandwidth-size"},
			{name: "hd-control-network", defaultValue: "floating_ip_with_hg_proxy"},
			{name: "preview-request"},
			{name: "debug"},
			{name: "lang"},
			{name: "output", choices: []string{"table", "json"}, defaultValue: config.DefaultOutput},
			{name: "help"},
		}
	}

	switch path {
	case "hyperbdrctl boot-config":
		return []flagHelpSpec{
			{name: "debug"},
			{name: "lang"},
			{name: "output", choices: []string{"table", "json"}, defaultValue: config.DefaultOutput},
			{name: "help"},
		}
	case "hyperbdrctl boot-config get":
		return []flagHelpSpec{
			{name: "id", required: true},
			{name: "debug"},
			{name: "lang"},
			{name: "output", choices: []string{"table", "json"}, defaultValue: config.DefaultOutput},
			{name: "help"},
		}
	case "hyperbdrctl boot-config apply":
		return bootConfigApplyFlagSpecsForProfile("generic")
	case "hyperbdrctl boot-config fetch-block-resources":
		return []flagHelpSpec{
			{name: "debug"},
			{name: "lang"},
			{name: "output", choices: []string{"table", "json"}, defaultValue: config.DefaultOutput},
			{name: "help"},
		}
	case "hyperbdrctl boot-config fetch-oss-resources":
		return []flagHelpSpec{
			{name: "debug"},
			{name: "lang"},
			{name: "output", choices: []string{"table", "json"}, defaultValue: config.DefaultOutput},
			{name: "help"},
		}
	case "hyperbdrctl host":
		return []flagHelpSpec{
			{name: "debug"},
			{name: "lang"},
			{name: "output", choices: []string{"table", "json"}, defaultValue: config.DefaultOutput},
			{name: "help"},
		}
	case "hyperbdrctl host list":
		return []flagHelpSpec{
			{name: "page", defaultValue: "1"},
			{name: "page-size", defaultValue: "10"},
			{name: "status"},
			{name: "boot-status"},
			{name: "kw"},
			{name: "cloud-type"},
			{name: "ids"},
			{name: "macs"},
			{name: "debug"},
			{name: "lang"},
			{name: "output", choices: []string{"table", "json"}, defaultValue: config.DefaultOutput},
			{name: "help"},
		}
	case "hyperbdrctl host detail":
		return []flagHelpSpec{
			{name: "id", required: true},
			{name: "debug"},
			{name: "lang"},
			{name: "output", choices: []string{"table", "json"}, defaultValue: config.DefaultOutput},
			{name: "help"},
		}
	case "hyperbdrctl host snapshots":
		return []flagHelpSpec{
			{name: "id", required: true},
			{name: "status"},
			{name: "sync-detail"},
			{name: "debug"},
			{name: "lang"},
			{name: "output", choices: []string{"table", "json"}, defaultValue: config.DefaultOutput},
			{name: "help"},
		}
	case "hyperbdrctl host sync":
		return []flagHelpSpec{
			{name: "id"},
			{name: "ids"},
			{name: "mode"},
			{name: "transfer-speed"},
			{name: "file"},
			{name: "debug"},
			{name: "lang"},
			{name: "output", choices: []string{"table", "json"}, defaultValue: config.DefaultOutput},
			{name: "help"},
		}
	case "hyperbdrctl host register":
		return []flagHelpSpec{
			{name: "vm-id"},
			{name: "vm-ids"},
			{name: "debug"},
			{name: "lang"},
			{name: "output", choices: []string{"table", "json"}, defaultValue: config.DefaultOutput},
			{name: "help"},
		}
	case "hyperbdrctl host boot":
		return []flagHelpSpec{
			{name: "id"},
			{name: "file"},
			{name: "snapshot-id"},
			{name: "boot-instance-purpose"},
			{name: "cloud-type"},
			{name: "debug"},
			{name: "lang"},
			{name: "output", choices: []string{"table", "json"}, defaultValue: config.DefaultOutput},
			{name: "help"},
		}
	case "hyperbdrctl host clean":
		return []flagHelpSpec{
			{name: "id"},
			{name: "ids"},
			{name: "file"},
			{name: "debug"},
			{name: "lang"},
			{name: "output", choices: []string{"table", "json"}, defaultValue: config.DefaultOutput},
			{name: "help"},
		}
	case "hyperbdrctl host deregister":
		return []flagHelpSpec{
			{name: "id"},
			{name: "ids"},
			{name: "force"},
			{name: "file"},
			{name: "debug"},
			{name: "lang"},
			{name: "output", choices: []string{"table", "json"}, defaultValue: config.DefaultOutput},
			{name: "help"},
		}
	case "hyperbdrctl host wait":
		return []flagHelpSpec{
			{name: "id"},
			{name: "ids"},
			{name: "operation", choices: []string{"sync", "boot", "clean", "deregister"}},
			{name: "interval-seconds", defaultValue: "60"},
			{name: "timeout-seconds", defaultValue: "3600"},
			{name: "include-steps"},
			{name: "debug"},
			{name: "lang"},
			{name: "output", choices: []string{"table", "json"}, defaultValue: config.DefaultOutput},
			{name: "help"},
		}
	case "hyperbdrctl config":
		return []flagHelpSpec{
			{name: "debug"},
			{name: "lang"},
			{name: "output", choices: []string{"table", "json"}, defaultValue: config.DefaultOutput},
			{name: "help"},
		}
	case "hyperbdrctl production-site":
		return []flagHelpSpec{
			{name: "debug"},
			{name: "lang"},
			{name: "output", choices: []string{"table", "json"}, defaultValue: config.DefaultOutput},
			{name: "help"},
		}
	case "hyperbdrctl production-site list":
		return []flagHelpSpec{
			{name: "type", required: true},
			{name: "kw"},
			{name: "binding-status"},
			{name: "page", defaultValue: "1"},
			{name: "page-size", defaultValue: "10"},
			{name: "debug"},
			{name: "lang"},
			{name: "output", choices: []string{"table", "json"}, defaultValue: config.DefaultOutput},
			{name: "help"},
		}
	case "hyperbdrctl production-site detail":
		return []flagHelpSpec{
			{name: "id", required: true},
			{name: "type"},
			{name: "binding-status"},
			{name: "page", defaultValue: "1"},
			{name: "page-size", defaultValue: "10"},
			{name: "debug"},
			{name: "lang"},
			{name: "output", choices: []string{"table", "json"}, defaultValue: config.DefaultOutput},
			{name: "help"},
		}
	case "hyperbdrctl production-site delete":
		return []flagHelpSpec{
			{name: "id", required: true},
			{name: "force"},
			{name: "debug"},
			{name: "lang"},
			{name: "output", choices: []string{"table", "json"}, defaultValue: config.DefaultOutput},
			{name: "help"},
		}
	case "hyperbdrctl production-site vm-list":
		return []flagHelpSpec{
			{name: "connection-type", required: true},
			{name: "connection-uuid"},
			{name: "registered", choices: []string{"0", "1"}},
			{name: "kw"},
			{name: "page", defaultValue: "1"},
			{name: "page-size", defaultValue: "10"},
			{name: "debug"},
			{name: "lang"},
			{name: "output", choices: []string{"table", "json"}, defaultValue: config.DefaultOutput},
			{name: "help"},
		}
	case "hyperbdrctl agent":
		return []flagHelpSpec{
			{name: "debug"},
			{name: "lang"},
			{name: "output", choices: []string{"table", "json"}, defaultValue: config.DefaultOutput},
			{name: "help"},
		}
	case "hyperbdrctl agent install":
		return []flagHelpSpec{
			{name: "debug"},
			{name: "lang"},
			{name: "output", choices: []string{"table", "json"}, defaultValue: config.DefaultOutput},
			{name: "help"},
		}
	case "hyperbdrctl sync-proxy":
		return []flagHelpSpec{
			{name: "debug"},
			{name: "lang"},
			{name: "output", choices: []string{"table", "json"}, defaultValue: config.DefaultOutput},
			{name: "help"},
		}
	case "hyperbdrctl sync-proxy install":
		return []flagHelpSpec{
			{name: "debug"},
			{name: "lang"},
			{name: "output", choices: []string{"table", "json"}, defaultValue: config.DefaultOutput},
			{name: "help"},
		}
	case "hyperbdrctl sync-proxy list":
		return []flagHelpSpec{
			{name: "type", defaultValue: "proxy"},
			{name: "status", defaultValue: "online"},
			{name: "debug"},
			{name: "lang"},
			{name: "output", choices: []string{"table", "json"}, defaultValue: config.DefaultOutput},
			{name: "help"},
		}
	case "hyperbdrctl sync-proxy delete":
		return []flagHelpSpec{
			{name: "id", required: true},
			{name: "debug"},
			{name: "lang"},
			{name: "output", choices: []string{"table", "json"}, defaultValue: config.DefaultOutput},
			{name: "help"},
		}
	case "hyperbdrctl production-site create":
		return []flagHelpSpec{
			{name: "type", required: true, choices: []string{"vmware", "aws"}},
			{name: "synch-node-id"},
			{name: "synch-node-ids"},
			{name: "auth-url"},
			{name: "auth-key"},
			{name: "auth-cert"},
			{name: "region-id"},
			{name: "preview-request"},
			{name: "debug"},
			{name: "lang"},
			{name: "output", choices: []string{"table", "json"}, defaultValue: config.DefaultOutput},
			{name: "help"},
		}
	case "hyperbdrctl license":
		return []flagHelpSpec{
			{name: "debug"},
			{name: "lang"},
			{name: "output", choices: []string{"table", "json"}, defaultValue: config.DefaultOutput},
			{name: "help"},
		}
	case "hyperbdrctl license list":
		return []flagHelpSpec{
			{name: "page", defaultValue: "1"},
			{name: "page-size", defaultValue: "10"},
			{name: "debug"},
			{name: "lang"},
			{name: "output", choices: []string{"table", "json"}, defaultValue: config.DefaultOutput},
			{name: "help"},
		}
	case "hyperbdrctl license reg-code":
		return []flagHelpSpec{
			{name: "debug"},
			{name: "lang"},
			{name: "output", choices: []string{"table", "json"}, defaultValue: config.DefaultOutput},
			{name: "help"},
		}
	case "hyperbdrctl license activate":
		return []flagHelpSpec{
			{name: "ddty", required: true},
			{name: "debug"},
			{name: "lang"},
			{name: "output", choices: []string{"table", "json"}, defaultValue: config.DefaultOutput},
			{name: "help"},
		}
	case "hyperbdrctl cloud-account":
		return []flagHelpSpec{
			{name: "debug"},
			{name: "lang"},
			{name: "output", choices: []string{"table", "json"}, defaultValue: config.DefaultOutput},
			{name: "help"},
		}
	case "hyperbdrctl cloud-account list":
		return []flagHelpSpec{
			{name: "page", defaultValue: "1"},
			{name: "page-size", defaultValue: "100"},
			{name: "storage-type", choices: []string{"block", "object"}},
			{name: "debug"},
			{name: "lang"},
			{name: "output", choices: []string{"table", "json"}, defaultValue: config.DefaultOutput},
			{name: "help"},
		}
	case "hyperbdrctl cloud-account detail":
		return []flagHelpSpec{
			{name: "id", required: true},
			{name: "debug"},
			{name: "lang"},
			{name: "output", choices: []string{"table", "json"}, defaultValue: config.DefaultOutput},
			{name: "help"},
		}
	case "hyperbdrctl cloud-account wait":
		return []flagHelpSpec{
			{name: "id", required: true},
			{name: "interval-seconds", defaultValue: "60"},
			{name: "timeout-seconds", defaultValue: "3600"},
			{name: "debug"},
			{name: "lang"},
			{name: "output", choices: []string{"table", "json"}, defaultValue: config.DefaultOutput},
			{name: "help"},
		}
	case "hyperbdrctl cloud-account delete":
		return []flagHelpSpec{
			{name: "id", required: true},
			{name: "force"},
			{name: "debug"},
			{name: "lang"},
			{name: "output", choices: []string{"table", "json"}, defaultValue: config.DefaultOutput},
			{name: "help"},
		}
	case "hyperbdrctl cloud-resource":
		return []flagHelpSpec{
			{name: "debug"},
			{name: "lang"},
			{name: "output", choices: []string{"table", "json"}, defaultValue: config.DefaultOutput},
			{name: "help"},
		}
	case "hyperbdrctl cloud-resource catalog":
		return []flagHelpSpec{
			{name: "debug"},
			{name: "lang"},
			{name: "output", choices: []string{"table", "json"}, defaultValue: config.DefaultOutput},
			{name: "help"},
		}
	case "hyperbdrctl cloud-sync-gateway":
		return []flagHelpSpec{
			{name: "debug"},
			{name: "lang"},
			{name: "output", choices: []string{"table", "json"}, defaultValue: config.DefaultOutput},
			{name: "help"},
		}
	case "hyperbdrctl cloud-sync-gateway list":
		return []flagHelpSpec{
			{name: "page", defaultValue: "1"},
			{name: "page-size", defaultValue: "100"},
			{name: "type", defaultValue: "HyperGate"},
			{name: "cloud-account-id"},
			{name: "debug"},
			{name: "lang"},
			{name: "output", choices: []string{"table", "json"}, defaultValue: config.DefaultOutput},
			{name: "help"},
		}
	case "hyperbdrctl cloud-sync-gateway detail":
		return []flagHelpSpec{
			{name: "id", required: true},
			{name: "debug"},
			{name: "lang"},
			{name: "output", choices: []string{"table", "json"}, defaultValue: config.DefaultOutput},
			{name: "help"},
		}
	case "hyperbdrctl cloud-sync-gateway delete":
		return []flagHelpSpec{
			{name: "id"},
			{name: "ids"},
			{name: "force"},
			{name: "debug"},
			{name: "lang"},
			{name: "output", choices: []string{"table", "json"}, defaultValue: config.DefaultOutput},
			{name: "help"},
		}
	case "hyperbdrctl cloud-sync-gateway wait":
		return []flagHelpSpec{
			{name: "id", required: true},
			{name: "interval-seconds", defaultValue: "60"},
			{name: "timeout-seconds", defaultValue: "3600"},
			{name: "debug"},
			{name: "lang"},
			{name: "output", choices: []string{"table", "json"}, defaultValue: config.DefaultOutput},
			{name: "help"},
		}
	case "hyperbdrctl target":
		return []flagHelpSpec{
			{name: "debug"},
			{name: "lang"},
			{name: "output", choices: []string{"table", "json"}, defaultValue: config.DefaultOutput},
			{name: "help"},
		}
	case "hyperbdrctl target account":
		return []flagHelpSpec{
			{name: "debug"},
			{name: "lang"},
			{name: "output", choices: []string{"table", "json"}, defaultValue: config.DefaultOutput},
			{name: "help"},
		}
	case "hyperbdrctl target resource":
		return []flagHelpSpec{
			{name: "debug"},
			{name: "lang"},
			{name: "output", choices: []string{"table", "json"}, defaultValue: config.DefaultOutput},
			{name: "help"},
		}
	case "hyperbdrctl target resource block":
		return []flagHelpSpec{
			{name: "debug"},
			{name: "lang"},
			{name: "output", choices: []string{"table", "json"}, defaultValue: config.DefaultOutput},
			{name: "help"},
		}
	case "hyperbdrctl target resource oss":
		return []flagHelpSpec{
			{name: "debug"},
			{name: "lang"},
			{name: "output", choices: []string{"table", "json"}, defaultValue: config.DefaultOutput},
			{name: "help"},
		}
	case "hyperbdrctl target resource fetch":
		return []flagHelpSpec{
			{name: "cloud-account-id", required: true},
			{name: "fetch-res"},
			{name: "region-id"},
			{name: "zone-id"},
			{name: "flavor-id"},
			{name: "debug"},
			{name: "lang"},
			{name: "output", choices: []string{"table", "json"}, defaultValue: config.DefaultOutput},
			{name: "help"},
		}
	case "hyperbdrctl target account list":
		return []flagHelpSpec{
			{name: "page", defaultValue: "1"},
			{name: "page-size", defaultValue: "100"},
			{name: "storage-type", choices: []string{"block", "object"}},
			{name: "debug"},
			{name: "lang"},
			{name: "output", choices: []string{"table", "json"}, defaultValue: config.DefaultOutput},
			{name: "help"},
		}
	case "hyperbdrctl target account detail":
		return []flagHelpSpec{
			{name: "id", required: true},
			{name: "debug"},
			{name: "lang"},
			{name: "output", choices: []string{"table", "json"}, defaultValue: config.DefaultOutput},
			{name: "help"},
		}
	case "hyperbdrctl target account wait":
		return []flagHelpSpec{
			{name: "id", required: true},
			{name: "interval-seconds", defaultValue: "60"},
			{name: "timeout-seconds", defaultValue: "3600"},
			{name: "debug"},
			{name: "lang"},
			{name: "output", choices: []string{"table", "json"}, defaultValue: config.DefaultOutput},
			{name: "help"},
		}
	case "hyperbdrctl target account fetch-block-resources":
		return []flagHelpSpec{
			{name: "debug"},
			{name: "lang"},
			{name: "output", choices: []string{"table", "json"}, defaultValue: config.DefaultOutput},
			{name: "help"},
		}
	case "hyperbdrctl target account fetch-oss-resources":
		return []flagHelpSpec{
			{name: "debug"},
			{name: "lang"},
			{name: "output", choices: []string{"table", "json"}, defaultValue: config.DefaultOutput},
			{name: "help"},
		}
	case "hyperbdrctl target account create":
		return []flagHelpSpec{
			{name: "file"},
			{name: "body"},
			{name: "preview-request"},
			{name: "debug"},
			{name: "lang"},
			{name: "output", choices: []string{"table", "json"}, defaultValue: config.DefaultOutput},
			{name: "help"},
		}
	case "hyperbdrctl target account create-block":
		return []flagHelpSpec{
			{name: "debug"},
			{name: "lang"},
			{name: "output", choices: []string{"table", "json"}, defaultValue: config.DefaultOutput},
			{name: "help"},
		}
	case "hyperbdrctl target account create-oss":
		return []flagHelpSpec{
			{name: "debug"},
			{name: "lang"},
			{name: "output", choices: []string{"table", "json"}, defaultValue: config.DefaultOutput},
			{name: "help"},
		}
	case "hyperbdrctl target account delete":
		return []flagHelpSpec{
			{name: "id", required: true},
			{name: "storage-type"},
			{name: "force", defaultValue: "false"},
			{name: "debug"},
			{name: "lang"},
			{name: "output", choices: []string{"table", "json"}, defaultValue: config.DefaultOutput},
			{name: "help"},
		}
	case "hyperbdrctl oss":
		return []flagHelpSpec{
			{name: "debug"},
			{name: "lang"},
			{name: "output", choices: []string{"table", "json"}, defaultValue: config.DefaultOutput},
			{name: "help"},
		}
	case "hyperbdrctl oss list":
		return []flagHelpSpec{
			{name: "page", defaultValue: "1"},
			{name: "page-size", defaultValue: "100"},
			{name: "debug"},
			{name: "lang"},
			{name: "output", choices: []string{"table", "json"}, defaultValue: config.DefaultOutput},
			{name: "help"},
		}
	case "hyperbdrctl oss detail":
		return []flagHelpSpec{
			{name: "id", required: true},
			{name: "debug"},
			{name: "lang"},
			{name: "output", choices: []string{"table", "json"}, defaultValue: config.DefaultOutput},
			{name: "help"},
		}
	case "hyperbdrctl oss wait":
		return []flagHelpSpec{
			{name: "id", required: true},
			{name: "interval-seconds", defaultValue: "60"},
			{name: "timeout-seconds", defaultValue: "3600"},
			{name: "debug"},
			{name: "lang"},
			{name: "output", choices: []string{"table", "json"}, defaultValue: config.DefaultOutput},
			{name: "help"},
		}
	case "hyperbdrctl oss catalog":
		return []flagHelpSpec{
			{name: "provider"},
			{name: "debug"},
			{name: "lang"},
			{name: "output", choices: []string{"table", "json"}, defaultValue: config.DefaultOutput},
			{name: "help"},
		}
	case "hyperbdrctl oss delete":
		return []flagHelpSpec{
			{name: "id", required: true},
			{name: "force"},
			{name: "debug"},
			{name: "lang"},
			{name: "output", choices: []string{"table", "json"}, defaultValue: config.DefaultOutput},
			{name: "help"},
		}
	case "hyperbdrctl target cloud-sync-gateway":
		return []flagHelpSpec{
			{name: "debug"},
			{name: "lang"},
			{name: "output", choices: []string{"table", "json"}, defaultValue: config.DefaultOutput},
			{name: "help"},
		}
	case "hyperbdrctl target cloud-sync-gateway list":
		return []flagHelpSpec{
			{name: "page", defaultValue: "1"},
			{name: "page-size", defaultValue: "100"},
			{name: "type", defaultValue: "HyperGate"},
			{name: "cloud-account-id"},
			{name: "debug"},
			{name: "lang"},
			{name: "output", choices: []string{"table", "json"}, defaultValue: config.DefaultOutput},
			{name: "help"},
		}
	case "hyperbdrctl target cloud-sync-gateway detail":
		return []flagHelpSpec{
			{name: "id", required: true},
			{name: "debug"},
			{name: "lang"},
			{name: "output", choices: []string{"table", "json"}, defaultValue: config.DefaultOutput},
			{name: "help"},
		}
	case "hyperbdrctl target cloud-sync-gateway delete":
		return []flagHelpSpec{
			{name: "id"},
			{name: "ids"},
			{name: "force", defaultValue: "false"},
			{name: "debug"},
			{name: "lang"},
			{name: "output", choices: []string{"table", "json"}, defaultValue: config.DefaultOutput},
			{name: "help"},
		}
	case "hyperbdrctl target cloud-sync-gateway wait":
		return []flagHelpSpec{
			{name: "id", required: true},
			{name: "interval-seconds", defaultValue: "60"},
			{name: "timeout-seconds", defaultValue: "3600"},
			{name: "debug"},
			{name: "lang"},
			{name: "output", choices: []string{"table", "json"}, defaultValue: config.DefaultOutput},
			{name: "help"},
		}
	case "hyperbdrctl target cloud-sync-gateway resources":
		return []flagHelpSpec{
			{name: "cloud-account-id", required: true},
			{name: "fetch-res"},
			{name: "region-id"},
			{name: "zone-id"},
			{name: "flavor-id"},
			{name: "flavor-vcpus"},
			{name: "flavor-ram"},
			{name: "purpose", defaultValue: "make_hg"},
			{name: "image-type"},
			{name: "debug"},
			{name: "lang"},
			{name: "output", choices: []string{"table", "json"}, defaultValue: config.DefaultOutput},
			{name: "help"},
		}
	case "hyperbdrctl target cloud-sync-gateway subnet-config":
		return []flagHelpSpec{
			{name: "cloud-account-id", required: true},
			{name: "cloud-type"},
			{name: "region-id"},
			{name: "zone-id", required: true},
			{name: "network-id", required: true},
			{name: "subnet-id"},
			{name: "debug"},
			{name: "lang"},
			{name: "output", choices: []string{"table", "json"}, defaultValue: config.DefaultOutput},
			{name: "help"},
		}
	case "hyperbdrctl target cloud-sync-gateway create":
		return []flagHelpSpec{
			{name: "debug"},
			{name: "lang"},
			{name: "output", choices: []string{"table", "json"}, defaultValue: config.DefaultOutput},
			{name: "help"},
		}
	case "hyperbdrctl target cloud-sync-gateway create openstack":
		return []flagHelpSpec{
			{name: "cloud-account-id", required: true},
			{name: "boot-loader-image-id", required: true},
			{name: "project-id"},
			{name: "region-id"},
			{name: "compute-zone-id"},
			{name: "image-id"},
			{name: "flavor-id"},
			{name: "network-id"},
			{name: "subnet-id"},
			{name: "fixed-ip"},
			{name: "volume-type-id"},
			{name: "system-disk-size", defaultValue: "50"},
			{name: "block-store-zone-id"},
			{name: "boot-loader-flavor-id"},
			{name: "project-domain-id"},
			{name: "boot-types-id", choices: []string{"boot_from_volume", "boot_from_image"}, defaultValue: "boot_from_volume"},
			{name: "volume-proxy-type", defaultValue: "s3"},
			{name: "hg-control-network", choices: []string{"floating_ip_without_proxy", "fixed_ip_without_proxy", "floating_ip_with_proxy", "fixed_ip_with_proxy"}, defaultValue: "floating_ip_without_proxy"},
			{name: "control-nat-ip"},
			{name: "hg-data-network", choices: []string{"floating_ip_without_proxy", "fixed_ip_without_proxy", "floating_ip_with_proxy", "fixed_ip_with_proxy"}, defaultValue: "floating_ip_without_proxy"},
			{name: "data-nat-ip"},
			{name: "preview-request"},
			{name: "debug"},
			{name: "lang"},
			{name: "output", choices: []string{"table", "json"}, defaultValue: config.DefaultOutput},
			{name: "help"},
		}
	case "hyperbdrctl config get":
		return []flagHelpSpec{
			{name: "show-secret"},
			{name: "debug"},
			{name: "lang"},
			{name: "output"},
			{name: "help"},
		}
	case "hyperbdrctl config set":
		return []flagHelpSpec{
			{name: "host", requiredOnInit: true},
			{name: "username", requiredOnInit: true},
			{name: "password", requiredOnInit: true},
			{name: "scene", choices: []string{"dr", "migration"}, defaultValue: config.DefaultScene},
			{name: "insecure"},
			{name: "debug"},
			{name: "lang", choices: []string{"en", "zh_cn"}, defaultValue: config.DefaultLang},
			{name: "output", choices: []string{"table", "json"}, defaultValue: config.DefaultOutput},
			{name: "help"},
		}
	default:
		return nil
	}
}

func objectStorageFlagSpecsForProfile(path, profile string) []flagHelpSpec {
	common := []flagHelpSpec{
		{name: "debug"},
		{name: "lang"},
		{name: "output", choices: []string{"table", "json"}, defaultValue: config.DefaultOutput},
		{name: "help"},
	}
	if path == "hyperbdrctl oss buckets" {
		switch profile {
		case "provider":
			return append([]flagHelpSpec{
				{name: "provider", required: true},
				{name: "region-id", required: true},
				{name: "access-key-id", required: true},
				{name: "access-key-secret", required: true},
				{name: "auth-url"},
				{name: "protocol", defaultValue: "s3"},
				{name: "bucket-lookup", defaultValue: "dns"},
				{name: "use-tls", defaultValue: "true"},
			}, common...)
		case "custom":
			return append([]flagHelpSpec{
				{name: "provider"},
				{name: "auth-url", required: true},
				{name: "region-id"},
				{name: "access-key-id", required: true},
				{name: "access-key-secret", required: true},
				{name: "protocol", defaultValue: "s3"},
				{name: "bucket-lookup", choices: []string{"dns", "path"}, defaultValue: "dns"},
				{name: "use-tls"},
			}, common...)
		default:
			return append([]flagHelpSpec{
				{name: "provider"},
				{name: "auth-url"},
				{name: "region-id"},
				{name: "access-key-id", required: true},
				{name: "access-key-secret", required: true},
				{name: "protocol", defaultValue: "s3"},
				{name: "bucket-lookup", defaultValue: "dns"},
				{name: "use-tls", defaultValue: "true"},
			}, common...)
		}
	}
	switch profile {
	case "provider":
		return append([]flagHelpSpec{
			{name: "display-name"},
			{name: "provider", required: true},
			{name: "region-id", required: true},
			{name: "access-key-id", required: true},
			{name: "access-key-secret", required: true},
			{name: "auth-url"},
			{name: "protocol"},
			{name: "bucket-lookup"},
			{name: "use-tls", defaultValue: "true"},
			{name: "bucket-mode", choices: []string{"existing", "new"}, defaultValue: "existing"},
			{name: "bucket-name", required: true},
			{name: "public-endpoint"},
			{name: "internal-endpoint"},
			{name: "app-id"},
			{name: "preview-request"},
		}, common...)
	case "custom":
		return append([]flagHelpSpec{
			{name: "display-name"},
			{name: "provider"},
			{name: "auth-url", required: true},
			{name: "region-id"},
			{name: "access-key-id", required: true},
			{name: "access-key-secret", required: true},
			{name: "protocol"},
			{name: "bucket-lookup", choices: []string{"dns", "path"}},
			{name: "use-tls"},
			{name: "bucket-mode", choices: []string{"existing", "new"}, defaultValue: "existing"},
			{name: "bucket-name", required: true},
			{name: "public-endpoint"},
			{name: "internal-endpoint"},
			{name: "app-id"},
			{name: "preview-request"},
		}, common...)
	default:
		return append([]flagHelpSpec{
			{name: "display-name"},
			{name: "provider"},
			{name: "auth-url"},
			{name: "region-id"},
			{name: "access-key-id", required: true},
			{name: "access-key-secret", required: true},
			{name: "protocol"},
			{name: "bucket-lookup"},
			{name: "use-tls", defaultValue: "true"},
			{name: "bucket-mode", choices: []string{"existing", "new"}, defaultValue: "existing"},
			{name: "bucket-name", required: true},
			{name: "public-endpoint"},
			{name: "internal-endpoint"},
			{name: "app-id"},
			{name: "preview-request"},
		}, common...)
	}
}

func cloudSyncGatewayCreateFlagSpecsForProfile(profile string) []flagHelpSpec {
	commonGlobal := []flagHelpSpec{
		{name: "debug"},
		{name: "lang"},
		{name: "output", choices: []string{"table", "json"}, defaultValue: config.DefaultOutput},
		{name: "help"},
	}
	commonCommand := []flagHelpSpec{
		{name: "set"},
		{name: "set-json"},
		{name: "preview-request"},
	}
	withCommonCommand := func(specs []flagHelpSpec) []flagHelpSpec {
		return append(append(specs, commonCommand...), commonGlobal...)
	}
	genericProvider := []flagHelpSpec{
		{name: "cloud-type", required: true},
		{name: "cloud-account-id", required: true},
		{name: "project-id"},
		{name: "region-id"},
		{name: "zone-id"},
		{name: "compute-zone-id"},
		{name: "image-id"},
		{name: "flavor-id"},
		{name: "network-id"},
		{name: "subnet-id"},
		{name: "fixed-ip"},
		{name: "system-disk-type-id"},
		{name: "volume-type-id"},
		{name: "system-disk-size"},
		{name: "block-store-zone-id"},
		{name: "boot-loader-image-id"},
		{name: "boot-loader-flavor-id"},
		{name: "project-domain-id"},
		{name: "boot-types-id", choices: []string{"boot_from_volume", "boot_from_image"}, defaultValue: "boot_from_volume"},
		{name: "volume-proxy-type", defaultValue: "s3"},
		{name: "hg-control-network", choices: []string{"floating_ip_without_proxy", "fixed_ip_without_proxy", "floating_ip_with_proxy", "fixed_ip_with_proxy"}, defaultValue: "floating_ip_without_proxy"},
		{name: "control-nat-ip"},
		{name: "hg-data-network", choices: []string{"floating_ip_without_proxy", "fixed_ip_without_proxy", "floating_ip_with_proxy", "fixed_ip_with_proxy"}, defaultValue: "floating_ip_without_proxy"},
		{name: "data-nat-ip"},
		{name: "bandwidth-size"},
		{name: "hd-control-network", defaultValue: "floating_ip_with_hg_proxy"},
	}
	aliyun := []flagHelpSpec{
		{name: "cloud-type", required: true},
		{name: "cloud-account-id", required: true},
		{name: "zone-id", required: true},
		{name: "image-id", required: true},
		{name: "flavor-id", required: true},
		{name: "network-id", required: true},
		{name: "subnet-id", required: true},
		{name: "fixed-ip"},
		{name: "system-disk-type-id", required: true},
		{name: "system-disk-size", defaultValue: "40"},
		{name: "boot-loader-image-id"},
		{name: "volume-proxy-type", defaultValue: "s3"},
		{name: "hg-control-network", choices: []string{"floating_ip_without_proxy", "fixed_ip_without_proxy", "floating_ip_with_proxy", "fixed_ip_with_proxy"}, defaultValue: "floating_ip_without_proxy"},
		{name: "control-nat-ip"},
		{name: "hg-data-network", choices: []string{"floating_ip_without_proxy", "fixed_ip_without_proxy", "floating_ip_with_proxy", "fixed_ip_with_proxy"}, defaultValue: "floating_ip_without_proxy"},
		{name: "data-nat-ip"},
		{name: "bandwidth-size"},
		{name: "hd-control-network", defaultValue: "floating_ip_with_hg_proxy"},
	}
	openstack := []flagHelpSpec{
		{name: "cloud-type", required: true},
		{name: "cloud-account-id", required: true},
		{name: "boot-loader-image-id"},
		{name: "project-id"},
		{name: "region-id"},
		{name: "compute-zone-id"},
		{name: "image-id"},
		{name: "flavor-id"},
		{name: "network-id"},
		{name: "subnet-id"},
		{name: "fixed-ip"},
		{name: "volume-type-id"},
		{name: "system-disk-size", defaultValue: "50"},
		{name: "block-store-zone-id"},
		{name: "boot-loader-flavor-id"},
		{name: "project-domain-id"},
		{name: "boot-types-id", choices: []string{"boot_from_volume", "boot_from_image"}, defaultValue: "boot_from_volume"},
		{name: "volume-proxy-type", defaultValue: "s3"},
		{name: "hg-control-network", choices: []string{"floating_ip_without_proxy", "fixed_ip_without_proxy", "floating_ip_with_proxy", "fixed_ip_with_proxy"}, defaultValue: "floating_ip_without_proxy"},
		{name: "control-nat-ip"},
		{name: "hg-data-network", choices: []string{"floating_ip_without_proxy", "fixed_ip_without_proxy", "floating_ip_with_proxy", "fixed_ip_with_proxy"}, defaultValue: "floating_ip_without_proxy"},
		{name: "data-nat-ip"},
	}
	accountScoped := func(specs []flagHelpSpec) []flagHelpSpec {
		result := make([]flagHelpSpec, 0, len(specs))
		for _, spec := range specs {
			if spec.name == "cloud-type" {
				continue
			}
			if spec.name == "boot-loader-image-id" {
				spec.required = false
			}
			result = append(result, spec)
		}
		return withCommonCommand(result)
	}
	switch profile {
	case "aliyun":
		return withCommonCommand(aliyun)
	case "aliyun-account":
		return withCommonCommand([]flagHelpSpec{
			{name: "cloud-account-id", required: true},
			{name: "zone-id", required: true},
			{name: "image-id", required: true},
			{name: "flavor-id", required: true},
			{name: "network-id", required: true},
			{name: "subnet-id", required: true},
			{name: "system-disk-type-id", required: true},
			{name: "system-disk-size", defaultValue: "40"},
			{name: "fixed-ip"},
			{name: "boot-loader-image-id"},
			{name: "volume-proxy-type", defaultValue: "s3"},
			{name: "hg-control-network", choices: []string{"floating_ip_without_proxy", "fixed_ip_without_proxy", "floating_ip_with_proxy", "fixed_ip_with_proxy"}, defaultValue: "floating_ip_without_proxy"},
			{name: "control-nat-ip"},
			{name: "hg-data-network", choices: []string{"floating_ip_without_proxy", "fixed_ip_without_proxy", "floating_ip_with_proxy", "fixed_ip_with_proxy"}, defaultValue: "floating_ip_without_proxy"},
			{name: "data-nat-ip"},
			{name: "bandwidth-size"},
			{name: "hd-control-network", defaultValue: "floating_ip_with_hg_proxy"},
		})
	case "huawei-account":
		return withCommonCommand([]flagHelpSpec{
			{name: "cloud-account-id", required: true},
			{name: "zone-id", required: true},
			{name: "image-id", required: true},
			{name: "flavor-id", required: true},
			{name: "network-id", required: true},
			{name: "subnet-id", required: true},
			{name: "system-disk-type-id", required: true},
			{name: "system-disk-size", defaultValue: "40"},
			{name: "fixed-ip"},
			{name: "volume-proxy-type", defaultValue: "s3"},
			{name: "hg-control-network", choices: []string{"floating_ip_without_proxy", "fixed_ip_without_proxy", "floating_ip_with_proxy", "fixed_ip_with_proxy"}, defaultValue: "floating_ip_without_proxy"},
			{name: "control-nat-ip"},
			{name: "hg-data-network", choices: []string{"floating_ip_without_proxy", "fixed_ip_without_proxy", "floating_ip_with_proxy", "fixed_ip_with_proxy"}, defaultValue: "floating_ip_without_proxy"},
			{name: "data-nat-ip"},
			{name: "bandwidth-size"},
			{name: "hd-control-network", defaultValue: "floating_ip_with_hg_proxy"},
		})
	case "provider-account":
		return accountScoped(genericProvider)
	case "openstack":
		return withCommonCommand(openstack)
	case "openstack-account":
		return accountScoped(openstack)
	case "generic":
		return withCommonCommand([]flagHelpSpec{
			{name: "cloud-account-id"},
		})
	default:
		return withCommonCommand(genericProvider)
	}
}

func cloudResourceFetchFlagSpecsForProfile(profile string) []flagHelpSpec {
	commonGlobal := []flagHelpSpec{
		{name: "debug"},
		{name: "lang"},
		{name: "output", choices: []string{"table", "json"}, defaultValue: config.DefaultOutput},
		{name: "help"},
	}
	accountFlags := []flagHelpSpec{
		{name: "cloud-account-id", required: true},
		{name: "fetch-res"},
		{name: "region-id"},
		{name: "zone-id"},
		{name: "flavor-id"},
		{name: "flavor-vcpus"},
		{name: "flavor-ram"},
		{name: "network-id"},
	}
	openStackContext := []flagHelpSpec{
		{name: "project-id"},
		{name: "project-domain-id"},
		{name: "project-name"},
		{name: "compute-zone-id"},
		{name: "block-store-zone-id"},
		{name: "os-type"},
	}
	directAK := []flagHelpSpec{
		{name: "cloud-type", required: true},
		{name: "storage-type", required: true, choices: []string{"block", "object"}},
		{name: "access-key-id", required: true},
		{name: "access-key-secret", required: true},
		{name: "fetch-res"},
		{name: "region-id"},
		{name: "zone-id"},
		{name: "flavor-id"},
		{name: "flavor-vcpus"},
		{name: "flavor-ram"},
		{name: "network-id"},
		{name: "boot-mode"},
	}
	directOpenStack := []flagHelpSpec{
		{name: "cloud-type", required: true},
		{name: "storage-type", required: true, choices: []string{"block", "object"}},
		{name: "auth-url", required: true},
		{name: "username", required: true},
		{name: "password", required: true},
		{name: "user-domain-id", required: true},
		{name: "fetch-res"},
		{name: "region-id"},
		{name: "zone-id"},
		{name: "flavor-id"},
		{name: "flavor-vcpus"},
		{name: "flavor-ram"},
		{name: "network-id"},
		{name: "project-id"},
		{name: "project-domain-id"},
		{name: "project-name"},
		{name: "compute-zone-id"},
		{name: "block-store-zone-id"},
		{name: "os-type"},
	}
	switch {
	case strings.HasPrefix(profile, "account|openstack|"):
		return append(append(accountFlags, openStackContext...), commonGlobal...)
	case strings.HasPrefix(profile, "account|"):
		return append(accountFlags, commonGlobal...)
	case profile == "storage|block" || profile == "storage|object":
		return append([]flagHelpSpec{
			{name: "cloud-type"},
			{name: "storage-type", required: true, choices: []string{"block", "object"}},
		}, commonGlobal...)
	case strings.HasPrefix(profile, "direct|openstack|"):
		return append(directOpenStack, commonGlobal...)
	case strings.HasPrefix(profile, "direct|"):
		return append(directAK, commonGlobal...)
	default:
		return append([]flagHelpSpec{
			{name: "cloud-account-id"},
			{name: "cloud-type"},
			{name: "storage-type", choices: []string{"block", "object"}},
		}, commonGlobal...)
	}
}

func bootConfigApplyFlagSpecsForProfile(profile string) []flagHelpSpec {
	commonGlobal := []flagHelpSpec{
		{name: "debug"},
		{name: "lang"},
		{name: "output", choices: []string{"table", "json"}, defaultValue: config.DefaultOutput},
		{name: "help"},
	}
	commonOverrides := []flagHelpSpec{
		{name: "file"},
		{name: "set"},
		{name: "set-json"},
		{name: "preview-request"},
	}
	accountBase := []flagHelpSpec{
		{name: "id", required: true},
		{name: "cloud-account-id", required: true},
		{name: "storage-id"},
		{name: "region-id"},
		{name: "zone-id"},
		{name: "flavor-id"},
		{name: "network-id"},
		{name: "subnet-id"},
		{name: "security-group-id"},
	}
	switch profile {
	case "account|aliyun|block":
		return append(append([]flagHelpSpec{
			{name: "id", required: true},
			{name: "cloud-account-id", required: true},
			{name: "storage-id", required: true},
			{name: "zone-id", required: true},
			{name: "flavor-id", required: true},
			{name: "volume-type-id", required: true},
			{name: "network-id", required: true},
			{name: "subnet-id", required: true},
			{name: "security-group-id", required: true},
			{name: "region-id"},
		}, commonOverrides...), commonGlobal...)
	case "account|block":
		return append(append(append(accountBase, []flagHelpSpec{
			{name: "volume-type-id"},
		}...), commonOverrides...), commonGlobal...)
	case "account|object":
		return append(append(append(accountBase, []flagHelpSpec{
			{name: "system-volume-type-id"},
			{name: "volume-type-id"},
			{name: "boot-loader-image-id"},
			{name: "boot-loader-flavor-id"},
		}...), commonOverrides...), commonGlobal...)
	case "account|openstack|block":
		return append(append(append(accountBase, []flagHelpSpec{
			{name: "project-id"},
			{name: "project-domain-id"},
			{name: "compute-zone-id"},
			{name: "image-id"},
			{name: "volume-type-id"},
		}...), commonOverrides...), commonGlobal...)
	case "account|openstack|object":
		return append(append(append(accountBase, []flagHelpSpec{
			{name: "project-id"},
			{name: "project-domain-id"},
			{name: "compute-zone-id"},
			{name: "image-id"},
			{name: "system-volume-type-id"},
			{name: "volume-type-id"},
			{name: "boot-loader-image-id"},
			{name: "boot-loader-flavor-id"},
		}...), commonOverrides...), commonGlobal...)
	default:
		return append([]flagHelpSpec{
			{name: "id", required: true},
			{name: "cloud-account-id"},
			{name: "file"},
			{name: "set"},
			{name: "set-json"},
			{name: "preview-request"},
		}, commonGlobal...)
	}
}

func cloudAccountCreateFlagSpecsForProfile(profile string) (specs []flagHelpSpec) {
	defer func() {
		if strings.Contains(profile, "|") {
			specs = omitCloudAccountCreateSelectionFlags(specs)
		}
	}()

	switch profile {
	case "block|aliyun":
		return []flagHelpSpec{
			{name: "cloud-type", required: true},
			{name: "storage-type", required: true, choices: []string{"block", "object"}},
			{name: "access-key-id", required: true},
			{name: "access-key-secret", required: true},
			{name: "region-id", required: true},
			{name: "region-name"},
			{name: "account-name"},
			{name: "auth-region-id"},
			{name: "file"},
			{name: "set"},
			{name: "set-json"},
			{name: "preview-request"},
			{name: "debug"},
			{name: "lang"},
			{name: "output", choices: []string{"table", "json"}, defaultValue: config.DefaultOutput},
			{name: "help"},
		}
	case "block|openstack":
		return []flagHelpSpec{
			{name: "cloud-type", required: true},
			{name: "storage-type", required: true, choices: []string{"block", "object"}},
			{name: "auth-url", required: true},
			{name: "username", required: true},
			{name: "password", required: true},
			{name: "user-domain-id", required: true},
			{name: "project-domain-id", required: true},
			{name: "project-name", required: true},
			{name: "region-name", required: true},
			{name: "ssh-port", defaultValue: "22"},
			{name: "ssh-pass"},
			{name: "linux-hd-username"},
			{name: "linux-hd-password"},
			{name: "linux-hd-port", defaultValue: "10729"},
			{name: "file"},
			{name: "set"},
			{name: "set-json"},
			{name: "preview-request"},
			{name: "debug"},
			{name: "lang"},
			{name: "output", choices: []string{"table", "json"}, defaultValue: config.DefaultOutput},
			{name: "help"},
		}
	case "block|huawei":
		return []flagHelpSpec{
			{name: "cloud-type", required: true},
			{name: "storage-type", required: true, choices: []string{"block", "object"}},
			{name: "access-key-id", required: true},
			{name: "access-key-secret", required: true},
			{name: "region-id", required: true},
			{name: "region-name"},
			{name: "account-name"},
			{name: "file"},
			{name: "set"},
			{name: "set-json"},
			{name: "preview-request"},
			{name: "debug"},
			{name: "lang"},
			{name: "output", choices: []string{"table", "json"}, defaultValue: config.DefaultOutput},
			{name: "help"},
		}
	case "objectstorage|aliyun":
		return []flagHelpSpec{
			{name: "cloud-type", required: true},
			{name: "storage-type", required: true, choices: []string{"block", "object"}},
			{name: "access-key-id", required: true},
			{name: "access-key-secret", required: true},
			{name: "region-id", required: true},
			{name: "region-name"},
			{name: "custom-name"},
			{name: "file"},
			{name: "set"},
			{name: "set-json"},
			{name: "use-internal-ip", choices: []string{"0", "1"}},
			{name: "boot-loader-image-id"},
			{name: "boot-loader-image-name"},
			{name: "boot-loader-flavor-id"},
			{name: "linux-boot-image-id"},
			{name: "windows-boot-image-id"},
			{name: "linux-uefi-boot-image-id"},
			{name: "windows-uefi-boot-image-id"},
			{name: "preview-request"},
			{name: "debug"},
			{name: "lang"},
			{name: "output", choices: []string{"table", "json"}, defaultValue: config.DefaultOutput},
			{name: "help"},
		}
	case "objectstorage|openstack":
		return []flagHelpSpec{
			{name: "cloud-type", required: true},
			{name: "storage-type", required: true, choices: []string{"block", "object"}},
			{name: "auth-url", required: true},
			{name: "username", required: true},
			{name: "password", required: true},
			{name: "user-domain-id", required: true},
			{name: "file"},
			{name: "set"},
			{name: "set-json"},
			{name: "project-domain-id"},
			{name: "project-id"},
			{name: "project-name"},
			{name: "region-id"},
			{name: "region-name"},
			{name: "boot-loader-image-id"},
			{name: "boot-loader-image-name"},
			{name: "disk-bus-type-id"},
			{name: "disk-bus-type-name"},
			{name: "custom-name"},
			{name: "use-internal-ip", choices: []string{"0", "1"}},
			{name: "linux-boot-image-id"},
			{name: "windows-boot-image-id"},
			{name: "preview-request"},
			{name: "debug"},
			{name: "lang"},
			{name: "output", choices: []string{"table", "json"}, defaultValue: config.DefaultOutput},
			{name: "help"},
		}
	case "objectstorage|huawei":
		return []flagHelpSpec{
			{name: "cloud-type", required: true},
			{name: "storage-type", required: true, choices: []string{"block", "object"}},
			{name: "access-key-id", required: true},
			{name: "access-key-secret", required: true},
			{name: "region-id", required: true},
			{name: "custom-name"},
			{name: "file"},
			{name: "set"},
			{name: "set-json"},
			{name: "preview-request"},
			{name: "debug"},
			{name: "lang"},
			{name: "output", choices: []string{"table", "json"}, defaultValue: config.DefaultOutput},
			{name: "help"},
		}
	case "block_storage", "object_storage":
		return []flagHelpSpec{
			{name: "cloud-type"},
			{name: "storage-type", required: true, choices: []string{"block", "object"}},
			{name: "debug"},
			{name: "lang"},
			{name: "output", choices: []string{"table", "json"}, defaultValue: config.DefaultOutput},
			{name: "help"},
		}
	case "generic", "":
		return []flagHelpSpec{
			{name: "cloud-type"},
			{name: "storage-type", choices: []string{"block", "object"}},
			{name: "file"},
			{name: "body"},
			{name: "preview-request"},
			{name: "debug"},
			{name: "lang"},
			{name: "output", choices: []string{"table", "json"}, defaultValue: config.DefaultOutput},
			{name: "help"},
		}
	default:
		if strings.HasPrefix(profile, "block|") {
			return []flagHelpSpec{
				{name: "cloud-type", required: true},
				{name: "storage-type", required: true, choices: []string{"block", "object"}},
				{name: "cloud-auth-type", choices: []string{"aksk", "password"}},
				{name: "account-name"},
				{name: "file"},
				{name: "set"},
				{name: "set-json"},
				{name: "preview-request"},
				{name: "debug"},
				{name: "lang"},
				{name: "output", choices: []string{"table", "json"}, defaultValue: config.DefaultOutput},
				{name: "help"},
			}
		}
		return []flagHelpSpec{
			{name: "cloud-type", required: true},
			{name: "storage-type", required: true, choices: []string{"block", "object"}},
			{name: "cloud-auth-type", choices: []string{"aksk", "password"}},
			{name: "custom-name"},
			{name: "file"},
			{name: "set"},
			{name: "set-json"},
			{name: "preview-request"},
			{name: "debug"},
			{name: "lang"},
			{name: "output", choices: []string{"table", "json"}, defaultValue: config.DefaultOutput},
			{name: "help"},
		}
	}
}

func omitCloudAccountCreateSelectionFlags(specs []flagHelpSpec) []flagHelpSpec {
	filtered := make([]flagHelpSpec, 0, len(specs))
	for _, spec := range specs {
		if spec.name == "cloud-type" || spec.name == "storage-type" {
			continue
		}
		filtered = append(filtered, spec)
	}
	return filtered
}

func bootConfigFetchProviderFlagSpecs() []flagHelpSpec {
	return []flagHelpSpec{
		{name: "cloud-account-id", required: true},
		{name: "fetch-res"},
		{name: "host-id"},
		{name: "storage-id"},
		{name: "network-addr-for-write-data"},
		{name: "network-addr-for-read-data"},
		{name: "region-id"},
		{name: "zone-id"},
		{name: "cloud-account-username"},
		{name: "cloud-account-use-public"},
		{name: "flavor-id"},
		{name: "boot-loader-flavor-id"},
		{name: "arch"},
		{name: "os-type-id"},
		{name: "os-type"},
		{name: "flavors"},
		{name: "flavor-vcpus"},
		{name: "flavor-ram"},
		{name: "max-nic-num"},
		{name: "system-volume-type-id"},
		{name: "volume-type-id"},
		{name: "default-volume-type-id"},
		{name: "default-pool-id"},
		{name: "dest-boot-mode"},
		{name: "network-id"},
		{name: "debug"},
		{name: "lang"},
		{name: "output", choices: []string{"table", "json"}, defaultValue: config.DefaultOutput},
		{name: "help"},
	}
}

func targetResourceDirectAuthFlagSpecs() []flagHelpSpec {
	return []flagHelpSpec{
		{name: "cloud-auth-type", choices: []string{"aksk", "password"}},
		{name: "fetch-res"},
		{name: "region-id"},
		{name: "zone-id"},
		{name: "boot-mode"},
		{name: "debug"},
		{name: "lang"},
		{name: "output", choices: []string{"table", "json"}, defaultValue: config.DefaultOutput},
		{name: "help"},
	}
}

func targetResourceOpenStackFlagSpecs() []flagHelpSpec {
	return []flagHelpSpec{
		{name: "auth-url", required: true},
		{name: "username", required: true},
		{name: "password", required: true},
		{name: "user-domain-id", required: true},
		{name: "fetch-res"},
		{name: "region-id"},
		{name: "zone-id"},
		{name: "project-id"},
		{name: "project-domain-id"},
		{name: "project-name"},
		{name: "compute-zone-id"},
		{name: "block-store-zone-id"},
		{name: "debug"},
		{name: "lang"},
		{name: "output", choices: []string{"table", "json"}, defaultValue: config.DefaultOutput},
		{name: "help"},
	}
}

func orderedFlagUsages(ctx *context, cmd *cobra.Command, specs []flagHelpSpec) string {
	if len(specs) == 0 {
		return ""
	}
	tmp := pflag.NewFlagSet(cmd.Name(), pflag.ContinueOnError)
	tmp.SortFlags = false
	for _, spec := range specs {
		if hideFlagFromPublicHelp(spec.name) {
			continue
		}
		flag := lookupAnyFlag(cmd, spec.name)
		if flag == nil {
			continue
		}
		copy := *flag
		copy.Usage = decorateFlagUsage(ctx, flag.Usage, spec)
		tmp.AddFlag(&copy)
	}
	return strings.TrimRight(tmp.FlagUsagesWrapped(120), "\n")
}

func lookupAnyFlag(cmd *cobra.Command, name string) *pflag.Flag {
	for _, flags := range []*pflag.FlagSet{
		cmd.NonInheritedFlags(),
		cmd.InheritedFlags(),
		cmd.PersistentFlags(),
		cmd.Flags(),
	} {
		if flags == nil {
			continue
		}
		if flag := flags.Lookup(name); flag != nil {
			return flag
		}
	}
	return nil
}

func decorateFlagUsage(ctx *context, usage string, spec flagHelpSpec) string {
	usage = strings.TrimSpace(usage)
	if spec.required {
		usage += ctx.loc.T("help.note_required")
	}
	if spec.requiredOnInit {
		usage += ctx.loc.T("help.note_init_required")
	}
	if spec.noteKey != "" && !hideHelpNoteInPublicFlags(spec.noteKey) {
		usage += ctx.loc.T(spec.noteKey)
	}
	if len(spec.choices) > 0 && !containsInlineChoicesHint(ctx, usage) {
		usage += ctx.loc.T("help.inline_separator") + ctx.loc.T("help.inline_choices") + " " + strings.Join(spec.choices, " / ")
	}
	if spec.defaultValue != "" && !containsInlineDefaultHint(ctx, usage) {
		usage += ctx.loc.T("help.inline_separator") + ctx.loc.T("help.inline_default") + " " + spec.defaultValue
	}
	return stripHiddenHelpNotes(ctx, usage)
}

func hideFlagFromPublicHelp(name string) bool {
	switch name {
	case "file", "body":
		return true
	default:
		return false
	}
}

func hideHelpNoteInPublicFlags(noteKey string) bool {
	switch noteKey {
	case "help.note_required_unless_file":
		return true
	default:
		return false
	}
}

func stripHiddenHelpNotes(ctx *context, usage string) string {
	for _, noteKey := range []string{"help.note_required_unless_file"} {
		note := ctx.loc.T(noteKey)
		if note == "" {
			continue
		}
		usage = strings.ReplaceAll(usage, note, "")
	}
	return strings.TrimSpace(usage)
}

func containsInlineChoicesHint(ctx *context, usage string) bool {
	if strings.Contains(usage, ctx.loc.T("help.inline_choices")) {
		return true
	}
	if ctx.loc.Lang() == "zh_cn" {
		return strings.Contains(usage, "可选 ")
	}
	return strings.Contains(strings.ToLower(usage), "choices ")
}

func containsInlineDefaultHint(ctx *context, usage string) bool {
	if strings.Contains(usage, ctx.loc.T("help.inline_default")) {
		return true
	}
	if ctx.loc.Lang() == "zh_cn" {
		return strings.Contains(usage, "默认值 ")
	}
	return strings.Contains(strings.ToLower(usage), "default ")
}
