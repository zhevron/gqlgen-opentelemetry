package gqlgen_opentelemetry

import (
	"fmt"

	"go.opentelemetry.io/otel/attribute"
)

func makeAttributeValue(value interface{}) attribute.Value {
	switch v := value.(type) {
	case bool:
		return attribute.BoolValue(v)
	case float32:
		return attribute.Float64Value(float64(v))
	case float64:
		return attribute.Float64Value(v)
	case int:
		return attribute.Int64Value(int64(v))
	case int32:
		return attribute.Int64Value(int64(v))
	case int64:
		return attribute.Int64Value(v)
	case string:
		return attribute.StringValue(v)
	case []interface{}:
		return makeAttributeSliceValue(v)
	default:
		return attribute.StringValue(fmt.Sprintf("%+v", v))
	}
}

func makeAttributeSliceValue(value []interface{}) attribute.Value {
	if len(value) == 0 {
		return attribute.StringSliceValue([]string{})
	}
	switch value[0].(type) {
	case bool:
		arr := make([]bool, len(value))
		for i, v := range value {
			arr[i] = v.(bool)
		}
		return attribute.BoolSliceValue(arr)
	case float32:
		arr := make([]float64, len(value))
		for i, v := range value {
			arr[i] = float64(v.(float32))
		}
		return attribute.Float64SliceValue(arr)
	case float64:
		arr := make([]float64, len(value))
		for i, v := range value {
			arr[i] = v.(float64)
		}
		return attribute.Float64SliceValue(arr)
	case int:
		arr := make([]int64, len(value))
		for i, v := range value {
			arr[i] = int64(v.(int))
		}
		return attribute.Int64SliceValue(arr)
	case int32:
		arr := make([]int64, len(value))
		for i, v := range value {
			arr[i] = int64(v.(int32))
		}
		return attribute.Int64SliceValue(arr)
	case int64:
		arr := make([]int64, len(value))
		for i, v := range value {
			arr[i] = v.(int64)
		}
		return attribute.Int64SliceValue(arr)
	case string:
		arr := make([]string, len(value))
		for i, v := range value {
			arr[i] = v.(string)
		}
		return attribute.StringSliceValue(arr)
	default:
		arr := make([]string, len(value))
		for i, v := range value {
			arr[i] = fmt.Sprintf("%+v", v)
		}
		return attribute.StringSliceValue(arr)
	}
}
