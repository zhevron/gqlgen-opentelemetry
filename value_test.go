package gqlgen_opentelemetry

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/attribute"
)

type valueTestStruct struct {
	Value string
}

type valueTestTable []struct {
	input        interface{}
	expectedType attribute.Type
}

func TestMakeAttributeValue(t *testing.T) {
	values := valueTestTable{
		{true, attribute.BOOL},
		{float32(1.5), attribute.FLOAT64},
		{float64(1.5), attribute.FLOAT64},
		{int(1), attribute.INT64},
		{int32(1), attribute.INT64},
		{int64(1), attribute.INT64},
		{"test", attribute.STRING},
		{valueTestStruct{Value: "test"}, attribute.STRING},
	}
	for _, v := range values {
		assert.Equal(t, v.expectedType, makeAttributeValue(v.input).Type())
	}
}

func TestMakeAttributeSliceValue(t *testing.T) {
	values := valueTestTable{
		{true, attribute.BOOLSLICE},
		{float32(1.5), attribute.FLOAT64SLICE},
		{float64(1.5), attribute.FLOAT64SLICE},
		{int(1), attribute.INT64SLICE},
		{int32(1), attribute.INT64SLICE},
		{int64(1), attribute.INT64SLICE},
		{"test", attribute.STRINGSLICE},
		{valueTestStruct{Value: "test"}, attribute.STRINGSLICE},
	}
	for _, v := range values {
		a := []interface{}{v.input}
		assert.Equal(t, v.expectedType, makeAttributeSliceValue(a).Type())
	}
}
