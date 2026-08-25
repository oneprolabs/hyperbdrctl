package commands

import (
	"strings"

	"hyperbdr-client/catalog"

	"github.com/spf13/cobra"
)

const (
	dynamicParameterHelpCommandAnnotation      = "dynamic-parameter-help-command"
	dynamicParameterHelpProviderAnnotation     = "dynamic-parameter-help-provider"
	dynamicParameterHelpCloudTypeAnnotation    = "dynamic-parameter-help-cloud-type"
	dynamicParameterHelpArchitectureAnnotation = "dynamic-parameter-help-architecture"
	dynamicParameterHelpStorageTypeAnnotation  = "dynamic-parameter-help-storage-type"

	dynamicParameterHelpCloudAccountCreate     = "cloud-account-create"
	dynamicParameterHelpCloudSyncGatewayCreate = "cloud-sync-gateway-create"
)

type dynamicParameterHelpContext struct {
	Command      string
	Provider     string
	CloudType    string
	Architecture string
	StorageType  string
}

type dynamicParameterHelpParameter struct {
	Flag           string
	Placeholder    string
	DescriptionKey string
	DefaultValue   string
	Choices        []string
	SourceKey      string
}

type dynamicParameterHelpGroup struct {
	Key                 string
	TitleKey            string
	LegacyUsageNoteKeys []string
	Parameters          []dynamicParameterHelpParameter
}

type dynamicParameterHelpAttachment struct {
	GroupKey         string
	Command          string
	Provider         string
	ExcludedProvider string
	CloudType        string
	Architecture     string
	StorageType      string
}

