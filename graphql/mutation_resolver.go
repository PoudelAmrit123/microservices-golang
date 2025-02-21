package main

import (
	"context"
	"errors"
	"time"

	"github.com/PoudelAmrit123/microservice/order"
)

type mutationResolver struct {
	server *Server
}

var (
	ErrInvalidParameter = errors.New("Invalid parameter")
)

// CreateAccount(ctx context.Context, account *AccountInput) (*Account, error)
//
//	CreateProduct(ctx context.Context, product *ProductInput) (*Product, error)
//	CreateOrder(ctx context.Context, order *OrderInput) (*Order, error)
func (r *mutationResolver) CreateAccount(ctx context.Context, in *AccountInput) (*Account, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	a, err := r.server.accountClient.PostAccount(ctx, in.Name)
	if err != nil {
		return nil, err
	}

	return &Account{
		ID:   a.ID,
		Name: a.Name,
	}, nil
}
func (r *mutationResolver) CreateProduct(ctx context.Context, in *ProductInput) (*Product, error) {

	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	cp, err := r.server.catalogClient.PostProduct(ctx, in.Name, in.Description, in.Price)
	if err != nil {
		return nil, err
	}

	return &Product{
		ID:          cp.ID,
		Description: cp.Description,
		Name:        cp.Name,
		Price:       cp.Price,
	}, nil

}

func (r *mutationResolver) CreateOrder(ctx context.Context, in *OrderInput) (*Order, error) {

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var products []order.OrderedProduct

	for _, p := range in.Prodcuts {
		if p.Quantity <= 0 {
			return nil, ErrInvalidParameter
		}
		products = append(products, order.OrderedProduct{
			ID:       p.ID,
			Quantity: uint32(p.Quantity),
		})

	}

	o, err := r.server.orderClient.PostOrder(ctx, in.AccountID, products)
	if err != nil {
		return nil, err
	}
	return &Order{
		ID:         o.ID,
		CreatedAt:  o.CreatedAt,
		TotalPrice: o.TotalPrice,
	}, nil

}
