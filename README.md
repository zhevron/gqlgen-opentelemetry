# OpenTelemetry instrumentation for gqlgen

This library provides OpenTelemetry (OTEL) instrumentation for `gqlgen` server requests.

## Usage
Add the `gqlgen_opentelemetry.Tracer` extension to your server:
```go
h := handler.NewDefaultServer(schema)
h.Use(gqlgen_opentelemetry.Tracer{})
```

## Options
The following options are available on the extension:

`IncludeFieldSpans`: Whether to create an additional child span for each field requested. (Default: `false`)

`IncludeVariables`: Whether to include variables and their values in the trace span attributes. (Default: `false`)

`Tracer`: The OTEL tracer to use. If none is provided, the global OTEL tracer provider will be used.
