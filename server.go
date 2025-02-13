package main

import (
	"flag"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/KaffeeMaschina/ozon_test_task/graph"
	config "github.com/KaffeeMaschina/ozon_test_task/internals"
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

	srv := handler.New(graph.NewExecutableSchema(graph.Config{Resolvers: &graph.Resolver{store}}))

	srv.AddTransport(transport.Options{})
	srv.AddTransport(transport.GET{})
	srv.AddTransport(transport.POST{})

	srv.SetQueryCache(lru.New[*ast.QueryDocument](1000))

	srv.Use(extension.Introspection{})
	srv.Use(extension.AutomaticPersistedQuery{
		Cache: lru.New[string](100),
	})

	http.Handle("/", playground.Handler("GraphQL playground", "/query"))
	http.Handle("/query", srv)

	log.Printf("connect to http://localhost:%s/ for GraphQL playground", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
