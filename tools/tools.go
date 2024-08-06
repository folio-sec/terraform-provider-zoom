//go:build tools

package tools

import (
	// Ensure documentation generator is not removed from go.mod.
	_ "github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs"
	_ "github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen"
)
