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
		fmt.Fprintf(ctx.out, "\n%s:\n%s\n", ctx.loc.T("help.section_global_flags"), flags)
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
	description := strings.TrimSpace(cmd.Short)
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
		{name: "help"},
	}
}

func flagSpecsForCommand(cmd *cobra.Command) []flagHelpSpec {
	path := cmd.CommandPath()
	switch {
	case strings.HasPrefix(path, "hyperbdrctl target account fetch-block-resources "):
		return []flagHelpSpec{
			{name: "access-key-id", required: true},
			{name: "access-key-secret", required: true},
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
				{name: "cloud-account-username", required: true},
				{name: "cloud-account-password", required: true},
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
			{name: "access-key-id", required: true},
			{name: "access-key-secret", required: true},
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
			{name: "cloud-account-username", required: true},
			{name: "cloud-account-password", required: true},
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
			{name: "cloud-auth-type", required: true, choices: []string{"aksk", "password"}},
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
			{name: "cloud-account-username", required: true},
			{name: "cloud-account-password", required: true},
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
			{name: "cloud-auth-type", required: true, choices: []string{"aksk", "password"}},
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
	case "hyperbdrctl config":
		return []flagHelpSpec{
			{name: "help"},
			{name: "debug"},
			{name: "lang"},
			{name: "output"},
		}
	case "hyperbdrctl source":
		return []flagHelpSpec{
			{name: "debug"},
			{name: "lang"},
			{name: "output", choices: []string{"table", "json"}, defaultValue: config.DefaultOutput},
			{name: "help"},
		}
	case "hyperbdrctl source list":
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
	case "hyperbdrctl source detail":
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
	case "hyperbdrctl source vms":
		return []flagHelpSpec{
			{name: "connection-type", required: true},
			{name: "connection-uuid"},
			{name: "registered"},
			{name: "kw"},
			{name: "page", defaultValue: "1"},
			{name: "page-size", defaultValue: "10"},
			{name: "debug"},
			{name: "lang"},
			{name: "output", choices: []string{"table", "json"}, defaultValue: config.DefaultOutput},
			{name: "help"},
		}
	case "hyperbdrctl source agent-install":
		return []flagHelpSpec{
			{name: "debug"},
			{name: "lang"},
			{name: "output", choices: []string{"table", "json"}, defaultValue: config.DefaultOutput},
			{name: "help"},
		}
	case "hyperbdrctl source agentless-install":
		return []flagHelpSpec{
			{name: "debug"},
			{name: "lang"},
			{name: "output", choices: []string{"table", "json"}, defaultValue: config.DefaultOutput},
			{name: "help"},
		}
	case "hyperbdrctl source sync-nodes":
		return []flagHelpSpec{
			{name: "type", defaultValue: "proxy"},
			{name: "status", defaultValue: "online"},
			{name: "debug"},
			{name: "lang"},
			{name: "output", choices: []string{"table", "json"}, defaultValue: config.DefaultOutput},
			{name: "help"},
		}
	case "hyperbdrctl source create":
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
	case "hyperbdrctl target account list":
		return []flagHelpSpec{
			{name: "page", defaultValue: "1"},
			{name: "page-size", defaultValue: "100"},
			{name: "storage-type"},
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
	case "hyperbdrctl target supports":
		return []flagHelpSpec{
			{name: "debug"},
			{name: "lang"},
			{name: "output", choices: []string{"table", "json"}, defaultValue: config.DefaultOutput},
			{name: "help"},
		}
	case "hyperbdrctl target oss":
		return []flagHelpSpec{
			{name: "debug"},
			{name: "lang"},
			{name: "output", choices: []string{"table", "json"}, defaultValue: config.DefaultOutput},
			{name: "help"},
		}
	case "hyperbdrctl target oss list":
		return []flagHelpSpec{
			{name: "page", defaultValue: "1"},
			{name: "page-size", defaultValue: "100"},
			{name: "type", defaultValue: "objectstorage"},
			{name: "debug"},
			{name: "lang"},
			{name: "output", choices: []string{"table", "json"}, defaultValue: config.DefaultOutput},
			{name: "help"},
		}
	case "hyperbdrctl target oss detail":
		return []flagHelpSpec{
			{name: "id", required: true},
			{name: "debug"},
			{name: "lang"},
			{name: "output", choices: []string{"table", "json"}, defaultValue: config.DefaultOutput},
			{name: "help"},
		}
	case "hyperbdrctl target oss wait":
		return []flagHelpSpec{
			{name: "id", required: true},
			{name: "interval-seconds", defaultValue: "60"},
			{name: "timeout-seconds", defaultValue: "3600"},
			{name: "debug"},
			{name: "lang"},
			{name: "output", choices: []string{"table", "json"}, defaultValue: config.DefaultOutput},
			{name: "help"},
		}
	case "hyperbdrctl target oss buckets":
		return []flagHelpSpec{
			{name: "auth-url", required: true},
			{name: "region-id"},
			{name: "access-key-id", required: true},
			{name: "access-key-secret", required: true},
			{name: "protocol", defaultValue: "s3"},
			{name: "bucket-lookup", defaultValue: "dns"},
			{name: "use-tls", defaultValue: "true"},
			{name: "debug"},
			{name: "lang"},
			{name: "output", choices: []string{"table", "json"}, defaultValue: config.DefaultOutput},
			{name: "help"},
		}
	case "hyperbdrctl target oss create":
		return []flagHelpSpec{
			{name: "file"},
			{name: "display-name", noteKey: "help.note_required_unless_file"},
			{name: "cloud-type", defaultValue: "aliyun"},
			{name: "auth-url", noteKey: "help.note_required_unless_file"},
			{name: "region-id", noteKey: "help.note_required_unless_file"},
			{name: "access-key-id", noteKey: "help.note_required_unless_file"},
			{name: "access-key-secret", noteKey: "help.note_required_unless_file"},
			{name: "protocol", defaultValue: "s3"},
			{name: "bucket-lookup", defaultValue: "dns"},
			{name: "use-tls", defaultValue: "true"},
			{name: "bucket-mode", choices: []string{"existing", "new"}, defaultValue: "existing"},
			{name: "bucket-name", noteKey: "help.note_required_unless_file"},
			{name: "public-endpoint"},
			{name: "internal-endpoint"},
			{name: "cloud-type-select"},
			{name: "app-id"},
			{name: "preview-request"},
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
			{name: "insecure", defaultValue: "false"},
			{name: "debug", defaultValue: "false"},
			{name: "lang", choices: []string{"en", "zh_cn"}, defaultValue: config.DefaultLang},
			{name: "output", choices: []string{"table", "json"}, defaultValue: config.DefaultOutput},
			{name: "help"},
		}
	default:
		return nil
	}
}

func orderedFlagUsages(ctx *context, cmd *cobra.Command, specs []flagHelpSpec) string {
	if len(specs) == 0 {
		return ""
	}
	tmp := pflag.NewFlagSet(cmd.Name(), pflag.ContinueOnError)
	tmp.SortFlags = false
	for _, spec := range specs {
		flag := lookupAnyFlag(cmd, spec.name)
		if flag == nil {
			continue
		}
		copy := *flag
		copy.Usage = decorateFlagUsage(ctx, flag.Usage, spec)
		tmp.AddFlag(&copy)
	}
	return strings.TrimRight(tmp.FlagUsagesWrapped(88), "\n")
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
	if spec.noteKey != "" {
		usage += ctx.loc.T(spec.noteKey)
	}
	if len(spec.choices) > 0 && !containsInlineChoicesHint(ctx, usage) {
		usage += ctx.loc.T("help.inline_separator") + ctx.loc.T("help.inline_choices") + " " + strings.Join(spec.choices, " / ")
	}
	if spec.defaultValue != "" && !containsInlineDefaultHint(ctx, usage) {
		usage += ctx.loc.T("help.inline_separator") + ctx.loc.T("help.inline_default") + " " + spec.defaultValue
	}
	return usage
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
