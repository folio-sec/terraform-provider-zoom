package util

import (
	"time"

	"github.com/folio-sec/terraform-provider-zoom/generated/api/zoomphone"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func FromOptString(o zoomphone.OptString) types.String {
	v, ok := o.Get()
	if !ok {
		return types.StringNull()
	}
	return types.StringValue(v)
}

func ToOptString(o types.String) zoomphone.OptString {
	if o.IsNull() {
		return zoomphone.OptString{}
	}
	return zoomphone.NewOptString(o.ValueString())
}

func FromOptInt64(o zoomphone.OptInt64) types.Int64 {
	v, ok := o.Get()
	if !ok {
		return types.Int64Null()
	}
	return types.Int64Value(v)
}

func ToOptInt64(o types.Int64) zoomphone.OptInt64 {
	if o.IsNull() {
		return zoomphone.OptInt64{}
	}
	return zoomphone.NewOptInt64(o.ValueInt64())
}

func FromOptDateTime(o zoomphone.OptDateTime) types.String {
	v, ok := o.Get()
	if !ok {
		return types.StringNull()
	}
	str := v.String()
	return types.StringValue(str)
}

func ToOptDateTime(o types.String) zoomphone.OptDateTime {
	if o.IsNull() {
		return zoomphone.OptDateTime{}
	}
	value, _ := time.Parse(o.ValueString(), "2006-01-02 15:04:05.999999999 -0700 MST")
	return zoomphone.NewOptDateTime(value)
}
