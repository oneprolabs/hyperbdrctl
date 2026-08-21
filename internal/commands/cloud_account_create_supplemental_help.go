package commands

import "strings"

func appendCloudAccountCreateSupplementalHelp(ctx *context, notes string, profile cloudAccountCreateProfile) string {
	notes = removeCloudAccountCreateLegacySetHelp(notes)
	notes = removeCloudAccountCreateLegacyPreviewHelp(notes)

	specs := omitDynamicParameterHelpFlags(
		cloudAccountCreateFlagSpecsForProfile(profile.StorageType+"|"+profile.Provider),
		dynamicParameterHelpContext{
			Command:      dynamicParameterHelpCloudAccountCreate,
			Provider:     profile.Provider,
			CloudType:    profile.Entry.CloudType,
			Architecture: profile.Entry.Architecture,
			StorageType:  profile.StorageType,
		},
	)
	var supplemental []string
	if cloudAccountCreateHelpHasFlags(specs, "set", "set-json") {
		supplemental = append(supplemental, ctx.loc.T("help.cloud_account.create.metadata_overrides"))
	}
	if cloudAccountCreateHelpHasFlags(specs, "preview-request") {
		supplemental = append(supplemental, ctx.loc.T("help.cloud_account.create.preview_request"))
	}
	notes = insertCloudAccountCreateSupplementalHelp(notes, strings.Join(supplemental, "\n\n"))

	return appendDynamicParameterHelp(ctx, notes, dynamicParameterHelpContext{
		Command:      dynamicParameterHelpCloudAccountCreate,
		Provider:     profile.Provider,
		CloudType:    profile.Entry.CloudType,
		Architecture: profile.Entry.Architecture,
		StorageType:  profile.StorageType,
	})
}

func insertCloudAccountCreateSupplementalHelp(notes, supplemental string) string {
	supplemental = strings.TrimSpace(supplemental)
	if supplemental == "" {
		return notes
	}
	for _, marker := range []string{
		"\n\nAfter creation succeeds",
		"\n\nAfter a successful create",
		"\n\nAfter create returns",
		"\n\n创建成功后",
	} {
		if index := strings.Index(notes, marker); index >= 0 {
			return notes[:index] + "\n\n" + supplemental + notes[index:]
		}
	}
	return strings.TrimSpace(notes) + "\n\n" + supplemental
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

func removeCloudAccountCreateLegacySetHelp(notes string) string {
	flagIndex := strings.Index(notes, "\n  --set key=value")
	if flagIndex < 0 {
		return notes
	}
	start := strings.LastIndex(notes[:flagIndex], "\n\n")
	if start < 0 {
		return notes
	}
	firstSectionEnd := strings.Index(notes[flagIndex:], "\n\n")
	if firstSectionEnd < 0 {
		return notes
	}
	firstSectionEnd += flagIndex
	secondSectionEnd := strings.Index(notes[firstSectionEnd+2:], "\n\n")
	if secondSectionEnd < 0 {
		return notes[:start]
	}
	secondSectionEnd += firstSectionEnd + 2
	return notes[:start] + notes[secondSectionEnd:]
}

func removeCloudAccountCreateLegacyPreviewHelp(notes string) string {
	flagIndex := strings.LastIndex(notes, "\n  --preview-request")
	if flagIndex < 0 {
		return notes
	}
	start := strings.LastIndex(notes[:flagIndex], "\n\n")
	if start < 0 {
		return notes
	}
	end := strings.Index(notes[flagIndex:], "\n\n")
	if end < 0 {
		return notes[:start]
	}
	return notes[:start] + notes[flagIndex+end:]
}