var dynamicParameterHelpGroups = map[string]dynamicParameterHelpGroup{
	"atomy-v2-cloud-account": {
		Key:                 "atomy-v2-cloud-account",
		TitleKey:            "help.dynamic_parameter.optional",
		LegacyUsageNoteKeys: []string{"help.dynamic_parameter.account_name.legacy"},
		Parameters: []dynamicParameterHelpParameter{
			{Flag: "account-name", Placeholder: "<name>", DescriptionKey: "help.dynamic_parameter.account_name"},
		},
	},
	"huawei-block-cloud-account-auth-project": {
		Key:      "huawei-block-cloud-account-auth-project",
		TitleKey: "help.dynamic_parameter.optional",
		Parameters: []dynamicParameterHelpParameter{
			{Flag: "auth-project-id", Placeholder: "<project_id>", DescriptionKey: "help.dynamic_parameter.auth_project_id"},
		},
	},
	"openstack-block-cloud-account-image-access": {
		Key:      "openstack-block-cloud-account-image-access",
		TitleKey: "help.dynamic_parameter.optional",
		Parameters: []dynamicParameterHelpParameter{
			{Flag: "ssh-port", Placeholder: "<port>", DescriptionKey: "help.target_account_create_block_openstack.flag.ssh-port", DefaultValue: "22"},
			{Flag: "ssh-pass", Placeholder: "<password>", DescriptionKey: "help.target_account_create_block_openstack.flag.ssh-pass"},
			{Flag: "linux-hd-username", Placeholder: "<username>", DescriptionKey: "help.target_account_create_block_openstack.flag.linux-hd-username"},
			{Flag: "linux-hd-password", Placeholder: "<password>", DescriptionKey: "help.target_account_create_block_openstack.flag.linux-hd-password"},
			{Flag: "linux-hd-port", Placeholder: "<port>", DescriptionKey: "help.target_account_create_block_openstack.flag.linux-hd-port", DefaultValue: "10729"},
		},
	},
	"atomy-v2-cloud-sync-gateway": {
		Key:      "atomy-v2-cloud-sync-gateway",
		TitleKey: "help.dynamic_parameter.optional",
		Parameters: []dynamicParameterHelpParameter{
			{
				Flag:           "hg-control-network",
				Placeholder:    "<mode>",
				DescriptionKey: "help.dynamic_parameter.hg_control_network",
				DefaultValue:   "floating_ip_without_proxy",
				Choices: []string{
					"floating_ip_without_proxy",
					"fixed_ip_without_proxy",
					"floating_ip_with_proxy",
					"fixed_ip_with_proxy",
				},
			},
			{
				Flag:           "hg-data-network",
				Placeholder:    "<mode>",
				DescriptionKey: "help.dynamic_parameter.hg_data_network",
				DefaultValue:   "floating_ip_without_proxy",
				Choices: []string{
					"floating_ip_without_proxy",
					"fixed_ip_without_proxy",
					"floating_ip_with_proxy",
					"fixed_ip_with_proxy",
				},
			},
			{Flag: "control-nat-ip", Placeholder: "<ip>", DescriptionKey: "help.dynamic_parameter.control_nat_ip"},
			{Flag: "data-nat-ip", Placeholder: "<ip>", DescriptionKey: "help.dynamic_parameter.data_nat_ip"},
			{Flag: "boot-loader-flavor-id", Placeholder: "<flavor_id>", DescriptionKey: "help.dynamic_parameter.boot_loader_flavor_id"},
			{
				Flag:           "boot-loader-image-id",
				Placeholder:    "<image_id>",
				DescriptionKey: "help.dynamic_parameter.boot_loader_image_id",
				SourceKey:      "help.dynamic_parameter.boot_loader_image_id.source",
			},
		},
	},
	"aliyun-cloud-sync-gateway-advanced": {
		Key:      "aliyun-cloud-sync-gateway-advanced",
		TitleKey: "help.dynamic_parameter.optional",
		Parameters: []dynamicParameterHelpParameter{
			{
				Flag:           "hd-control-network",
				Placeholder:    "<mode>",
				DescriptionKey: "help.dynamic_parameter.hd_control_network",
				DefaultValue:   "floating_ip_with_hg_proxy",
				Choices: []string{
					"floating_ip_without_proxy",
					"fixed_ip_without_proxy",
					"floating_ip_with_hg_proxy",
					"fixed_ip_with_hg_proxy",
				},
			},
			{
				Flag:           "fixed-ip",
				Placeholder:    "<ip>",
				DescriptionKey: "help.dynamic_parameter.fixed_ip",
				SourceKey:      "help.dynamic_parameter.fixed_ip.source",
			},
			{
				Flag:           "system-disk-size",
				Placeholder:    "<size_gib>",
				DescriptionKey: "help.dynamic_parameter.system_disk_size",
				DefaultValue:   "40",
			},
			{
				Flag:           "bandwidth-size",
				Placeholder:    "<size_mbps>",
				DescriptionKey: "help.dynamic_parameter.bandwidth_size",
				SourceKey:      "help.dynamic_parameter.bandwidth_size.source",
			},
		},
	},
	"huawei-cloud-sync-gateway-advanced": {
		Key:      "huawei-cloud-sync-gateway-advanced",
		TitleKey: "help.dynamic_parameter.optional",
		Parameters: []dynamicParameterHelpParameter{
			{
				Flag:           "hd-control-network",
				Placeholder:    "<mode>",
				DescriptionKey: "help.dynamic_parameter.hd_control_network",
				DefaultValue:   "floating_ip_with_hg_proxy",
				Choices: []string{
					"floating_ip_without_proxy",
					"fixed_ip_without_proxy",
					"floating_ip_with_hg_proxy",
					"fixed_ip_with_hg_proxy",
				},
			},
			{
				Flag:           "fixed-ip",
				Placeholder:    "<ip>",
				DescriptionKey: "help.dynamic_parameter.fixed_ip",
				SourceKey:      "help.dynamic_parameter.fixed_ip.source",
			},
			{
				Flag:           "system-disk-size",
				Placeholder:    "<size_gib>",
				DescriptionKey: "help.dynamic_parameter.system_disk_size",
			},
			{
				Flag:           "bandwidth-size",
				Placeholder:    "<size_mbps>",
				DescriptionKey: "help.dynamic_parameter.bandwidth_size",
			},
		},
	},
	"openstack-cloud-sync-gateway-advanced": {
		Key:      "openstack-cloud-sync-gateway-advanced",
		TitleKey: "help.dynamic_parameter.optional",
		Parameters: []dynamicParameterHelpParameter{
			{
				Flag:           "hg-control-network",
				Placeholder:    "<mode>",
				DescriptionKey: "help.dynamic_parameter.hg_control_network",
				DefaultValue:   "floating_ip_without_proxy",
				Choices: []string{
					"floating_ip_without_proxy",
					"fixed_ip_without_proxy",
					"floating_ip_with_proxy",
					"fixed_ip_with_proxy",
				},
			},
			{
				Flag:           "hg-data-network",
				Placeholder:    "<mode>",
				DescriptionKey: "help.dynamic_parameter.hg_data_network",
				DefaultValue:   "floating_ip_without_proxy",
				Choices: []string{
					"floating_ip_without_proxy",
					"fixed_ip_without_proxy",
					"floating_ip_with_proxy",
					"fixed_ip_with_proxy",
				},
			},
			{Flag: "control-nat-ip", Placeholder: "<ip>", DescriptionKey: "help.dynamic_parameter.control_nat_ip"},
			{Flag: "data-nat-ip", Placeholder: "<ip>", DescriptionKey: "help.dynamic_parameter.data_nat_ip"},
			{
				Flag:           "boot-loader-flavor-id",
				Placeholder:    "<flavor_id>",
				DescriptionKey: "help.dynamic_parameter.boot_loader_flavor_id",
				SourceKey:      "help.dynamic_parameter.openstack.boot_loader_flavor_id.source",
			},
			{
				Flag:           "boot-loader-image-id",
				Placeholder:    "<image_id>",
				DescriptionKey: "help.dynamic_parameter.boot_loader_image_id",
				SourceKey:      "help.dynamic_parameter.openstack.boot_loader_image_id.source",
			},
			{
				Flag:           "fixed-ip",
				Placeholder:    "<ip>",
				DescriptionKey: "help.dynamic_parameter.fixed_ip",
				SourceKey:      "help.dynamic_parameter.fixed_ip.source",
			},
			{
				Flag:           "system-disk-size",
				Placeholder:    "<size_gib>",
				DescriptionKey: "help.dynamic_parameter.system_disk_size",
				DefaultValue:   "50",
			},
			{
				Flag:           "block-store-zone-id",
				Placeholder:    "<zone_id>",
				DescriptionKey: "flag.block-store-zone-id",
				SourceKey:      "help.dynamic_parameter.openstack.block_store_zone_id.source",
			},
			{
				Flag:           "project-domain-id",
				Placeholder:    "<domain_id>",
				DescriptionKey: "flag.project-domain-id",
				SourceKey:      "help.dynamic_parameter.openstack.project_domain_id.source",
			},
			{
				Flag:           "boot-types-id",
				Placeholder:    "<boot_type>",
				DescriptionKey: "flag.boot-types-id",
				DefaultValue:   "boot_from_volume",
				Choices: []string{
					"boot_from_volume",
					"boot_from_image",
				},
			},
		},
	},
}

