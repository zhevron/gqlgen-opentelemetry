# OpenTelemetry instrumentation for gqlgen

![Build Status](https://github.com/zhevron/gqlgen-opentelemetry/actions/workflows/go.yml/badge.svg?branch=main)
[![Go Reference](https://pkg.go.dev/badge/github.com/zhevron/gqlgen-opentelemetry.svg)](https://pkg.go.dev/github.com/zhevron/gqlgen-opentelemetry)

This library provides OpenTelemetry (OTEL) instrumentation for `gqlgen` server requests.

## Installation
Add the package to your project:
```sh
go get github.com/zhevron/gqlgen-opentelemetry
```

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

`TracerProvider`: The OTEL tracer provider to instantiate a tracer from. If none is provided, the global OTEL tracer provider will be used.
