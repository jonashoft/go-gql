package main

import (
	"graphql-go/graph"
	"graphql-go/persistence"
	"log"
	"net/http"
	"os"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
)

const defaultPort = "8080"

// Middleware to enforce nonce check
func nonceMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Host == "localhost:"+defaultPort {
			next.ServeHTTP(w, r)
			return
		}
		// Check for nonce in query parameters
		nonce := r.Header.Get("X-Nonce")
		// For demonstration, let's assume the valid nonce is "12345"
		validNonce := os.Getenv("VALID_NONCE")

		// Validate the nonce
		if nonce != validNonce {
			// Respond with an error if the nonce is invalid
			http.Error(w, "Invalid or missing nonce", http.StatusUnauthorized)
			return
		}

		// If the nonce is valid, proceed with the request
		next.ServeHTTP(w, r)
	})
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-Nonce")

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

	// Create your GraphQL server handler
	srv := handler.NewDefaultServer(graph.NewExecutableSchema(graph.Config{Resolvers: &graph.Resolver{
		DB: persistence.ConnectGORM(),
	}}))

	// Wrap the /query handler with nonceMiddleware to protect your GraphQL API
	http.Handle("/query", corsMiddleware(nonceMiddleware(srv)))

	// Optionally, protect the GraphQL playground with the nonceMiddleware as well
	// This means accessing the playground will also require a valid nonce
	playgroundHandler := playground.Handler("GraphQL playground", "/query")
	http.Handle("/", playgroundHandler)

	log.Printf("connect to http://localhost:%s/ for GraphQL playground", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