var dynamicParameterHelpAttachments = []dynamicParameterHelpAttachment{
	{
		GroupKey:     "atomy-v2-cloud-account",
		Command:      dynamicParameterHelpCloudAccountCreate,
		Architecture: catalog.AtomyV2,
	},
	{
		GroupKey:    "huawei-block-cloud-account-auth-project",
		Command:     dynamicParameterHelpCloudAccountCreate,
		Provider:    "huawei",
		StorageType: "block",
	},
	{
		GroupKey:    "openstack-block-cloud-account-image-access",
		Command:     dynamicParameterHelpCloudAccountCreate,
		Provider:    "openstack",
		StorageType: "block",
	},
	{
		GroupKey:         "atomy-v2-cloud-sync-gateway",
		Command:          dynamicParameterHelpCloudSyncGatewayCreate,
		ExcludedProvider: "huawei",
		Architecture:     catalog.AtomyV2,
	},
	{
		GroupKey:    "aliyun-cloud-sync-gateway-advanced",
		Command:     dynamicParameterHelpCloudSyncGatewayCreate,
		Provider:    "aliyun",
		StorageType: "block",
	},
	{
		GroupKey:    "huawei-cloud-sync-gateway-advanced",
		Command:     dynamicParameterHelpCloudSyncGatewayCreate,
		Provider:    "huawei",
		StorageType: "block",
	},
	{
		GroupKey:    "openstack-cloud-sync-gateway-advanced",
		Command:     dynamicParameterHelpCloudSyncGatewayCreate,
		Provider:    "openstack",
		StorageType: "block",
	},
}

