package main

import (
	"graphql-go/auth"
	"graphql-go/graph"
	"graphql-go/persistence"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/go-chi/chi/v5"
	"github.com/gorilla/websocket"
)

const defaultPort = "8080"

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// If the request is an OPTIONS request, return immediately with a 200 response
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Continue processing the request
		next.ServeHTTP(w, r)
	})
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	gorm := persistence.ConnectGORM()

	// Create a resolver with the database connection
	resolver := graph.NewResolver(gorm)

	// Create your GraphQL server handler
	srv := handler.New(graph.NewExecutableSchema(graph.Config{Resolvers: resolver}))

	// Add WebSocket transport support for subscriptions
	srv.AddTransport(transport.Websocket{
		KeepAlivePingInterval: 10 * time.Second,
		Upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				// Allow all connections
				return true
			},
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		},
	})
	srv.AddTransport(transport.Options{})
	srv.AddTransport(transport.GET{})
	srv.AddTransport(transport.POST{})
	srv.Use(extension.Introspection{})

	// Create a new Chi router
	router := chi.NewRouter()

	// Use the Chi router methods to handle routes
	router.Use(corsMiddleware)

	gqlRouter := chi.NewRouter()
	gqlRouter.Use(auth.Middleware(gorm))
	gqlRouter.Handle("/", srv)

	authRouter := chi.NewRouter()
	authRouter.HandleFunc("/login", auth.HandleLogin)
	authRouter.HandleFunc("/auth/callback", auth.HandleCallback)
	// only use below on localhost
	if os.Getenv("DEBUG") == "true" {
		authRouter.HandleFunc("/login-dev", auth.HandleLoginDev)
	}

	router.Mount("/", authRouter)
	router.Mount("/query", gqlRouter)

	// Optionally, protect the GraphQL playground with the nonceMiddleware as well
	// This means accessing the playground will also require a valid nonce
	playgroundHandler := playground.Handler("GraphQL playground", "/query")
	router.Handle("/", playgroundHandler)

	log.Printf("connect to http://localhost:%s/ for GraphQL playground", port)
	log.Fatal(http.ListenAndServe(":"+port, router))
}
