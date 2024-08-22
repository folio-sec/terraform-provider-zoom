package util

import (
	"time"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

type OptValue[A any] interface {
	Get() (v A, ok bool)
}

func FromOptBool(o OptValue[bool]) types.Bool {
	v, ok := o.Get()
	if !ok {
		return types.BoolNull()
	}
	return types.BoolValue(v)
}

func FromOptString(o OptValue[string]) types.String {
	v, ok := o.Get()
	if !ok {
		return types.StringNull()
	}
	return types.StringValue(v)
}

func FromOptInt64(o OptValue[int64]) types.Int64 {
	v, ok := o.Get()
	if !ok {
		return types.Int64Null()
	}
	return types.Int64Value(v)
}

func FromOptInt(o OptValue[int]) types.Int32 {
	v, ok := o.Get()
	if !ok {
		return types.Int32Null()
	}
	return types.Int32Value(int32(v))
}

func FromOptDateTime(o OptValue[time.Time]) types.String {
	v, ok := o.Get()
	if !ok {
		return types.StringNull()
	}
	str := v.String()
	return types.StringValue(str)
}
