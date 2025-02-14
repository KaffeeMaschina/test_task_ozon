This project is my first experience with GraphQL.

I use a [gqlgen](https://github.com/99designs/gqlgen) library. That generates models, resolvers and inner logic for GraphQL server.

There are two types of storing data: in-memory cache by default, or PostgreSQL using flag -usePostgres.

In-memory cache is implemented with maps and RWMutex(for concurrent access).

In Makefile you can find a dependency [goose](https://github.com/pressly/goose) and commands for migrations.

Set `CONFIG_PATH` env variable to `./config/local.yaml` before use.

Text GraphQL queries with http requests in /docs/template
