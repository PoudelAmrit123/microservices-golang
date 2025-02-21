package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/kelseyhightower/envconfig"
)

type AppConfig struct {
	AccountURL string `envconfig:"ACCOUNT_SERIVCE_URL"`
	CatalogURL string `envconfig:"CATALOG_SERIVCE_URL"`
	OrderURL   string `envconfig:"ORDER_SERVICE_URL"`
}

func main() {
	var cfg AppConfig

	err := envconfig.Process("", &cfg)

	if err != nil {
		log.Fatal(err.Error())
	}
	fmt.Println(cfg.AccountURL)

	s, err := NewGraphQlServer(cfg.AccountURL, cfg.CatalogURL, cfg.OrderURL)
	if err != nil {
		log.Fatal(err.Error())
	}

	http.Handle("/graphql", handler.NewDefaultServer(s.ToExecutableSchema()))
	http.Handle("/playground", playground.Handler("amrit", "/graphql"))

	log.Fatal(http.ListenAndServe(":8080", nil))
}
