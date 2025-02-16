package main

import (
	"flag"
	"fmt"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/KaffeeMaschina/ozon_test_task/config"
	graph2 "github.com/KaffeeMaschina/ozon_test_task/internals/graph"
	"github.com/KaffeeMaschina/ozon_test_task/internals/storage"
	"github.com/vektah/gqlparser/v2/ast"
	"log/slog"
	"net/http"
	"os"
)

const (
	port = "8080"
)

func main() {
	// TODO: Change config for better realisation
	log := NewLogger()

	cfg := config.MustLoad()
	log.Info("config is loaded")

	var store storage.Storage
	var err error
	var usePostgres bool

	flag.BoolVar(&usePostgres, "usePostgres", false, "use postgres instead of in memory cache")
	flag.Parse()

	if usePostgres {

		store, err = storage.NewPostgresStorage(cfg.Username, cfg.Password,
			cfg.DBPort, cfg.Database)
		if err != nil {
			log.Error(err.Error())
			os.Exit(1)
		}
		log.Info("using postgres")

	} else {
		store = storage.NewCache()
		log.Info("using cache")
	}

	srv := handler.New(graph2.NewExecutableSchema(graph2.Config{Resolvers: &graph2.Resolver{store, log}}))

	srv.AddTransport(transport.Options{})
	srv.AddTransport(transport.GET{})
	srv.AddTransport(transport.POST{})

	srv.SetQueryCache(lru.New[*ast.QueryDocument](1000))

	srv.Use(extension.Introspection{})
	srv.Use(extension.AutomaticPersistedQuery{
		Cache: lru.New[string](100),
	})

	http.Handle("/query", srv)

	log.Info(fmt.Sprintf("connected to http://localhost:%s/", port))
	log.Error(http.ListenAndServe(":"+port, nil).Error())
}

func NewLogger() *slog.Logger {
	opts := &slog.HandlerOptions{Level: slog.LevelDebug}
	logger := slog.New(slog.NewTextHandler(os.Stdout, opts))
	return logger
}
