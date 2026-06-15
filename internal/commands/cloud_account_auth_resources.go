package commands

import (
	"strings"

	"hyperbdr-client/internal/client"
	"hyperbdr-client/internal/normalize/cloudinfo"
	"hyperbdr-client/internal/output"
)

func writeAuthResourcesResponse(ctx *context, resp client.APIResponse, provider, cloudType, storageType, fetchRes, flavorVCPUs, flavorRAM string) error {
	data := authResourcesResponseData(resp)
	if ctx.cfg.Output == "json" {
		return writeResponse(ctx, resp, "", nil)
	}

	envelope := cloudinfo.BuildAuthResourceEnvelope(provider, cloudType, storageType, data, normalizeAuthRequestedResources(fetchRes), map[string]interface{}{
		"flavor_vcpus": flavorVCPUs,
		"flavor_ram":   flavorRAM,
	})
	if len(envelope.Sections) == 0 {
		return writeValue(ctx, data)
	}

	for i, section := range envelope.Sections {
		if i > 0 {
			if _, err := ctx.out.Write([]byte("\n")); err != nil {
				return err
			}
		}
		if err := writeSectionTitle(ctx, ctx.loc.T(section.TitleKey)); err != nil {
			return err
		}
		if err := writeAuthResourceBody(ctx, envelope, section); err != nil {
			return err
		}
	}
	return nil
}

func authResourcesResponseData(resp client.APIResponse) interface{} {
	if resp.Data != nil {
		return resp.Data
	}
	return resp.Raw
}

func normalizeAuthRequestedResources(fetchRes string) []string {
	if strings.TrimSpace(fetchRes) == "" {
		return nil
	}
	parts := strings.Split(fetchRes, ",")
	resources := make([]string, 0, len(parts))
	seen := map[string]bool{}
	for _, part := range parts {
		resource := normalizeAuthResourceName(part)
		if resource == "" || seen[resource] {
			continue
		}
		seen[resource] = true
		resources = append(resources, resource)
	}
	return resources
}

func normalizeAuthResourceName(name string) string {
	resource := strings.TrimSpace(strings.ToLower(strings.ReplaceAll(name, "-", "_")))
	switch resource {
	case "region":
		return "regions"
	case "flavor":
		return "flavors"
	default:
		return resource
	}
}

func writeAuthResourceBody(ctx *context, envelope cloudinfo.ResourceEnvelope, section cloudinfo.ResourceSection) error {
	renderer := resolveAuthResourceRenderer(envelope.Provider, envelope.CloudType, envelope.StorageType, section)
	if renderer == nil {
		return writeValue(ctx, section.Data)
	}

	rows, err := renderer.Normalize(section)
	if err != nil {
		return err
	}
	return writeAuthTable(ctx, rows, renderer.Columns(rows))
}

func writeAuthTable(ctx *context, rows []map[string]interface{}, cols []output.Column) error {
	return output.Table(ctx.out, ctx.loc, rows, gatewayVisibleColumns(rows, cols))
}