func setDynamicParameterHelpContext(cmd *cobra.Command, context dynamicParameterHelpContext) {
	addAnnotationValue(cmd, dynamicParameterHelpCommandAnnotation, context.Command)
	addAnnotationValue(cmd, dynamicParameterHelpProviderAnnotation, context.Provider)
	addAnnotationValue(cmd, dynamicParameterHelpCloudTypeAnnotation, context.CloudType)
	addAnnotationValue(cmd, dynamicParameterHelpArchitectureAnnotation, context.Architecture)
	addAnnotationValue(cmd, dynamicParameterHelpStorageTypeAnnotation, context.StorageType)
}

func dynamicParameterHelpContextFromCommand(cmd *cobra.Command) dynamicParameterHelpContext {
	return dynamicParameterHelpContext{
		Command:      cmd.Annotations[dynamicParameterHelpCommandAnnotation],
		Provider:     cmd.Annotations[dynamicParameterHelpProviderAnnotation],
		CloudType:    cmd.Annotations[dynamicParameterHelpCloudTypeAnnotation],
		Architecture: cmd.Annotations[dynamicParameterHelpArchitectureAnnotation],
		StorageType:  cmd.Annotations[dynamicParameterHelpStorageTypeAnnotation],
	}
}

func appendDynamicParameterHelp(ctx *context, notes string, helpContext dynamicParameterHelpContext) string {
	notes, dynamicText := prepareDynamicParameterHelp(ctx, notes, helpContext)
	return insertDynamicParameterHelpBeforePreview(notes, dynamicText)
}

// prepareDynamicParameterHelp removes legacy copies of matching dynamic help and
// renders the matching groups. Callers that own a structured usage-note layout
// can append the returned text at an explicit position.
func prepareDynamicParameterHelp(ctx *context, notes string, helpContext dynamicParameterHelpContext) (string, string) {
	groups := matchingDynamicParameterHelpGroups(helpContext)
	if len(groups) == 0 {
		return notes, ""
	}

	for _, group := range groups {
		for _, key := range group.LegacyUsageNoteKeys {
			legacy := ctx.loc.T(key)
			notes = strings.ReplaceAll(notes, "\n\n"+legacy, "")
			notes = strings.ReplaceAll(notes, legacy, "")
		}
	}
	return notes, renderDynamicParameterHelp(ctx, groups)
}

func renderDynamicParameterHelp(ctx *context, groups []dynamicParameterHelpGroup) string {
	var text strings.Builder
	lastTitleKey := ""
	for _, group := range groups {
		if group.TitleKey != lastTitleKey {
			if text.Len() > 0 {
				text.WriteString("\n\n")
			}
			text.WriteString(ctx.loc.T(group.TitleKey))
			lastTitleKey = group.TitleKey
		}
		for index, parameter := range group.Parameters {
			if index > 0 {
				text.WriteString("\n")
			}
			appendDynamicParameterHelpParameter(&text, ctx, parameter)
		}
	}
	return text.String()
}

func insertDynamicParameterHelpBeforePreview(notes, dynamicText string) string {
	notes = strings.TrimSpace(notes)
	if dynamicText == "" {
		return notes
	}
	previewFlagIndex := strings.LastIndex(notes, "\n  --preview-request")
	if previewFlagIndex < 0 {
		return notes + "\n\n" + dynamicText
	}
	sectionIndex := strings.LastIndex(notes[:previewFlagIndex], "\n\n")
	if sectionIndex < 0 {
		return notes + "\n\n" + dynamicText
	}
	return strings.TrimRight(notes[:sectionIndex], "\n") + "\n\n" + dynamicText + notes[sectionIndex:]
}

