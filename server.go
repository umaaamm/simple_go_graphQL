package main

import (
	"bytes"
	"context"
	"log"
	"net/http"
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
	"github.com/umaaamm/contact/internal/auth"
	"github.com/umaaamm/contact/mongo"
)

const defaultPort = "8080"

func main() {
	err := mongo.Connect("mongodb://localhost:27017", "graphQL_test")
	if err != nil {
		log.Fatalf("MongoDB Connection Error: %v", err)
	}

	app := fiber.New()

	app.Use(auth.Middleware())

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

	app.All("/query", func(c *fiber.Ctx) error {
		req, err := convertFiberToHTTPRequest(c)
		if err != nil {
			return err
		}
		ctx := c.UserContext()
		if ctx == nil {
			ctx = context.Background()
		}
		req = req.WithContext(ctx)

		rw := &fiberResponseWriter{c}

		srv.ServeHTTP(rw, req)
		return nil
	})

	app.Get("/", adaptor.HTTPHandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		playground.Handler("GraphQL Playground", "/query").ServeHTTP(w, r)
	}))

	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}
	log.Printf("connect to http://localhost:%s/ for GraphQL playground", port)
	log.Fatal(app.Listen(":" + port))
}

// fiber → *http.Request
func convertFiberToHTTPRequest(c *fiber.Ctx) (*http.Request, error) {
	req, err := http.NewRequest(
		string(c.Method()),
		c.OriginalURL(),
		bytes.NewReader(c.Body()),
	)
	if err != nil {
		return nil, err
	}
	c.Request().Header.VisitAll(func(k, v []byte) {
		req.Header.Set(string(k), string(v))
	})
	req.RemoteAddr = c.IP()
	return req, nil
}

// *http.ResponseWriter → fiber
type fiberResponseWriter struct {
	c *fiber.Ctx
}

func (w *fiberResponseWriter) Header() http.Header {
	h := http.Header{}
	w.c.Response().Header.VisitAll(func(key, val []byte) {
		h.Set(string(key), string(val))
	})
	return h
}

func (w *fiberResponseWriter) Write(b []byte) (int, error) {
	return w.c.Write(b)
}

func (w *fiberResponseWriter) WriteHeader(statusCode int) {
	w.c.Status(statusCode)
}
