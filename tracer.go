package gqlgen_opentelemetry

import (
	"context"

	"github.com/99designs/gqlgen/graphql"
	"github.com/vektah/gqlparser/v2/ast"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	semconv "go.opentelemetry.io/otel/semconv/v1.20.0"
	"go.opentelemetry.io/otel/trace"
)

const (
	extensionName          = "github.com/zhevron/gqlgen-opentelemetry"
	graphqlFieldAlias      = "graphql.field.alias"
	graphqlFieldName       = "graphql.field.name"
	graphqlFieldPath       = "graphql.field.path"
	graphqlFieldType       = "graphql.field.type"
	graphqlVariablesPrefix = "graphql.variables."
)

type Tracer struct {
	IncludeFieldSpans bool
	IncludeVariables  bool
	Tracer            trace.Tracer
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
	operationType := getOperationTypeAttribute(oc)
	spanName := makeSpanName(operationName, operationType.Value.AsString())
	ctx, span := t.getTracer(ctx).Start(ctx, spanName, trace.WithAttributes(
		semconv.GraphqlOperationName(operationName),
		operationType,
		semconv.GraphqlDocument(oc.RawQuery),
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
	if !t.IncludeFieldSpans || !fc.IsMethod || !fc.IsResolver {
		return next(ctx)
	}
	spanName := fc.Field.ObjectDefinition.Name + "." + fc.Field.Name
	attributes := []attribute.KeyValue{
		attribute.String(graphqlFieldName, fc.Field.Name),
		attribute.String(graphqlFieldPath, fc.Path().String()),
		attribute.String(graphqlFieldType, fc.Field.ObjectDefinition.Name),
	}
	if fc.Field.Alias != fc.Field.Name {
		attributes = append(attributes, attribute.String(graphqlFieldAlias, fc.Field.Alias))
	}
	ctx, span := t.getTracer(ctx).Start(ctx, spanName, trace.WithAttributes(attributes...))
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

func (t Tracer) getTracer(ctx context.Context) trace.Tracer {
	if t.Tracer != nil {
		return t.Tracer
	}
	if span := trace.SpanFromContext(ctx); span.SpanContext().IsValid() {
		return span.TracerProvider().Tracer(extensionName)
	} else {
		return otel.GetTracerProvider().Tracer(extensionName)
	}
}

func makeSpanName(operationName, operationType string) string {
	spanName := operationType
	if spanName == "" {
		return "GraphQL Operation"
	}
	if operationName != "" {
		spanName += " " + operationName
	}
	return spanName
}

func getOperationName(oc *graphql.OperationContext) string {
	if oc.Operation.Name != "" {
		return oc.Operation.Name
	}
	return oc.RawQuery
}

func getOperationTypeAttribute(oc *graphql.OperationContext) attribute.KeyValue {
	switch oc.Operation.Operation {
	case ast.Mutation:
		return semconv.GraphqlOperationTypeMutation
	case ast.Subscription:
		return semconv.GraphqlOperationTypeSubscription
	case ast.Query:
		return semconv.GraphqlOperationTypeQuery
	default:
		return semconv.GraphqlOperationTypeQuery
	}
}

var _ interface {
	graphql.HandlerExtension
	graphql.OperationInterceptor
	graphql.FieldInterceptor
} = Tracer{}
