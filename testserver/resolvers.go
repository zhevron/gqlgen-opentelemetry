//go:generate rm -rf generated
//go:generate go run -mod=mod github.com/99designs/gqlgen generate
package testserver

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

type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
