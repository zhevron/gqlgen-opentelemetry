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
	extensionVersion       = "1.0.4"
	graphqlFieldAlias      = "graphql.field.alias"
	graphqlFieldName       = "graphql.field.name"
	graphqlFieldPath       = "graphql.field.path"
	graphqlFieldType       = "graphql.field.type"
	graphqlVariablesPrefix = "graphql.variables."
)

var baseAttributes = []attribute.KeyValue{
	semconv.OTelLibraryName(extensionName),
	semconv.OTelLibraryVersion(extensionVersion),
}

type Tracer struct {
	IncludeFieldSpans bool
	IncludeVariables  bool
	TracerProvider    trace.TracerProvider
}

func (Tracer) ExtensionName() string {
	return extensionName
}

func (t Tracer) Validate(schema graphql.ExecutableSchema) error {
	return nil
}

func (t Tracer) InterceptResponse(ctx context.Context, next graphql.ResponseHandler) *graphql.Response {
	if !graphql.HasOperationContext(ctx) {
		return next(ctx)
	}
	oc := graphql.GetOperationContext(ctx)
	operationType := getOperationTypeAttribute(oc)
	attributes := append(baseAttributes,
		semconv.GraphqlOperationName(oc.Operation.Name),
		operationType,
		semconv.GraphqlDocument(oc.RawQuery),
	)
	spanName := makeSpanName(oc.Operation.Name, operationType.Value.AsString())
	ctx, span := t.getTracer(ctx).Start(ctx, spanName, trace.WithSpanKind(trace.SpanKindServer), trace.WithAttributes(attributes...))
	defer span.End()
	span.SetAttributes(baseAttributes...)
	if t.IncludeVariables {
		for name, value := range oc.Variables {
			span.SetAttributes(attribute.KeyValue{
				Key:   attribute.Key(graphqlVariablesPrefix + name),
				Value: makeAttributeValue(value),
			})
		}
	}
	res := next(ctx)
	if res != nil && len(res.Errors) > 0 {
		span.SetStatus(codes.Error, res.Errors.Error())
		for _, err := range res.Errors {
			span.RecordError(err)
		}
	}
	return res
}

func (t Tracer) InterceptField(ctx context.Context, next graphql.Resolver) (interface{}, error) {
	fc := graphql.GetFieldContext(ctx)
	if !t.IncludeFieldSpans || !fc.IsMethod || !fc.IsResolver {
		return next(ctx)
	}
	spanName := fc.Field.ObjectDefinition.Name + "." + fc.Field.Name
	attributes := append(baseAttributes,
		attribute.String(graphqlFieldName, fc.Field.Name),
		attribute.String(graphqlFieldPath, fc.Path().String()),
		attribute.String(graphqlFieldType, fc.Field.ObjectDefinition.Name),
	)
	if fc.Field.Alias != fc.Field.Name {
		attributes = append(attributes, attribute.String(graphqlFieldAlias, fc.Field.Alias))
	}
	ctx, span := t.getTracer(ctx).Start(ctx, spanName, trace.WithSpanKind(trace.SpanKindServer), trace.WithAttributes(attributes...))
	defer span.End()
	res, err := next(ctx)
	if errList := graphql.GetFieldErrors(ctx, fc); len(errList) > 0 {
		span.SetStatus(codes.Error, errList.Error())
		for _, err := range errList {
			span.RecordError(err)
		}
	}
	return res, err
}

func (t Tracer) getTracer(ctx context.Context) trace.Tracer {
	if span := trace.SpanFromContext(ctx); span.SpanContext().IsValid() {
		return span.TracerProvider().Tracer(extensionName, trace.WithInstrumentationVersion(extensionVersion))
	} else {
		tp := t.TracerProvider
		if tp == nil {
			tp = otel.GetTracerProvider()
		}
		return tp.Tracer(extensionName, trace.WithInstrumentationVersion(extensionVersion))
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
	graphql.ResponseInterceptor
	graphql.FieldInterceptor
} = Tracer{}
