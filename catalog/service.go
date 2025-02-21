package catalog

import (
	"context"

	"github.com/segmentio/ksuid"
)

type Service interface {
	PostProduct(ctx context.Context, name string, description string, price float64) (*Product, error)
	GetProduct(ctx context.Context, id string) (*Product, error)
	GetProducts(ctx context.Context, skip uint64, take uint64) ([]Product, error)
	GetProductsByIDs(ctx context.Context, ids []string) ([]Product, error)
	SearchProducts(ctx context.Context, query string, skip uint64, take uint64) ([]Product, error)
}

type Product struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
}

type catalogService struct {
	repository Repository
}

func NewService(r Repository) Service {
	return &catalogService{r}
}

func (r *catalogService) PostProduct(ctx context.Context, name string, description string, price float64) (*Product, error) {

	p := &Product{
		ID:          ksuid.New().String(),
		Name:        name,
		Description: description,
		Price:       price,
	}
	if err := r.repository.PutProduct(ctx, *p); err != nil {
		return nil, err
	}
	return p, nil
}

func (r *catalogService) GetProduct(ctx context.Context, id string) (*Product, error) {

	return r.repository.GetProductByID(ctx, id)

}

func (r *catalogService) GetProducts(ctx context.Context, skip uint64, take uint64) ([]Product, error) {
	if take > 100 || (skip == 0 && take == 0) {
		take = 100
	}

	return r.repository.ListProducts(ctx, skip, take)
}

func (r *catalogService) GetProductsByIDs(ctx context.Context, ids []string) ([]Product, error) {
	return r.repository.ListProductsWithIDs(ctx, ids)
}

func (r *catalogService) SearchProducts(ctx context.Context, query string, skip uint64, take uint64) ([]Product, error) {

	if take > 100 || (skip == 0 && take == 0) {
		take = 100
	}
	return r.repository.SearchProducts(ctx, query, skip, take)
}
