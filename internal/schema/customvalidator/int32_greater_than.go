package customvalidator

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

func Int32GreaterThan(val int32) int32GreaterThan {
	return int32GreaterThan{
		min: val,
	}
}

type int32GreaterThan struct {
	min int32
}

func (v int32GreaterThan) Description(ctx context.Context) string {
	return fmt.Sprintf("Value must be greater than %d", v.min)
}
func (v int32GreaterThan) MarkdownDescription(ctx context.Context) string {
	return fmt.Sprintf("Value must be greater than `%d`", v.min)
}

func (v int32GreaterThan) ValidateInt32(ctx context.Context, req validator.Int32Request, resp *validator.Int32Response) {
	if req.ConfigValue.IsUnknown() || req.ConfigValue.IsNull() {
		return
	}

	if req.ConfigValue.ValueInt32() < v.min {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid value provided",
			fmt.Sprintf("Value must be greater than %d, got: %d.", v.min, req.ConfigValue.ValueInt32()),
		)
		return
	}
}
