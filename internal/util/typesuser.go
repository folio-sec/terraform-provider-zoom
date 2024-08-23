package util

import (
	"github.com/folio-sec/terraform-provider-zoom/generated/api/zoomuser"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func ToUserOptBool(o types.Bool) zoomuser.OptBool {
	if o.IsNull() || o.IsUnknown() {
		return zoomuser.OptBool{}
	}
	return zoomuser.NewOptBool(o.ValueBool())
}

func ToUserOptString(o types.String) zoomuser.OptString {
	if o.IsNull() || o.IsUnknown() {
		return zoomuser.OptString{}
	}
	return zoomuser.NewOptString(o.ValueString())
}

func ToUserOptInt64(o types.Int64) zoomuser.OptInt64 {
	if o.IsNull() || o.IsUnknown() {
		return zoomuser.OptInt64{}
	}
	return zoomuser.NewOptInt64(o.ValueInt64())
}

func ToUserOptInt(o types.Int32) zoomuser.OptInt {
	if o.IsNull() || o.IsUnknown() {
		return zoomuser.OptInt{}
	}
	return zoomuser.NewOptInt(int(o.ValueInt32()))
}

func ToUserOptDateTime(o timetypes.RFC3339) zoomuser.OptDateTime {
	if o.IsNull() || o.IsUnknown() {
		return zoomuser.OptDateTime{}
	}

	value, _ := o.ValueRFC3339Time()
	return zoomuser.NewOptDateTime(value)
}
