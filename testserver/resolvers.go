package testserver

// THIS CODE WILL BE UPDATED WITH SCHEMA CHANGES. PREVIOUS IMPLEMENTATION FOR SCHEMA CHANGES WILL BE KEPT IN THE COMMENT SECTION. IMPLEMENTATION FOR UNCHANGED SCHEMA WILL BE KEPT.

import (
	"context"

	"github.com/zhevron/gqlgen-opentelemetry/v2/testserver/generated"
)

type Resolver struct{}

// Greet is the resolver for the greet field.
func (r *mutationResolver) Greet(ctx context.Context, name string) (string, error) {
	return "Hello " + name, nil
}

// Greeting is the resolver for the greeting field.
func (r *queryResolver) Greeting(ctx context.Context) (string, error) {
	return "Hello world", nil
}

// Mutation returns generated.MutationResolver implementation.
func (r *Resolver) Mutation() generated.MutationResolver { return &mutationResolver{r} }

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

type (
	mutationResolver struct{ *Resolver }
	queryResolver    struct{ *Resolver }
)
