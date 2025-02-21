package main

import (
	"context"
	"log"
	"time"
)

type queryResolver struct {
	server *Server
}

func (r *queryResolver) Account(ctx context.Context, pagination *PaginationInput, id *string) ([]*Account, error) {

	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	//*Query for single account

	if id != nil {
		r, err := r.server.accountClient.GetAccount(ctx, *id)
		if err != nil {
			log.Println(err)
			return nil, err
		}

		return []*Account{
			{
				ID:   r.ID,
				Name: r.Name,
			},
		}, nil
	}

	skip, take := uint64(0), uint64(0)
	if pagination != nil {
		skip, take = pagination.bound()
	}

	accountList, err := r.server.accountClient.GetAccounts(ctx, skip, take)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	// type Account {
	// 	id : String!,
	// 	name : String!,
	// 	orders : [Order!]!
	// }

	var accounts []*Account
	for _, a := range accountList {

		account := &Account{
			ID:   a.ID,
			Name: a.Name,
		}
		accounts = append(accounts, account)

	}
	return accounts, nil

}

func (r *queryResolver) Products(ctx context.Context, pagination *PaginationInput, query *string, id *string) ([]*Product, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	if id != nil {

		r, err := r.server.catalogClient.GetProduct(ctx, *id)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		// 	ID          string  `json:"id"`
		// Name        string  `json:"name"`
		// Description string  `json:"description"`
		// Price       float64 `json:"price"`

		return []*Product{
			{
				ID:          r.ID,
				Name:        r.Name,
				Description: r.Description,
				Price:       r.Price,
			},
		}, nil
	}

	skip, take := uint64(0), uint64(0)
	if pagination != nil {
		skip, take = pagination.bound()
	}

	q := ""
	if query != nil {
		q = *query
	}

	productList, err := r.server.catalogClient.GetProducts(ctx, skip, take, nil, q)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	var products []*Product

	for _, p := range productList {
		product := &Product{
			ID:          p.ID,
			Name:        p.Name,
			Description: p.Description,
			Price:       p.Price,
		}
		products = append(products, product)

	}
	return products, nil
}

func (p PaginationInput) bound() (uint64, uint64) {
	skipValue := uint64(0)
	takeValue := uint64(100)

	if p.Skip != nil {
		skipValue = uint64(*p.Skip)
	}

	if p.Take != nil {
		takeValue = uint64(*p.Take)
	}
	return skipValue, takeValue
}
