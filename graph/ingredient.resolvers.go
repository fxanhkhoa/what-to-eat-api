package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.
// Code generated by github.com/99designs/gqlgen version v0.17.40

import (
	"context"
	"what-to-eat/be/auth"
	"what-to-eat/be/graph/model"
	"what-to-eat/be/graph/service"
)

// CreateIngredient is the resolver for the createIngredient field.
func (r *mutationResolver) CreateIngredient(ctx context.Context, createIngredientInput model.CreateIngredientInput) (*model.Ingredient, error) {
	user := auth.ForContext(ctx)
	ingredient, err := service.NewIngredientService().Create(createIngredientInput, user)
	return ingredient, err
}

// UpdateIngredient is the resolver for the updateIngredient field.
func (r *mutationResolver) UpdateIngredient(ctx context.Context, updateIngredientInput model.UpdateIngredientInput) (*model.Ingredient, error) {
	user := auth.ForContext(ctx)
	ingredient, err := service.NewIngredientService().Update(updateIngredientInput, user)
	return ingredient, err
}

// RemoveIngredient is the resolver for the removeIngredient field.
func (r *mutationResolver) RemoveIngredient(ctx context.Context, slug string) (*model.Ingredient, error) {
	user := auth.ForContext(ctx)
	ingredient, err := service.NewIngredientService().Remove(slug, user)
	return ingredient, err
}

// Ingredients is the resolver for the ingredients field.
func (r *queryResolver) Ingredients(ctx context.Context, keyword *string, page *int, limit *int) ([]*model.Ingredient, error) {
	ingredients, err := service.NewIngredientService().Find(keyword, page, limit)
	return ingredients, err
}

// Ingredient is the resolver for the ingredient field.
func (r *queryResolver) Ingredient(ctx context.Context, slug string) (*model.Ingredient, error) {
	ingredient, err := service.NewIngredientService().FindOne(slug)
	return ingredient, err
}
