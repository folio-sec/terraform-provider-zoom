package customvalidator

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

func Int64GreaterThan(val int64) int64GreaterThan {
	return int64GreaterThan{
		min: val,
	}
}

type int64GreaterThan struct {
	min int64
}

func (v int64GreaterThan) Description(ctx context.Context) string {
	return fmt.Sprintf("Value must be greater than %d", v.min)
}
func (v int64GreaterThan) MarkdownDescription(ctx context.Context) string {
	return fmt.Sprintf("Value must be greater than `%d`", v.min)
}

func (v int64GreaterThan) ValidateInt64(ctx context.Context, req validator.Int64Request, resp *validator.Int64Response) {
	if req.ConfigValue.IsUnknown() || req.ConfigValue.IsNull() {
		return
	}

	if req.ConfigValue.ValueInt64() < v.min {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid value provided",
			fmt.Sprintf("Value must be greater than %d, got: %d.", v.min, req.ConfigValue.ValueInt64()),
		)
		return
	}
}
