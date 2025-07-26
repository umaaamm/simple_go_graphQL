package main

import (
	"log"
	"os"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/gofiber/adaptor/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/vektah/gqlparser/v2/ast"

	"github.com/umaaamm/contact/graph"
	"github.com/umaaamm/contact/mongo"
)

const defaultPort = "8080"

func main() {
	err := mongo.Connect("mongodb://localhost:27017", "graphQL_test")
	if err != nil {
		log.Fatalf("MongoDB Connection Error: %v", err)
	}

	app := fiber.New()

	// 🌟 Setup gqlgen handler
	srv := handler.New(graph.NewExecutableSchema(graph.Config{
		Resolvers: &graph.Resolver{},
	}))

	// ➕ Add transports (WebSocket, POST, etc.)
	srv.AddTransport(transport.Options{})
	srv.AddTransport(transport.GET{})
	srv.AddTransport(transport.POST{})
	// Optionally: srv.AddTransport(transport.Websocket{})

	srv.SetQueryCache(lru.New[*ast.QueryDocument](1000))

	srv.Use(extension.Introspection{})
	srv.Use(extension.AutomaticPersistedQuery{
		Cache: lru.New[string](100),
	})

	app.Get("/", adaptor.HTTPHandlerFunc(playground.Handler("GraphQL playground", "/query")))
	app.All("/query", adaptor.HTTPHandler(srv))

	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}
	log.Printf("connect to http://localhost:%s/ for GraphQL playground", port)
	log.Fatal(app.Listen(":" + port))
}
