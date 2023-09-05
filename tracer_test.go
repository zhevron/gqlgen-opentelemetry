package gqlgen_opentelemetry

import (
	"testing"

	"github.com/99designs/gqlgen/client"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/stretchr/testify/suite"
	"github.com/zhevron/gqlgen-opentelemetry/testserver"
	"github.com/zhevron/gqlgen-opentelemetry/testserver/generated"
	"go.opentelemetry.io/otel/attribute"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	semconv "go.opentelemetry.io/otel/semconv/v1.20.0"
)

type TracerSuite struct {
	suite.Suite
	Exporter       *tracetest.InMemoryExporter
	TracerProvider *sdktrace.TracerProvider
}

func (s *TracerSuite) SetupSuite() {
	s.Exporter = tracetest.NewInMemoryExporter()
	s.TracerProvider = sdktrace.NewTracerProvider(
		sdktrace.WithSyncer(s.Exporter),
	)
}

func (s *TracerSuite) TearDownTest() {
	s.Exporter.Reset()
}

func (s *TracerSuite) TestQuery_SpanCreated() {
	c := s.createTestClient(&Tracer{})

	query := "query GetGreeting { greeting }"
	var res struct{ Greeting string }
	c.MustPost(query, &res)

	spans := s.Exporter.GetSpans()
	s.Require().Len(spans, 1)
	s.Require().Len(spans[0].Attributes, 6)

	operationName := findAttributeByName(spans[0].Attributes, semconv.GraphqlOperationNameKey)
	s.Require().NotNil(operationName)
	s.Require().Equal(operationName.Value.AsString(), "GetGreeting")

	operationType := findAttributeByName(spans[0].Attributes, semconv.GraphqlOperationTypeKey)
	s.Require().NotNil(operationType)
	s.Require().Equal(*operationType, semconv.GraphqlOperationTypeQuery)

	document := findAttributeByName(spans[0].Attributes, semconv.GraphqlDocumentKey)
	s.Require().NotNil(document)
	s.Require().Equal(document.Value.AsString(), query)
}

func (s *TracerSuite) TestQuery_WithoutFieldSpans() {
	c := s.createTestClient(&Tracer{
		IncludeFieldSpans: false,
	})

	var res struct{ Greeting string }
	c.MustPost("query { greeting }", &res)

	spans := s.Exporter.GetSpans()
	s.Require().Len(spans, 1)
}

func (s *TracerSuite) TestQuery_WithFieldSpans() {
	c := s.createTestClient(&Tracer{
		IncludeFieldSpans: true,
	})

	var res struct{ Greeting string }
	c.MustPost("query { greeting }", &res)

	spans := s.Exporter.GetSpans()
	s.Require().Len(spans, 2)
	span := findSpanByName(spans, "Query.greeting")
	s.Require().NotNil(span)
	s.Require().Len(span.Attributes, 5)

	fieldName := findAttributeByName(span.Attributes, graphqlFieldName)
	s.Require().NotNil(fieldName)
	s.Require().Equal(fieldName.Value.AsString(), "greeting")
}

func (s *TracerSuite) TestQuery_WithFieldSpans_Alias() {
	c := s.createTestClient(&Tracer{
		IncludeFieldSpans: true,
	})

	var res struct{ MyGreeting string }
	c.MustPost("query { myGreeting: greeting }", &res)

	spans := s.Exporter.GetSpans()
	s.Require().Len(spans, 2)
	span := findSpanByName(spans, "Query.greeting")
	s.Require().NotNil(span)

	fieldName := findAttributeByName(span.Attributes, graphqlFieldName)
	s.Require().NotNil(fieldName)
	s.Require().Equal(fieldName.Value.AsString(), "greeting")

	fieldAlias := findAttributeByName(span.Attributes, graphqlFieldAlias)
	s.Require().NotNil(fieldAlias)
	s.Require().Equal(fieldAlias.Value.AsString(), "myGreeting")
}

func (s *TracerSuite) TestMutation_SpanCreated() {
	c := s.createTestClient(&Tracer{})

	query := "mutation Greet($name: String!) { greet(name: $name) }"
	var res struct{ Greet string }
	c.MustPost(query, &res, client.Var("name", "gqlgen"))

	spans := s.Exporter.GetSpans()
	s.Require().Len(spans, 1)
	s.Require().Len(spans[0].Attributes, 6)

	operationName := findAttributeByName(spans[0].Attributes, semconv.GraphqlOperationNameKey)
	s.Require().NotNil(operationName)
	s.Require().Equal(operationName.Value.AsString(), "Greet")

	operationType := findAttributeByName(spans[0].Attributes, semconv.GraphqlOperationTypeKey)
	s.Require().NotNil(operationType)
	s.Require().Equal(*operationType, semconv.GraphqlOperationTypeMutation)

	document := findAttributeByName(spans[0].Attributes, semconv.GraphqlDocumentKey)
	s.Require().NotNil(document)
	s.Require().Equal(document.Value.AsString(), query)
}

func (s *TracerSuite) TestMutation_WithoutVariables() {
	c := s.createTestClient(&Tracer{
		IncludeVariables: false,
	})

	var res struct{ Greet string }
	c.MustPost("mutation Greet($name: String!) { greet(name: $name) }", &res, client.Var("name", "gqlgen"))

	spans := s.Exporter.GetSpans()
	s.Require().Len(spans, 1)

	nameVariable := findAttributeByName(spans[0].Attributes, graphqlVariablesPrefix+"name")
	s.Require().Nil(nameVariable)
}

func (s *TracerSuite) TestMutation_WithVariables() {
	c := s.createTestClient(&Tracer{
		IncludeVariables: true,
	})

	var res struct{ Greet string }
	c.MustPost("mutation Greet($name: String!) { greet(name: $name) }", &res, client.Var("name", "gqlgen"))

	spans := s.Exporter.GetSpans()
	s.Require().Len(spans, 1)

	nameVariable := findAttributeByName(spans[0].Attributes, graphqlVariablesPrefix+"name")
	s.Require().NotNil(nameVariable)
	s.Require().Equal(nameVariable.Value.AsString(), "gqlgen")
}

func (s *TracerSuite) createTestClient(tracer *Tracer) *client.Client {
	tracer.TracerProvider = s.TracerProvider
	handler := handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{
		Resolvers: &testserver.Resolver{},
	}))
	handler.Use(tracer)
	handler.Use(extension.FixedComplexityLimit(100))
	return client.New(handler)
}

func TestTracerSuite(t *testing.T) {
	suite.Run(t, new(TracerSuite))
}

func findAttributeByName(attributes []attribute.KeyValue, name attribute.Key) *attribute.KeyValue {
	for _, a := range attributes {
		if a.Key == name {
			return &a
		}
	}
	return nil
}

func findSpanByName(spans tracetest.SpanStubs, name string) *tracetest.SpanStub {
	for _, span := range spans {
		if span.Name == name {
			return &span
		}
	}
	return nil
}
