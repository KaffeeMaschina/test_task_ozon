package main

import (
	"flag"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/KaffeeMaschina/ozon_test_task/config"
	graph2 "github.com/KaffeeMaschina/ozon_test_task/internals/graph"
	"github.com/KaffeeMaschina/ozon_test_task/internals/storage"
	"github.com/vektah/gqlparser/v2/ast"
	"log"
	"net/http"
)

const (
	port = "8080"
)

func main() {
	cfg := config.MustLoad()
	log.Println("config is loaded")

	var store storage.Storage
	var err error
	var usePostgres bool

	flag.BoolVar(&usePostgres, "usePostgres", false, "use postgres instead of in memory cache")
	flag.Parse()

	if usePostgres {
		store, err = storage.NewPostgresStorage(cfg.Username, cfg.Password,
			cfg.DBPort, cfg.Database)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("using postgres")
	} else {
		store = storage.NewCache()
		log.Println("using cache")
	}

	srv := handler.New(graph2.NewExecutableSchema(graph2.Config{Resolvers: &graph2.Resolver{store}}))

	srv.AddTransport(transport.Options{})
	srv.AddTransport(transport.GET{})
	srv.AddTransport(transport.POST{})

	srv.SetQueryCache(lru.New[*ast.QueryDocument](1000))

	srv.Use(extension.Introspection{})
	srv.Use(extension.AutomaticPersistedQuery{
		Cache: lru.New[string](100),
	})

	http.Handle("/query", srv)

	log.Printf("connected to http://localhost:%s/", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
