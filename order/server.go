package order

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"

	"github.com/PoudelAmrit123/microservice/account"
	"github.com/PoudelAmrit123/microservice/catalog"
	"github.com/PoudelAmrit123/microservice/order/pb/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type grpcServer struct {
	pb.UnimplementedOrderServiceServer
	service       Service
	accountClinet *account.Client
	catalog       *catalog.Client
}

func ListenGRPC(s Service, accountURL string, catalogURL string, port int) error {

	accountClinet, err := account.NewClient(accountURL)
	if err != nil {
		return err
	}

	catalogClient, err := catalog.NewClient(catalogURL)
	if err != nil {
		return err
	}
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		accountClinet.Close()
		catalogClient.Close()
		return err
	}

	serv := grpc.NewServer()
	pb.RegisterOrderServiceServer(serv, &grpcServer{
		pb.UnimplementedOrderServiceServer{},
		s,
		accountClinet,
		catalogClient,
	})

	reflection.Register(serv)

	return serv.Serve(lis)

}

func (s *grpcServer) PostOrder(ctx context.Context, r *pb.PostOrderRequest) (*pb.PostOrderResponse, error) {

	//*Checking if the account exists or not

	_, err := s.accountClinet.GetAccount(ctx, r.AccountId)
	if err != nil {
		log.Println("Error getting account: ", err)
		return nil, errors.New("account not found")
	}

	//*Getting products that are orderd

	//*here while creating the order we are receiving the only accountId and productOrder as input as defined in graphQl server so we have to first fetch the that particular product description price and other informatin from our catalog and then create a new order type that consiste of all the information required for that field . (if we have defined to provide all the informaton while mutation the order we donot have to do this extra work right )

	productIDs := []string{}

	for _, p := range r.Products {
		productIDs = append(productIDs, p.ProductId)

	}

	orderedProducts, err := s.catalog.GetProducts(ctx, 0, 0, productIDs, "")
	if err != nil {
		log.Println("Error getting products: ", err)
		return nil, errors.New("products not found")
	}

	//*making the products

	products := []OrderedProduct{}
	for _, p := range orderedProducts {
		product := OrderedProduct{
			ID:          p.ID,
			Quantity:    0,
			Price:       p.Price,
			Name:        p.Name,
			Description: p.Description,
		}
		//* the rp is the product id that is provided by us while creating the order and p is the product id details we obtain from the finding hte product detaisl
		for _, rp := range r.Products {
			if rp.ProductId == p.ID {
				product.Quantity = rp.Quantity
				break
			}

		}

		if product.Quantity != 0 {
			products = append(products, product)
		}
	}

	order, err := s.service.PostOrder(ctx, r.AccountId, products)
	if err != nil {
		log.Println("Error posting order: ", err)
		return nil, errors.New("could not post order")
	}

	//*	we get this
	//*	type Order struct {
	//*		ID         string
	//*		CreatedAt  time.Time
	//*		TotalPrice float64
	//*		AccountID  string
	//*		Products   []OrderedProduct
	//*	}

	//*	type OrderedProduct struct {
	//*		ID          string
	//*		Name        string
	//*		Description string
	//*		Price       float64
	//*		Quantity    uint32
	//*	}

	//*we have to send back this

	//*	type Product {
	//*		id : String!,
	//*		name : String!,
	//*		description : String!,
	//*		price : Float!
	//*	}

	//*	type Order {
	//*		id : String!,
	///*/*		createdAt : Time!,
	//*		products : [Product!]!,
	//*		totalPrice : Float!
	//*	}

	orderProto := &pb.Order{
		Id:         order.ID,
		AccountId:  order.AccountID,
		TotalPrice: order.TotalPrice,
		Products:   []*pb.Order_OrderProduct{},
	}
	orderProto.CreatedAt, _ = order.CreatedAt.MarshalBinary()
	for _, p := range order.Products {

		orderProto.Products = append(orderProto.Products, &pb.Order_OrderProduct{
			Id:          p.ID,
			Name:        p.Name,
			Description: p.Description,
			Price:       p.Price,
			Quantity:    p.Quantity,
		})

	}

	return &pb.PostOrderResponse{
		Order: orderProto,
	}, nil

}

func (s *grpcServer) GetOrdersForAccount(ctx context.Context, r *pb.GetOrdersForAccountRequest) (*pb.GetOrdersForAccountResponse, error) {
	//*We get this
	// type Order struct {
	// 	ID         string
	// 	CreatedAt  time.Time
	// 	TotalPrice float64
	// 	AccountID  string
	// 	Products   []OrderedProduct
	// }

	//*we have to return this
	// message Order {
	// 	message OrderProduct {
	// 		string id = 1;
	// 		string name = 2;
	// 		string description = 3;
	// 		double price = 4;
	// 		uint32 quantity = 5;
	// 	}

	// 	string id = 1;
	// 	bytes createdAt = 2;
	// 	string accountId = 3;
	// 	double totalPrice = 4;
	// 	repeated OrderProduct products = 5;
	// }
	accountOrders, err := s.service.GetOrdersForAccount(ctx, r.AccountId)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	//* Getting all ordered products

	productIDMap := map[string]bool{}

	for _, o := range accountOrders {
		for _, p := range o.Products {

			productIDMap[p.ID] = true

		}

	}

	productIDs := []string{}

	for id := range productIDMap {
		productIDs = append(productIDs, id)
	}

	products, err := s.catalog.GetProducts(ctx, 0, 0, productIDs, "")
	if err != nil {
		log.Println("Error getting account products: ", err)
		return nil, err
	}
	orders := []*pb.Order{}

	for _, o := range accountOrders {

		op := &pb.Order{
			Id:         o.ID,
			AccountId:  o.AccountID,
			TotalPrice: o.TotalPrice,
			Products:   []*pb.Order_OrderProduct{},
		}
		op.CreatedAt, _ = o.CreatedAt.MarshalBinary()

		//*adding the products description on the product

		for _, product := range o.Products {
			for _, p := range products {

				if p.ID == product.ID {
					product.Name = p.Name
					product.Description = p.Description
					product.Price = p.Price
					break
				}

			}

			op.Products = append(op.Products, &pb.Order_OrderProduct{
				Id:          product.ID,
				Name:        product.Name,
				Description: product.Description,
				Price:       product.Price,
				Quantity:    product.Quantity,
			})
		}
		orders = append(orders, op)

	}

	return &pb.GetOrdersForAccountResponse{Orders: orders}, nil

	// message Order {
	// 	message OrderProduct {
	// 		string id = 1;
	// 		string name = 2;
	// 		string description = 3;
	// 		double price = 4;
	// 		uint32 quantity = 5;
	// 	}

	// 	string id = 1;
	// 	bytes createdAt = 2;
	// 	string accountId = 3;
	// 	double totalPrice = 4;
	// 	repeated OrderProduct products = 5;
	// }
	//*Creating our order as per proto file

}
