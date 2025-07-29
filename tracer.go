package gqlgen_opentelemetry

import (
	"context"

	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/vektah/gqlparser/v2/ast"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	semconv "go.opentelemetry.io/otel/semconv/v1.34.0"
	"go.opentelemetry.io/otel/trace"
)

const (
	extensionName          = "github.com/zhevron/gqlgen-opentelemetry"
	extensionVersion       = "1.0.4"
	graphqlComplexity      = attribute.Key("graphql.operation.complexity")
	graphqlFieldAlias      = attribute.Key("graphql.field.alias")
	graphqlFieldName       = attribute.Key("graphql.field.name")
	graphqlFieldPath       = attribute.Key("graphql.field.path")
	graphqlFieldType       = attribute.Key("graphql.field.type")
	graphqlVariablesPrefix = "graphql.variables."
)

var baseAttributes = []attribute.KeyValue{
	semconv.OTelScopeName(extensionName),
	semconv.OTelScopeVersion(extensionVersion),
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
	operationName := oc.OperationName
	if oc.Operation != nil && oc.Operation.Name != "" {
		operationName = oc.Operation.Name
	}
	operationType := getOperationTypeAttribute(oc)
	spanName := makeSpanName(operationName, operationType.Value.AsString())
	ctx, span := t.getTracer(ctx).Start(ctx, spanName, trace.WithSpanKind(trace.SpanKindServer), trace.WithAttributes(baseAttributes...))
	defer span.End()
	span.SetAttributes(
		operationType,
		semconv.GraphQLDocument(oc.RawQuery),
	)
	if operationName != "" {
		span.SetAttributes(semconv.GraphQLOperationName(operationName))
	}
	if stats := extension.GetComplexityStats(ctx); stats != nil {
		span.SetAttributes(graphqlComplexity.Int(stats.Complexity))
	}
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
	ctx, span := t.getTracer(ctx).Start(ctx, spanName, trace.WithSpanKind(trace.SpanKindServer), trace.WithAttributes(baseAttributes...))
	defer span.End()
	span.SetAttributes(
		graphqlFieldName.String(fc.Field.Name),
		graphqlFieldPath.String(fc.Path().String()),
		graphqlFieldType.String(fc.Field.ObjectDefinition.Name),
	)
	if fc.Field.Alias != fc.Field.Name {
		span.SetAttributes(graphqlFieldAlias.String(fc.Field.Alias))
	}
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
	if oc.Operation == nil {
		return attribute.String("", "")
	}
	switch oc.Operation.Operation {
	case ast.Mutation:
		return semconv.GraphQLOperationTypeMutation
	case ast.Subscription:
		return semconv.GraphQLOperationTypeSubscription
	case ast.Query:
		return semconv.GraphQLOperationTypeQuery
	default:
		return semconv.GraphQLOperationTypeQuery
	}
}

var _ interface {
	graphql.HandlerExtension
	graphql.ResponseInterceptor
	graphql.FieldInterceptor
} = Tracer{}
