package gqlgen_opentelemetry

import (
	"context"
	"fmt"

	"github.com/99designs/gqlgen/graphql"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

const (
	extensionName          = "github.com/zhevron/gqlgen-opentelemetry"
	graphqlComponent       = "graphql"
	graphqlFieldName       = "graphql.field.name"
	graphqlFieldPath       = "graphql.field.path"
	graphqlFieldType       = "graphql.field.type"
	graphqlOperationName   = "graphql.operation.name"
	graphqlOperationType   = "graphql.operation.type"
	graphqlVariablesPrefix = "graphql.variables."
	graphqlQuery           = "graphql.query"
)

type Tracer struct {
	IncludeVariables bool
	Tracer           trace.Tracer
}

func (Tracer) ExtensionName() string {
	return extensionName
}

func (t Tracer) Validate(schema graphql.ExecutableSchema) error {
	return nil
}

func (t Tracer) InterceptOperation(ctx context.Context, next graphql.OperationHandler) graphql.ResponseHandler {
	oc := graphql.GetOperationContext(ctx)
	operationName := getOperationName(oc)
	ctx, span := t.tracer(ctx).Start(ctx, operationName, trace.WithAttributes(
		attribute.String("component", graphqlComponent),
		attribute.String(graphqlOperationName, operationName),
		attribute.String(graphqlOperationType, string(oc.Operation.Operation)),
		attribute.String(graphqlQuery, oc.RawQuery),
	))
	if t.IncludeVariables {
		for name, value := range oc.Variables {
			span.SetAttributes(attribute.KeyValue{
				Key:   attribute.Key(graphqlVariablesPrefix + name),
				Value: makeAttributeValue(value),
			})
		}
	}
	handler := next(ctx)
	span.End()
	return handler
}

func (t Tracer) InterceptField(ctx context.Context, next graphql.Resolver) (interface{}, error) {
	fc := graphql.GetFieldContext(ctx)
	if !fc.IsMethod || !fc.IsResolver {
		return next(ctx)
	}
	ctx, span := t.tracer(ctx).Start(ctx, fc.Field.ObjectDefinition.Name+"."+fc.Field.Name, trace.WithAttributes(
		attribute.String(graphqlFieldName, fc.Field.Name),
		attribute.String(graphqlFieldPath, fc.Path().String()),
		attribute.String(graphqlFieldType, fc.Field.ObjectDefinition.Name),
	))
	res, err := next(ctx)
	if errList := graphql.GetFieldErrors(ctx, fc); len(errList) > 0 {
		span.SetStatus(codes.Error, errList.Error())
		for _, err := range errList {
			span.RecordError(err)
		}
	}
	span.End()
	return res, err
}

func (t Tracer) tracer(ctx context.Context) trace.Tracer {
	if t.Tracer != nil {
		return t.Tracer
	}
	if span := trace.SpanFromContext(ctx); span.SpanContext().IsValid() {
		return span.TracerProvider().Tracer(extensionName)
	} else {
		return otel.GetTracerProvider().Tracer(extensionName)
	}
}

func getOperationName(oc *graphql.OperationContext) string {
	if oc.Operation.Name != "" {
		return oc.Operation.Name
	}
	return oc.RawQuery
}

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
		return attribute.StringValue(fmt.Sprintf("%v", v))
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
			arr[i] = fmt.Sprintf("%v", v)
		}
		return attribute.StringSliceValue(arr)
	}
}

var _ interface {
	graphql.HandlerExtension
	graphql.OperationInterceptor
	graphql.FieldInterceptor
} = Tracer{}
