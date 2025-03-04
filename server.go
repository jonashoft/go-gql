package main

import (
	"graphql-go/auth"
	"graphql-go/graph"
	"graphql-go/persistence"
	"log"
	"net/http"
	"os"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/go-chi/chi/v5"
)

const defaultPort = "8080"

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
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

	// Create your GraphQL server handler
	srv := handler.NewDefaultServer(graph.NewExecutableSchema(graph.Config{Resolvers: &graph.Resolver{
		DB: gorm,
	}}))

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
