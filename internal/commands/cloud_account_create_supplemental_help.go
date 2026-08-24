package commands

import "strings"

func appendCloudAccountCreateSupplementalHelp(ctx *context, notes string, profile cloudAccountCreateProfile) string {
	helpContext := dynamicParameterHelpContext{
		Command:      dynamicParameterHelpCloudAccountCreate,
		Provider:     profile.Provider,
		CloudType:    profile.Entry.CloudType,
		Architecture: profile.Entry.Architecture,
		StorageType:  profile.StorageType,
	}
	specs := omitDynamicParameterHelpFlags(
		cloudAccountCreateFlagSpecsForProfile(profile.StorageType+"|"+profile.Provider),
		helpContext,
	)
	notes, dynamicText := prepareDynamicParameterHelp(ctx, notes, helpContext)
	sections := make([]string, 0, 3)
	if cloudAccountCreateHelpHasFlags(specs, "set", "set-json") {
		sections = append(sections, ctx.loc.T("help.cloud_account.create.metadata_overrides"))
	}
	if dynamicText != "" {
		sections = append(sections, dynamicText)
	}
	if cloudAccountCreateHelpHasFlags(specs, "preview-request") {
		sections = append(sections, ctx.loc.T("help.cloud_account.create.preview_request"))
	}
	return appendUsageNoteSections(notes, sections...)
}

func appendUsageNoteSections(notes string, sections ...string) string {
	parts := []string{strings.TrimSpace(notes)}
	for _, section := range sections {
		if section = strings.TrimSpace(section); section != "" {
			parts = append(parts, section)
		}
	}
	return strings.Join(parts, "\n\n")
}

func cloudAccountCreateHelpHasFlags(specs []flagHelpSpec, names ...string) bool {
	available := make(map[string]bool, len(specs))
	for _, spec := range specs {
		available[spec.name] = true
	}
	for _, name := range names {
		if !available[name] {
			return false
		}
	}
	return true
}