func appendDynamicParameterHelpParameter(text *strings.Builder, ctx *context, parameter dynamicParameterHelpParameter) {
	text.WriteString("\n  --")
	text.WriteString(parameter.Flag)
	if parameter.Placeholder != "" {
		text.WriteString(" ")
		text.WriteString(parameter.Placeholder)
	}
	if parameter.DescriptionKey != "" {
		text.WriteString("\n    ")
		text.WriteString(ctx.loc.T(parameter.DescriptionKey))
	}
	if parameter.DefaultValue != "" {
		text.WriteString("\n    ")
		text.WriteString(ctx.loc.T("help.dynamic_parameter.default"))
		text.WriteString(parameter.DefaultValue)
	}
	if len(parameter.Choices) > 0 {
		text.WriteString("\n    ")
		text.WriteString(ctx.loc.T("help.dynamic_parameter.choices"))
		for _, choice := range parameter.Choices {
			text.WriteString("\n      ")
			text.WriteString(choice)
		}
	}
	if parameter.SourceKey != "" {
		appendDynamicParameterHelpIndentedText(text, ctx.loc.T(parameter.SourceKey), "    ")
	}
}

func appendDynamicParameterHelpIndentedText(text *strings.Builder, value, indentation string) {
	for _, line := range strings.Split(strings.TrimSpace(value), "\n") {
		text.WriteString("\n")
		text.WriteString(indentation)
		text.WriteString(line)
	}
}

func omitDynamicParameterHelpFlags(specs []flagHelpSpec, helpContext dynamicParameterHelpContext) []flagHelpSpec {
	groups := matchingDynamicParameterHelpGroups(helpContext)
	if len(groups) == 0 {
		return specs
	}
	omitted := map[string]bool{}
	for _, group := range groups {
		for _, parameter := range group.Parameters {
			omitted[parameter.Flag] = true
		}
	}
	filtered := make([]flagHelpSpec, 0, len(specs))
	for _, spec := range specs {
		if !omitted[spec.name] {
			filtered = append(filtered, spec)
		}
	}
	return filtered
}

func matchingDynamicParameterHelpGroups(helpContext dynamicParameterHelpContext) []dynamicParameterHelpGroup {
	groups := make([]dynamicParameterHelpGroup, 0)
	seen := map[string]bool{}
	for _, attachment := range dynamicParameterHelpAttachments {
		if !dynamicParameterHelpAttachmentMatches(attachment, helpContext) || seen[attachment.GroupKey] {
			continue
		}
		group, ok := dynamicParameterHelpGroups[attachment.GroupKey]
		if !ok {
			continue
		}
		seen[attachment.GroupKey] = true
		groups = append(groups, group)
	}
	return groups
}

func dynamicParameterHelpAttachmentMatches(attachment dynamicParameterHelpAttachment, helpContext dynamicParameterHelpContext) bool {
	excludedProvider := strings.TrimSpace(attachment.ExcludedProvider)
	return (excludedProvider == "" || !strings.EqualFold(excludedProvider, strings.TrimSpace(helpContext.Provider))) &&
		dynamicParameterHelpFieldMatches(attachment.Command, helpContext.Command) &&
		dynamicParameterHelpFieldMatches(attachment.Provider, helpContext.Provider) &&
		dynamicParameterHelpFieldMatches(attachment.CloudType, helpContext.CloudType) &&
		dynamicParameterHelpFieldMatches(attachment.Architecture, helpContext.Architecture) &&
		dynamicParameterHelpFieldMatches(attachment.StorageType, helpContext.StorageType)
}

func dynamicParameterHelpFieldMatches(expected, actual string) bool {
	return expected == "" || strings.EqualFold(strings.TrimSpace(expected), strings.TrimSpace(actual))
}
