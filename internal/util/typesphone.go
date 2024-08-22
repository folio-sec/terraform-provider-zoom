package util

import (
	"time"

	"github.com/folio-sec/terraform-provider-zoom/generated/api/zoomphone"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func ToPhoneOptBool(o types.Bool) zoomphone.OptBool {
	if o.IsNull() || o.IsUnknown() {
		return zoomphone.OptBool{}
	}
	return zoomphone.NewOptBool(o.ValueBool())
}

func ToPhoneOptString(o types.String) zoomphone.OptString {
	if o.IsNull() || o.IsUnknown() {
		return zoomphone.OptString{}
	}
	return zoomphone.NewOptString(o.ValueString())
}

func ToPhoneOptInt64(o types.Int64) zoomphone.OptInt64 {
	if o.IsNull() || o.IsUnknown() {
		return zoomphone.OptInt64{}
	}
	return zoomphone.NewOptInt64(o.ValueInt64())
}

func ToPhoneOptInt(o types.Int32) zoomphone.OptInt {
	if o.IsNull() || o.IsUnknown() {
		return zoomphone.OptInt{}
	}
	return zoomphone.NewOptInt(int(o.ValueInt32()))
}

func ToPhoneOptDateTime(o types.String) zoomphone.OptDateTime {
	if o.IsNull() || o.IsUnknown() {
		return zoomphone.OptDateTime{}
	}
	value, _ := time.Parse(o.ValueString(), "2006-01-02 15:04:05.999999999 -0700 MST")
	return zoomphone.NewOptDateTime(value)
}
